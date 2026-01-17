import Foundation

// MARK: - Pool

/// A matching pool for donut-style introductions
struct Pool: Codable, Identifiable, Hashable, Sendable {
    let id: String
    let guildId: String
    let name: String
    let description: String?
    let matchingFrequency: MatchingFrequency
    let matchSize: Int
    let isActive: Bool
    let nextMatchDate: Date?
    let memberCount: Int
    let createdOn: Date
    let updatedOn: Date

    enum CodingKeys: String, CodingKey {
        case id
        case guildId = "guild_id"
        case name
        case description
        case matchingFrequency = "matching_frequency"
        case matchSize = "match_size"
        case isActive = "is_active"
        case nextMatchDate = "next_match_date"
        case memberCount = "member_count"
        case createdOn = "created_on"
        case updatedOn = "updated_on"
    }

    /// Days until next match
    var daysUntilNextMatch: Int? {
        guard let nextMatch = nextMatchDate else { return nil }
        let calendar = Calendar.current
        let days = calendar.dateComponents([.day], from: Date(), to: nextMatch).day
        return days
    }
}

// MARK: - Matching Frequency

enum MatchingFrequency: String, Codable, Sendable, CaseIterable {
    case daily
    case weekly
    case biweekly
    case monthly

    var displayName: String {
        switch self {
        case .daily: return "Daily"
        case .weekly: return "Weekly"
        case .biweekly: return "Bi-Weekly"
        case .monthly: return "Monthly"
        }
    }

    var description: String {
        switch self {
        case .daily: return "Match every day"
        case .weekly: return "Match every week"
        case .biweekly: return "Match every two weeks"
        case .monthly: return "Match every month"
        }
    }
}

// MARK: - Pool Member

/// A member of a matching pool
struct PoolMember: Codable, Identifiable, Hashable, Sendable {
    let id: String
    let poolId: String
    let userId: String
    let isActive: Bool
    let preferences: PoolPreferences?
    let joinedOn: Date
    let lastMatchedOn: Date?

    enum CodingKeys: String, CodingKey {
        case id
        case poolId = "pool_id"
        case userId = "user_id"
        case isActive = "is_active"
        case preferences
        case joinedOn = "joined_on"
        case lastMatchedOn = "last_matched_on"
    }
}

// MARK: - Pool Preferences

/// Member preferences for pool matching
struct PoolPreferences: Codable, Hashable, Sendable {
    let availableDays: [Int]?
    let preferredTimes: [String]?
    let excludeRecentMatches: Bool
    let notes: String?

    enum CodingKeys: String, CodingKey {
        case availableDays = "available_days"
        case preferredTimes = "preferred_times"
        case excludeRecentMatches = "exclude_recent_matches"
        case notes
    }
}

// MARK: - Pool Match

/// A match result from a pool
struct PoolMatch: Codable, Identifiable, Hashable, Sendable {
    let id: String
    let poolId: String
    let matchDate: Date
    let participants: [String]
    let status: MatchStatus
    let meetingScheduled: Date?
    let feedbackSubmitted: Bool
    let createdOn: Date

    enum CodingKeys: String, CodingKey {
        case id
        case poolId = "pool_id"
        case matchDate = "match_date"
        case participants
        case status
        case meetingScheduled = "meeting_scheduled"
        case feedbackSubmitted = "feedback_submitted"
        case createdOn = "created_on"
    }
}

// MARK: - Match Status

enum MatchStatus: String, Codable, Sendable, CaseIterable {
    case pending
    case accepted
    case scheduled
    case completed
    case skipped

    var displayName: String {
        switch self {
        case .pending: return "Pending"
        case .accepted: return "Accepted"
        case .scheduled: return "Scheduled"
        case .completed: return "Completed"
        case .skipped: return "Skipped"
        }
    }

    var iconName: String {
        switch self {
        case .pending: return "clock"
        case .accepted: return "checkmark"
        case .scheduled: return "calendar"
        case .completed: return "checkmark.circle.fill"
        case .skipped: return "forward.fill"
        }
    }
}

// MARK: - Pool Match Display

/// Match with participant profiles
struct PoolMatchDisplay: Codable, Sendable {
    let match: PoolMatch
    let participants: [PublicProfile]
    let pool: Pool
}

// MARK: - Pool With Details

/// Pool with member info
struct PoolWithDetails: Codable, Sendable {
    let pool: Pool
    let myMembership: PoolMember?
    let recentMatches: [PoolMatchDisplay]?

    enum CodingKeys: String, CodingKey {
        case pool
        case myMembership = "my_membership"
        case recentMatches = "recent_matches"
    }
}

// MARK: - Request Types

struct CreatePoolRequest: Codable, Sendable {
    let guildId: String
    let name: String
    let description: String?
    let matchingFrequency: MatchingFrequency
    let matchSize: Int

    enum CodingKeys: String, CodingKey {
        case guildId = "guild_id"
        case name
        case description
        case matchingFrequency = "matching_frequency"
        case matchSize = "match_size"
    }
}

struct UpdatePoolRequest: Codable, Sendable {
    var name: String?
    var description: String?
    var matchingFrequency: MatchingFrequency?
    var matchSize: Int?
    var isActive: Bool?

    enum CodingKeys: String, CodingKey {
        case name
        case description
        case matchingFrequency = "matching_frequency"
        case matchSize = "match_size"
        case isActive = "is_active"
    }
}

struct JoinPoolRequest: Codable, Sendable {
    let preferences: PoolPreferences?
}

struct UpdatePoolPreferencesRequest: Codable, Sendable {
    let preferences: PoolPreferences
}

struct RespondToMatchRequest: Codable, Sendable {
    let accept: Bool
}

struct ScheduleMatchRequest: Codable, Sendable {
    let meetingTime: Date

    enum CodingKeys: String, CodingKey {
        case meetingTime = "meeting_time"
    }
}

struct MatchFeedbackRequest: Codable, Sendable {
    let metUp: Bool
    let rating: Int?
    let notes: String?

    enum CodingKeys: String, CodingKey {
        case metUp = "met_up"
        case rating
        case notes
    }
}
