import Foundation

// MARK: - Registration & Login

/// Request to register a new user with email/password
struct RegisterRequest: Codable, Sendable {
    let email: String
    let password: String
    let firstname: String?
    let lastname: String?
}

/// Request to login with email/password
struct LoginRequest: Codable, Sendable {
    let email: String
    let password: String
}

/// Request to refresh access token
struct RefreshRequest: Codable, Sendable {
    let refreshToken: String

    enum CodingKeys: String, CodingKey {
        case refreshToken = "refresh_token"
    }
}

// MARK: - OAuth

/// Request to complete OAuth flow (Google or Apple)
struct OAuthRequest: Codable, Sendable {
    let code: String
    let codeVerifier: String
    let redirectUri: String?

    enum CodingKeys: String, CodingKey {
        case code
        case codeVerifier = "code_verifier"
        case redirectUri = "redirect_uri"
    }

    init(code: String, codeVerifier: String, redirectUri: String? = nil) {
        self.code = code
        self.codeVerifier = codeVerifier
        self.redirectUri = redirectUri
    }
}

// MARK: - Passkeys

/// Request to begin passkey registration
struct PasskeyBeginRegisterRequest: Codable, Sendable {
    let displayName: String?

    enum CodingKeys: String, CodingKey {
        case displayName = "display_name"
    }
}

/// Request to complete passkey registration
struct PasskeyCompleteRegisterRequest: Codable, Sendable {
    let credential: PasskeyCredential
}

/// Request to begin passkey login
struct PasskeyBeginLoginRequest: Codable, Sendable {
    let email: String?
}

/// Request to complete passkey login
struct PasskeyCompleteLoginRequest: Codable, Sendable {
    let credential: PasskeyCredential
}

/// WebAuthn credential data
struct PasskeyCredential: Codable, Sendable {
    let id: String
    let rawId: String
    let type: String
    let response: PasskeyCredentialResponse

    enum CodingKeys: String, CodingKey {
        case id
        case rawId = "raw_id"
        case type
        case response
    }
}

/// WebAuthn credential response data
struct PasskeyCredentialResponse: Codable, Sendable {
    let clientDataJSON: String
    let authenticatorData: String?
    let signature: String?
    let attestationObject: String?
    let userHandle: String?

    enum CodingKeys: String, CodingKey {
        case clientDataJSON = "client_data_json"
        case authenticatorData = "authenticator_data"
        case signature
        case attestationObject = "attestation_object"
        case userHandle = "user_handle"
    }
}

// MARK: - Account Linking

/// Request to link an OAuth identity to existing account
struct LinkOAuthRequest: Codable, Sendable {
    let provider: String
    let code: String
    let codeVerifier: String

    enum CodingKeys: String, CodingKey {
        case provider
        case code
        case codeVerifier = "code_verifier"
    }
}
