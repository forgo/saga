import Foundation

// MARK: - Generic Response Wrappers

struct DataResponse<T: Codable>: Codable where T: Sendable {
    let data: T
    let links: [String: String]?

    enum CodingKeys: String, CodingKey {
        case data
        case links = "_links"
    }
}

struct CollectionResponse<T: Codable>: Codable where T: Sendable {
    let data: [T]
    let pagination: PaginationInfo?
    let links: [String: String]?

    enum CodingKeys: String, CodingKey {
        case data, pagination
        case links = "_links"
    }
}

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

// MARK: - Error Response (RFC 9457 Problem Details)

struct ProblemDetails: Codable, Error, Sendable {
    let type: String
    let title: String
    let status: Int
    var detail: String?
    var instance: String?
    var errors: [FieldError]?
    var code: Int?
    var limit: Int?
    var current: Int?

    var localizedDescription: String {
        detail ?? title
    }
}

struct FieldError: Codable, Sendable {
    let field: String
    let message: String
}

// MARK: - API Error

enum APIError: Error, LocalizedError {
    case invalidURL
    case networkError(Error)
    case decodingError(Error)
    case httpError(statusCode: Int, data: Data?)
    case problemDetails(ProblemDetails)
    case unauthorized
    case forbidden
    case notFound
    case rateLimited(retryAfter: Int)
    case serverError
    case unknown

    var errorDescription: String? {
        switch self {
        case .invalidURL:
            return "Invalid URL"
        case .networkError(let error):
            return "Network error: \(error.localizedDescription)"
        case .decodingError(let error):
            return "Failed to decode response: \(error.localizedDescription)"
        case .httpError(let statusCode, _):
            return "HTTP error: \(statusCode)"
        case .problemDetails(let problem):
            return problem.detail ?? problem.title
        case .unauthorized:
            return "Unauthorized - please log in"
        case .forbidden:
            return "Access denied"
        case .notFound:
            return "Resource not found"
        case .rateLimited(let retryAfter):
            return "Too many requests. Retry after \(retryAfter) seconds"
        case .serverError:
            return "Server error - please try again later"
        case .unknown:
            return "An unknown error occurred"
        }
    }
}
