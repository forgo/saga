import Foundation

// MARK: - Resonance

/// A user's resonance score and level
struct Resonance: Codable, Sendable {
    let userId: String
    let level: ResonanceLevel
    let totalScore: Int
    let breakdown: ResonanceBreakdown
    let recentActivity: [ResonanceActivity]
    let updatedOn: Date

    enum CodingKeys: String, CodingKey {
        case userId = "user_id"
        case level
        case totalScore = "total_score"
        case breakdown
        case recentActivity = "recent_activity"
        case updatedOn = "updated_on"
    }

    /// Progress to next level (0.0 to 1.0)
    var progressToNextLevel: Double {
        let currentMin = level.minScore
        let nextMin = level.nextLevel?.minScore ?? (currentMin + 1000)
        let range = nextMin - currentMin
        let progress = totalScore - currentMin
        return min(1.0, max(0.0, Double(progress) / Double(range)))
    }

    /// Points needed for next level
    var pointsToNextLevel: Int? {
        guard let nextLevel = level.nextLevel else { return nil }
        return max(0, nextLevel.minScore - totalScore)
    }
}

// MARK: - Resonance Level

enum ResonanceLevel: String, Codable, Sendable, CaseIterable {
    case newcomer
    case explorer
    case contributor
    case connector
    case catalyst
    case luminary

    var displayName: String {
        switch self {
        case .newcomer: return "Newcomer"
        case .explorer: return "Explorer"
        case .contributor: return "Contributor"
        case .connector: return "Connector"
        case .catalyst: return "Catalyst"
        case .luminary: return "Luminary"
        }
    }

    var description: String {
        switch self {
        case .newcomer: return "Just getting started"
        case .explorer: return "Actively exploring the community"
        case .contributor: return "Making meaningful contributions"
        case .connector: return "Building strong connections"
        case .catalyst: return "Sparking community growth"
        case .luminary: return "A beacon in the community"
        }
    }

    var iconName: String {
        switch self {
        case .newcomer: return "leaf"
        case .explorer: return "binoculars"
        case .contributor: return "hand.raised"
        case .connector: return "link"
        case .catalyst: return "sparkles"
        case .luminary: return "sun.max.fill"
        }
    }

    var minScore: Int {
        switch self {
        case .newcomer: return 0
        case .explorer: return 100
        case .contributor: return 500
        case .connector: return 1500
        case .catalyst: return 5000
        case .luminary: return 15000
        }
    }

    var nextLevel: ResonanceLevel? {
        switch self {
        case .newcomer: return .explorer
        case .explorer: return .contributor
        case .contributor: return .connector
        case .connector: return .catalyst
        case .catalyst: return .luminary
        case .luminary: return nil
        }
    }

    var color: String {
        switch self {
        case .newcomer: return "#9CA3AF" // gray
        case .explorer: return "#10B981" // green
        case .contributor: return "#3B82F6" // blue
        case .connector: return "#8B5CF6" // purple
        case .catalyst: return "#F59E0B" // amber
        case .luminary: return "#EF4444" // red/gold
        }
    }
}

// MARK: - Resonance Breakdown

/// Breakdown of resonance score by category
struct ResonanceBreakdown: Codable, Sendable {
    let attendance: Int
    let hosting: Int
    let connections: Int
    let trust: Int
    let contributions: Int
    let engagement: Int

    /// Total of all categories
    var total: Int {
        attendance + hosting + connections + trust + contributions + engagement
    }

    /// Categories as array for display
    var categories: [(name: String, score: Int, icon: String)] {
        [
            ("Attendance", attendance, "calendar.badge.checkmark"),
            ("Hosting", hosting, "star.fill"),
            ("Connections", connections, "person.2.fill"),
            ("Trust", trust, "shield.fill"),
            ("Contributions", contributions, "hand.raised.fill"),
            ("Engagement", engagement, "bubble.left.and.bubble.right.fill")
        ]
    }
}

