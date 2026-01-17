import Foundation
import KeychainAccess

/// Manages authentication state and token storage
@Observable
final class AuthService: @unchecked Sendable {
    static let shared = AuthService()

    private(set) var currentUser: User?
    private(set) var identities: [Identity] = []
    private(set) var passkeys: [Passkey] = []
    private(set) var isAuthenticated = false
    private(set) var isLoading = false
    private(set) var error: Error?

    private let keychain = Keychain(service: "app.saga")
    private let apiClient = APIClient.shared

    private enum KeychainKey {
        static let accessToken = "access_token"
        static let refreshToken = "refresh_token"
    }

    init() {
        Task {
            await loadStoredTokens()
        }
    }

    // MARK: - Token Storage

    private func loadStoredTokens() async {
        guard let accessToken = keychain[KeychainKey.accessToken],
              let refreshToken = keychain[KeychainKey.refreshToken] else {
            return
        }

        await apiClient.setTokens(access: accessToken, refresh: refreshToken)

        // Verify tokens are still valid
        do {
            let response = try await apiClient.getCurrentUser()
            await MainActor.run {
                self.currentUser = response.data.user
                self.identities = response.data.identities
                self.passkeys = response.data.passkeys ?? []
                self.isAuthenticated = true
            }
        } catch {
            // Tokens are invalid, clear them
            await clearTokens()
        }
    }

    private func saveTokens(_ tokens: TokenResponse) async {
        keychain[KeychainKey.accessToken] = tokens.accessToken
        keychain[KeychainKey.refreshToken] = tokens.refreshToken
        await apiClient.setTokens(access: tokens.accessToken, refresh: tokens.refreshToken)
    }

    private func clearTokens() async {
        try? keychain.remove(KeychainKey.accessToken)
        try? keychain.remove(KeychainKey.refreshToken)
        await apiClient.clearTokens()
    }

    // MARK: - Email/Password Auth

    /// Register a new user with email and password
    func register(email: String, password: String, firstname: String? = nil, lastname: String? = nil) async throws {
        await setLoading(true)
        defer { Task { await setLoading(false) } }

        let request = RegisterRequest(email: email, password: password, firstname: firstname, lastname: lastname)
        let response = try await apiClient.register(request)
        await saveTokens(response.data.token)

        await MainActor.run {
            self.currentUser = response.data.user
            self.isAuthenticated = true
            self.error = nil
        }

        // Fetch full user details including identities
        try? await refreshCurrentUser()
    }

    /// Login with email and password
    func login(email: String, password: String) async throws {
        await setLoading(true)
        defer { Task { await setLoading(false) } }

        let request = LoginRequest(email: email, password: password)
        let response = try await apiClient.login(request)
        await saveTokens(response.data.token)

        await MainActor.run {
            self.currentUser = response.data.user
            self.isAuthenticated = true
            self.error = nil
        }

        // Fetch full user details including identities
        try? await refreshCurrentUser()
    }

    /// Logout and clear all auth state
    func logout() async throws {
        await setLoading(true)
        defer { Task { await setLoading(false) } }

        try await apiClient.logout()
        await clearTokens()

        await MainActor.run {
            self.currentUser = nil
            self.identities = []
            self.passkeys = []
            self.isAuthenticated = false
        }
    }

    // MARK: - OAuth

    /// Login or register with Google OAuth
    func loginWithGoogle(code: String, codeVerifier: String) async throws {
        await setLoading(true)
        defer { Task { await setLoading(false) } }

        let request = OAuthRequest(code: code, codeVerifier: codeVerifier)
        let response = try await apiClient.loginWithGoogle(request)
        await saveTokens(response.data.token)

        await MainActor.run {
            self.currentUser = response.data.user
            self.isAuthenticated = true
            self.error = nil
        }

        try? await refreshCurrentUser()
    }

    /// Login or register with Apple OAuth
    func loginWithApple(code: String, codeVerifier: String) async throws {
        await setLoading(true)
        defer { Task { await setLoading(false) } }

        let request = OAuthRequest(code: code, codeVerifier: codeVerifier)
        let response = try await apiClient.loginWithApple(request)
        await saveTokens(response.data.token)

        await MainActor.run {
            self.currentUser = response.data.user
            self.isAuthenticated = true
            self.error = nil
        }

        try? await refreshCurrentUser()
    }

