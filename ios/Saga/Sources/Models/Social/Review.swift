import Foundation

// MARK: - Review

/// A review of a user from an event or interaction
struct Review: Codable, Identifiable, Hashable, Sendable {
    let id: String
    let reviewerId: String
    let revieweeId: String
    let eventId: String?
    let rating: Int
    let content: String?
    let visibility: ReviewVisibility
    let createdOn: Date
    let updatedOn: Date?

    enum CodingKeys: String, CodingKey {
        case id
        case reviewerId = "reviewer_id"
        case revieweeId = "reviewee_id"
        case eventId = "event_id"
        case rating
        case content
        case visibility
        case createdOn = "created_on"
        case updatedOn = "updated_on"
    }

    /// Star display for rating (1-5)
    var starDisplay: String {
        String(repeating: "★", count: rating) + String(repeating: "☆", count: max(0, 5 - rating))
    }
}

// MARK: - Review With Author

/// A review with the reviewer's public profile
struct ReviewWithAuthor: Codable, Sendable {
    let review: Review
    let author: PublicProfile
}

// MARK: - Review Summary

/// Aggregated review statistics for a user
struct ReviewSummary: Codable, Sendable {
    let userId: String
    let totalReviews: Int
    let averageRating: Double
    let ratingBreakdown: [Int: Int]

    enum CodingKeys: String, CodingKey {
        case userId = "user_id"
        case totalReviews = "total_reviews"
        case averageRating = "average_rating"
        case ratingBreakdown = "rating_breakdown"
    }

    /// Formatted average rating
    var formattedRating: String {
        String(format: "%.1f", averageRating)
    }
}

// MARK: - Request Types

struct CreateReviewRequest: Codable, Sendable {
    let revieweeId: String
    let eventId: String?
    let rating: Int
    let content: String?
    let visibility: ReviewVisibility

    enum CodingKeys: String, CodingKey {
        case revieweeId = "reviewee_id"
        case eventId = "event_id"
        case rating
        case content
        case visibility
    }
}

struct UpdateReviewRequest: Codable, Sendable {
    var rating: Int?
    var content: String?
    var visibility: ReviewVisibility?
}