// MARK: - Resonance Activity

/// A recent activity that earned resonance points
struct ResonanceActivity: Codable, Identifiable, Sendable {
    let id: String
    let type: ResonanceActivityType
    let points: Int
    let description: String
    let createdOn: Date

    enum CodingKeys: String, CodingKey {
        case id
        case type
        case points
        case description
        case createdOn = "created_on"
    }
}

// MARK: - Resonance Activity Type

enum ResonanceActivityType: String, Codable, Sendable {
    case eventAttended = "event_attended"
    case eventHosted = "event_hosted"
    case connectionMade = "connection_made"
    case trustGranted = "trust_granted"
    case trustReceived = "trust_received"
    case irlConfirmed = "irl_confirmed"
    case reviewLeft = "review_left"
    case poolMatched = "pool_matched"
    case adventureJoined = "adventure_joined"
    case voteParticipated = "vote_participated"
    case profileCompleted = "profile_completed"
    case streakBonus = "streak_bonus"

    var displayName: String {
        switch self {
        case .eventAttended: return "Attended Event"
        case .eventHosted: return "Hosted Event"
        case .connectionMade: return "Made Connection"
        case .trustGranted: return "Granted Trust"
        case .trustReceived: return "Received Trust"
        case .irlConfirmed: return "IRL Confirmed"
        case .reviewLeft: return "Left Review"
        case .poolMatched: return "Pool Match"
        case .adventureJoined: return "Joined Adventure"
        case .voteParticipated: return "Voted"
        case .profileCompleted: return "Completed Profile"
        case .streakBonus: return "Streak Bonus"
        }
    }

    var iconName: String {
        switch self {
        case .eventAttended: return "calendar.badge.checkmark"
        case .eventHosted: return "star.fill"
        case .connectionMade: return "person.badge.plus"
        case .trustGranted: return "shield.fill"
        case .trustReceived: return "shield.checkered"
        case .irlConfirmed: return "person.2.fill"
        case .reviewLeft: return "text.bubble.fill"
        case .poolMatched: return "person.2.circle"
        case .adventureJoined: return "figure.hiking"
        case .voteParticipated: return "hand.raised.fill"
        case .profileCompleted: return "person.crop.circle.badge.checkmark"
        case .streakBonus: return "flame.fill"
        }
    }
}

// MARK: - Device Registration

/// Device registration for push notifications
struct Device: Codable, Identifiable, Hashable, Sendable {
    let id: String
    let userId: String
    let deviceToken: String
    let platform: DevicePlatform
    let name: String?
    let isActive: Bool
    let lastActiveOn: Date
    let createdOn: Date

    enum CodingKeys: String, CodingKey {
        case id
        case userId = "user_id"
        case deviceToken = "device_token"
        case platform
        case name
        case isActive = "is_active"
        case lastActiveOn = "last_active_on"
        case createdOn = "created_on"
    }
}

// MARK: - Device Platform

enum DevicePlatform: String, Codable, Sendable {
    case ios
    case android
    case web

    var displayName: String {
        switch self {
        case .ios: return "iOS"
        case .android: return "Android"
        case .web: return "Web"
        }
    }

    var iconName: String {
        switch self {
        case .ios: return "iphone"
        case .android: return "candybarphone"
        case .web: return "globe"
        }
    }
}

// MARK: - Request Types

struct RegisterDeviceRequest: Codable, Sendable {
    let deviceToken: String
    let platform: DevicePlatform
    let name: String?

    enum CodingKeys: String, CodingKey {
        case deviceToken = "device_token"
        case platform
        case name
    }
}

struct UpdateDeviceRequest: Codable, Sendable {
    var deviceToken: String?
    var name: String?
    var isActive: Bool?

    enum CodingKeys: String, CodingKey {
        case deviceToken = "device_token"
        case name
        case isActive = "is_active"
    }
}
