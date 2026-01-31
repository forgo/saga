import Foundation

// MARK: - Generic Response Wrappers

/// Wraps a single data item response
struct DataResponse<T: Codable>: Codable where T: Sendable {
    let data: T
    let links: [String: String]?

    enum CodingKeys: String, CodingKey {
        case data
        case links = "_links"
    }
}

/// Wraps a collection of items with optional pagination
struct CollectionResponse<T: Codable>: Codable where T: Sendable {
    let data: [T]
    let pagination: PaginationInfo?
    let links: [String: String]?

    enum CodingKeys: String, CodingKey {
        case data, pagination
        case links = "_links"
    }
}

/// Pagination metadata for collection responses
struct PaginationInfo: Codable, Sendable {
    let cursor: String?
    let hasMore: Bool?

    enum CodingKeys: String, CodingKey {
        case cursor
        case hasMore = "has_more"
    }
}

// MARK: - Auth Responses

/// Response from login/register endpoints containing user and tokens
struct AuthResponse: Codable, Sendable {
    let user: User
    let token: TokenResponse
}

/// Token pair returned from authentication
struct TokenResponse: Codable, Sendable {
    let accessToken: String
    let refreshToken: String
    let tokenType: String
    let expiresIn: Int

    enum CodingKeys: String, CodingKey {
        case accessToken = "access_token"
        case refreshToken = "refresh_token"
        case tokenType = "token_type"
        case expiresIn = "expires_in"
    }
}

// MARK: - Passkey Responses

/// Challenge response for passkey operations
struct PasskeyChallengeResponse: Codable, Sendable {
    let challenge: String
    let timeout: Int?
    let rpId: String?
    let allowCredentials: [AllowedCredential]?

    enum CodingKeys: String, CodingKey {
        case challenge, timeout
        case rpId = "rp_id"
        case allowCredentials = "allow_credentials"
    }
}

struct AllowedCredential: Codable, Sendable {
    let id: String
    let type: String
}

/// Response from passkey registration
struct PasskeyRegistrationResponse: Codable, Sendable {
    let passkey: Passkey
}

// MARK: - Message Response

/// Simple message response
struct MessageResponse: Codable, Sendable {
    let message: String
}
