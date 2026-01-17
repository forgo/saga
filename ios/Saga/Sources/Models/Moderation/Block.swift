import Foundation

// MARK: - Block

/// A block relationship between users
struct Block: Codable, Identifiable, Hashable, Sendable {
    let id: String
    let blockerId: String
    let blockedId: String
    let reason: String?
    let createdOn: Date

    enum CodingKeys: String, CodingKey {
        case id
        case blockerId = "blocker_id"
        case blockedId = "blocked_id"
        case reason
        case createdOn = "created_on"
    }
}

// MARK: - Block Display

/// Block with user profile info
struct BlockDisplay: Codable, Identifiable, Sendable {
    let block: Block
    let blockedUser: PublicProfile

    var id: String { block.id }

    enum CodingKeys: String, CodingKey {
        case block
        case blockedUser = "blocked_user"
    }
}

// MARK: - Request Types

struct CreateBlockRequest: Codable, Sendable {
    let blockedId: String
    let reason: String?

    enum CodingKeys: String, CodingKey {
        case blockedId = "blocked_id"
        case reason
    }
}

// MARK: - Moderation Status

/// Current moderation status for a user
struct ModerationStatus: Codable, Sendable {
    let userId: String
    let isSuspended: Bool
    let suspendedUntil: Date?
    let suspensionReason: String?
    let warningCount: Int
    let lastWarningOn: Date?
    let restrictions: [ModerationRestriction]

    enum CodingKeys: String, CodingKey {
        case userId = "user_id"
        case isSuspended = "is_suspended"
        case suspendedUntil = "suspended_until"
        case suspensionReason = "suspension_reason"
        case warningCount = "warning_count"
        case lastWarningOn = "last_warning_on"
        case restrictions
    }

    /// Whether the user has any active restrictions
    var hasRestrictions: Bool {
        isSuspended || !restrictions.isEmpty
    }
}

// MARK: - Moderation Restriction

/// A restriction placed on a user's account
struct ModerationRestriction: Codable, Hashable, Sendable {
    let type: RestrictionType
    let reason: String?
    let expiresOn: Date?
    let createdOn: Date

    enum CodingKeys: String, CodingKey {
        case type
        case reason
        case expiresOn = "expires_on"
        case createdOn = "created_on"
    }

    /// Whether the restriction is still active
    var isActive: Bool {
        if let expiresOn = expiresOn {
            return expiresOn > Date()
        }
        return true
    }
}

// MARK: - Restriction Type

enum RestrictionType: String, Codable, Sendable, CaseIterable {
    case cannotMessage = "cannot_message"
    case cannotCreateEvents = "cannot_create_events"
    case cannotJoinEvents = "cannot_join_events"
    case cannotPost = "cannot_post"
    case limitedVisibility = "limited_visibility"

    var displayName: String {
        switch self {
        case .cannotMessage: return "Cannot Send Messages"
        case .cannotCreateEvents: return "Cannot Create Events"
        case .cannotJoinEvents: return "Cannot Join Events"
        case .cannotPost: return "Cannot Post Content"
        case .limitedVisibility: return "Limited Profile Visibility"
        }
    }

    var iconName: String {
        switch self {
        case .cannotMessage: return "message.badge.slash"
        case .cannotCreateEvents: return "calendar.badge.minus"
        case .cannotJoinEvents: return "person.badge.minus"
        case .cannotPost: return "square.and.pencil.slash"
        case .limitedVisibility: return "eye.slash"
        }
    }
}
