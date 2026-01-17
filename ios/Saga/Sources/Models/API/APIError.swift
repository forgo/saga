import Foundation

// MARK: - Error Response (RFC 9457 Problem Details)

/// RFC 9457 Problem Details error response
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

/// Individual field validation error
struct FieldError: Codable, Sendable {
    let field: String
    let message: String
}

// MARK: - API Error

/// API error types
enum APIError: Error, LocalizedError {
    case invalidURL
    case networkError(Error)
    case decodingError(Error)
    case httpError(statusCode: Int, data: Data?)
    case problemDetails(ProblemDetails)
    case unauthorized
    case forbidden
    case notFound
    case conflict
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
        case .conflict:
            return "Resource conflict"
        case .rateLimited(let retryAfter):
            return "Too many requests. Retry after \(retryAfter) seconds"
        case .serverError:
            return "Server error - please try again later"
        case .unknown:
            return "An unknown error occurred"
        }
    }

    /// User-friendly error message for display
    var userMessage: String {
        switch self {
        case .networkError:
            return "Unable to connect. Please check your internet connection."
        case .unauthorized:
            return "Please sign in to continue."
        case .forbidden:
            return "You don't have permission to do that."
        case .notFound:
            return "The requested item was not found."
        case .rateLimited:
            return "You're doing that too fast. Please wait a moment."
        case .serverError:
            return "Something went wrong. Please try again."
        case .problemDetails(let problem):
            return problem.detail ?? problem.title
        default:
            return errorDescription ?? "An error occurred."
        }
    }
}