    // MARK: - Passkey Auth

    /// Complete passkey login after WebAuthn assertion
    func loginWithPasskey(credential: PasskeyCredential) async throws {
        await setLoading(true)
        defer { Task { await setLoading(false) } }

        let response = try await apiClient.completePasskeyLogin(credential: credential)
        await saveTokens(response.data.token)

        await MainActor.run {
            self.currentUser = response.data.user
            self.isAuthenticated = true
            self.error = nil
        }

        try? await refreshCurrentUser()
    }

    /// Register a new passkey for the current user
    func registerPasskey(credential: PasskeyCredential) async throws -> Passkey {
        await setLoading(true)
        defer { Task { await setLoading(false) } }

        let response = try await apiClient.completePasskeyRegistration(credential: credential)

        await MainActor.run {
            self.passkeys.append(response.data.passkey)
        }

        return response.data.passkey
    }

    /// Delete a passkey
    func deletePasskey(id: String) async throws {
        try await apiClient.deletePasskey(id: id)

        await MainActor.run {
            self.passkeys.removeAll { $0.id == id }
        }
    }

    // MARK: - Account Linking

    /// Link an OAuth provider to the current account
    func linkOAuthIdentity(provider: String, code: String, codeVerifier: String) async throws {
        await setLoading(true)
        defer { Task { await setLoading(false) } }

        let response = try await apiClient.linkOAuthIdentity(provider: provider, code: code, codeVerifier: codeVerifier)

        await MainActor.run {
            self.identities.append(response.data)
        }
    }

    /// Unlink an identity from the account
    func unlinkIdentity(id: String) async throws {
        await setLoading(true)
        defer { Task { await setLoading(false) } }

        try await apiClient.unlinkIdentity(id: id)

        await MainActor.run {
            self.identities.removeAll { $0.id == id }
        }
    }

    // MARK: - User Data

    /// Refresh current user data
    func refreshCurrentUser() async throws {
        let response = try await apiClient.getCurrentUser()

        await MainActor.run {
            self.currentUser = response.data.user
            self.identities = response.data.identities
            self.passkeys = response.data.passkeys ?? []
        }
    }

    // MARK: - Helpers

    private func setLoading(_ loading: Bool) async {
        await MainActor.run {
            self.isLoading = loading
        }
    }

    /// Check if user has a specific identity provider linked
    func hasLinkedProvider(_ provider: IdentityProvider) -> Bool {
        identities.contains { $0.provider == provider }
    }

    /// Get identity for a specific provider
    func identity(for provider: IdentityProvider) -> Identity? {
        identities.first { $0.provider == provider }
    }

    // MARK: - Demo/Testing Support

    #if DEBUG
    /// Demo user credentials (matches API seed data)
    enum DemoCredentials {
        static let email = "demo@forgo.software"
        static let password = "password123"
    }

    /// Login with demo user - uses seed data from API
    func loginWithDemoUser() async throws {
        try await login(email: DemoCredentials.email, password: DemoCredentials.password)
    }

    /// Bypass auth for UI testing by setting state directly
    /// - Parameters:
    ///   - user: The mock user to set
    ///   - accessToken: Access token for API calls
    ///   - refreshToken: Refresh token for token refresh
    func setAuthStateForTesting(user: User, accessToken: String, refreshToken: String) async {
        keychain[KeychainKey.accessToken] = accessToken
        keychain[KeychainKey.refreshToken] = refreshToken
        await apiClient.setTokens(access: accessToken, refresh: refreshToken)

        await MainActor.run {
            self.currentUser = user
            self.isAuthenticated = true
            self.error = nil
        }
    }

    /// Clear all auth state for testing
    func clearAuthStateForTesting() async {
        await clearTokens()
        await MainActor.run {
            self.currentUser = nil
            self.identities = []
            self.passkeys = []
            self.isAuthenticated = false
        }
    }
    #endif
}
