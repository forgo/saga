import Foundation
import AuthenticationServices
import KeychainAccess

/// Manages authentication state and token storage
@Observable
final class AuthService: @unchecked Sendable {
    static let shared = AuthService()

    private(set) var currentUser: User?
    private(set) var isAuthenticated = false
    private(set) var isLoading = false

    private let keychain = Keychain(service: "app.babysync")
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

    func register(email: String, password: String, firstname: String? = nil, lastname: String? = nil) async throws {
        await MainActor.run { isLoading = true }
        defer { Task { @MainActor in isLoading = false } }

        let request = RegisterRequest(email: email, password: password, firstname: firstname, lastname: lastname)
        let response = try await apiClient.register(request)
        await saveTokens(response.data.token)

        await MainActor.run {
            self.currentUser = response.data.user
            self.isAuthenticated = true
        }
    }

    func login(email: String, password: String) async throws {
        await MainActor.run { isLoading = true }
        defer { Task { @MainActor in isLoading = false } }

        let request = LoginRequest(email: email, password: password)
        let response = try await apiClient.login(request)
        await saveTokens(response.data.token)

        await MainActor.run {
            self.currentUser = response.data.user
            self.isAuthenticated = true
        }
    }

    func logout() async throws {
        await MainActor.run { isLoading = true }
        defer { Task { @MainActor in isLoading = false } }

        try await apiClient.logout()
        await clearTokens()

        await MainActor.run {
            self.currentUser = nil
            self.isAuthenticated = false
        }
    }

    // MARK: - OAuth

    func loginWithGoogle(code: String, codeVerifier: String) async throws {
        await MainActor.run { isLoading = true }
        defer { Task { @MainActor in isLoading = false } }

        let request = OAuthRequest(code: code, codeVerifier: codeVerifier)
        let response = try await apiClient.loginWithGoogle(request)
        await saveTokens(response.data.token)

        await MainActor.run {
            self.currentUser = response.data.user
            self.isAuthenticated = true
        }
    }

    func loginWithApple(code: String, codeVerifier: String) async throws {
        await MainActor.run { isLoading = true }
        defer { Task { @MainActor in isLoading = false } }

        let request = OAuthRequest(code: code, codeVerifier: codeVerifier)
        let response = try await apiClient.loginWithApple(request)
        await saveTokens(response.data.token)

        await MainActor.run {
            self.currentUser = response.data.user
            self.isAuthenticated = true
        }
    }
}

// MARK: - PKCE Helper

struct PKCEHelper {
    let codeVerifier: String
    let codeChallenge: String

    init() {
        // Generate random 32-byte code verifier
        var buffer = [UInt8](repeating: 0, count: 32)
        _ = SecRandomCopyBytes(kSecRandomDefault, buffer.count, &buffer)
        codeVerifier = Data(buffer).base64URLEncoded()

        // SHA256 hash and base64url encode for challenge
        let data = Data(codeVerifier.utf8)
        var hash = [UInt8](repeating: 0, count: 32)
        data.withUnsafeBytes { buffer in
            _ = CC_SHA256(buffer.baseAddress, CC_LONG(buffer.count), &hash)
        }
        codeChallenge = Data(hash).base64URLEncoded()
    }
}

// CC_SHA256 import
import CommonCrypto

extension Data {
    func base64URLEncoded() -> String {
        base64EncodedString()
            .replacingOccurrences(of: "+", with: "-")
            .replacingOccurrences(of: "/", with: "_")
            .replacingOccurrences(of: "=", with: "")
    }
}
