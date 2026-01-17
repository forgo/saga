import Foundation

struct RegisterRequest: Codable, Sendable {
    let email: String
    let password: String
    var firstname: String?
    var lastname: String?
}

struct LoginRequest: Codable, Sendable {
    let email: String
    let password: String
}

struct OAuthRequest: Codable, Sendable {
    let code: String
    let codeVerifier: String

    enum CodingKeys: String, CodingKey {
        case code
        case codeVerifier = "code_verifier"
    }
}

struct RefreshRequest: Codable, Sendable {
    let refreshToken: String

    enum CodingKeys: String, CodingKey {
        case refreshToken = "refresh_token"
    }
}

struct PasskeyLoginStartRequest: Codable, Sendable {
    var email: String?
}

struct PasskeyLoginFinishRequest: Codable, Sendable {
    let credential: PasskeyCredential
}

struct PasskeyRegistrationFinishRequest: Codable, Sendable {
    let credential: PasskeyCredential
    let name: String
}

struct PasskeyCredential: Codable, Sendable {
    let id: String
    let rawId: String
    let type: String
    let response: PasskeyResponse

    enum CodingKeys: String, CodingKey {
        case id
        case rawId = "rawId"
        case type
        case response
    }
}

struct PasskeyResponse: Codable, Sendable {
    let clientDataJSON: String
    let authenticatorData: String?
    let signature: String?
    let attestationObject: String?
    let userHandle: String?
}
