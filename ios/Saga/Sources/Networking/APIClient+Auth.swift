import Foundation

// MARK: - Auth API Extension

extension APIClient {

    // MARK: - Email/Password Auth

    /// Register a new user with email and password
    func register(_ request: RegisterRequest) async throws -> DataResponse<AuthResponse> {
        try await post(path: "auth/register", body: request, requiresAuth: false)
    }

    /// Login with email and password
    func login(_ request: LoginRequest) async throws -> DataResponse<AuthResponse> {
        try await post(path: "auth/login", body: request, requiresAuth: false)
    }

    /// Logout and invalidate tokens
    func logout() async throws {
        try await postNoContent(path: "auth/logout")
        clearTokens()
    }

    /// Get current authenticated user
    func getCurrentUser() async throws -> DataResponse<UserWithIdentities> {
        try await get(path: "auth/me")
    }

    // MARK: - OAuth

    /// Login/register with Google OAuth
    func loginWithGoogle(_ request: OAuthRequest) async throws -> DataResponse<AuthResponse> {
        try await post(path: "auth/oauth/google", body: request, requiresAuth: false)
    }

    /// Login/register with Apple OAuth
    func loginWithApple(_ request: OAuthRequest) async throws -> DataResponse<AuthResponse> {
        try await post(path: "auth/oauth/apple", body: request, requiresAuth: false)
    }

    // MARK: - Passkeys

    /// Begin passkey registration - get challenge
    func beginPasskeyRegistration(displayName: String? = nil) async throws -> DataResponse<PasskeyChallengeResponse> {
        let request = PasskeyBeginRegisterRequest(displayName: displayName)
        return try await post(path: "auth/passkeys/register/begin", body: request)
    }

    /// Complete passkey registration with credential
    func completePasskeyRegistration(credential: PasskeyCredential) async throws -> DataResponse<PasskeyRegistrationResponse> {
        let request = PasskeyCompleteRegisterRequest(credential: credential)
        return try await post(path: "auth/passkeys/register/complete", body: request)
    }

    /// Begin passkey login - get challenge
    func beginPasskeyLogin(email: String? = nil) async throws -> DataResponse<PasskeyChallengeResponse> {
        let request = PasskeyBeginLoginRequest(email: email)
        return try await post(path: "auth/passkeys/login/begin", body: request, requiresAuth: false)
    }

    /// Complete passkey login with assertion
    func completePasskeyLogin(credential: PasskeyCredential) async throws -> DataResponse<AuthResponse> {
        let request = PasskeyCompleteLoginRequest(credential: credential)
        return try await post(path: "auth/passkeys/login/complete", body: request, requiresAuth: false)
    }

    /// List user's registered passkeys
    func listPasskeys() async throws -> CollectionResponse<Passkey> {
        try await get(path: "auth/passkeys")
    }

    /// Delete a passkey
    func deletePasskey(id: String) async throws {
        try await delete(path: "auth/passkeys/\(id)")
    }

    // MARK: - Account Linking

    /// Link an OAuth identity to the current account
    func linkOAuthIdentity(provider: String, code: String, codeVerifier: String) async throws -> DataResponse<Identity> {
        let request = LinkOAuthRequest(provider: provider, code: code, codeVerifier: codeVerifier)
        return try await post(path: "auth/link/\(provider)", body: request)
    }

    /// Unlink an identity from the account
    func unlinkIdentity(id: String) async throws {
        try await delete(path: "auth/identities/\(id)")
    }

    // MARK: - Token Management

    /// Refresh the access token using refresh token
    func refreshTokens(refreshToken: String) async throws -> DataResponse<TokenResponse> {
        let request = RefreshRequest(refreshToken: refreshToken)
        return try await post(path: "auth/refresh", body: request, requiresAuth: false)
    }
}
