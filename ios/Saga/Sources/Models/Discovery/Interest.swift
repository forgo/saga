import Foundation

// MARK: - Interest Category

/// A category of interests
struct InterestCategory: Codable, Identifiable, Hashable, Sendable {
    let id: String
    let name: String
    let description: String?
    let iconName: String?
    let color: String?
    let parentId: String?
    let sortOrder: Int

    enum CodingKeys: String, CodingKey {
        case id
        case name
        case description
        case iconName = "icon_name"
        case color
        case parentId = "parent_id"
        case sortOrder = "sort_order"
    }
}

// MARK: - Interest

/// A specific interest within a category
struct Interest: Codable, Identifiable, Hashable, Sendable {
    let id: String
    let categoryId: String
    let name: String
    let description: String?
    let iconName: String?

    enum CodingKeys: String, CodingKey {
        case id
        case categoryId = "category_id"
        case name
        case description
        case iconName = "icon_name"
    }
}

// MARK: - User Interest

/// A user's interest with their skill level and intent
struct UserInterest: Codable, Identifiable, Hashable, Sendable {
    let id: String
    let userId: String
    let interestId: String
    let interestName: String?
    let categoryName: String?
    let skillLevel: SkillLevel
    let intent: InterestIntent
    let notes: String?
    let createdOn: Date

    enum CodingKeys: String, CodingKey {
        case id
        case userId = "user_id"
        case interestId = "interest_id"
        case interestName = "interest_name"
        case categoryName = "category_name"
        case skillLevel = "skill_level"
        case intent
        case notes
        case createdOn = "created_on"
    }
}

// MARK: - Skill Level

enum SkillLevel: String, Codable, Sendable, CaseIterable {
    case beginner
    case intermediate
    case advanced
    case expert

    var displayName: String {
        switch self {
        case .beginner: return "Beginner"
        case .intermediate: return "Intermediate"
        case .advanced: return "Advanced"
        case .expert: return "Expert"
        }
    }

    var description: String {
        switch self {
        case .beginner: return "Just starting out"
        case .intermediate: return "Some experience"
        case .advanced: return "Significant experience"
        case .expert: return "Deep expertise"
        }
    }

    var iconName: String {
        switch self {
        case .beginner: return "star"
        case .intermediate: return "star.leadinghalf.filled"
        case .advanced: return "star.fill"
        case .expert: return "star.circle.fill"
        }
    }
}

// MARK: - Interest Intent

enum InterestIntent: String, Codable, Sendable, CaseIterable {
    case teach
    case learn
    case both

    var displayName: String {
        switch self {
        case .teach: return "Teach"
        case .learn: return "Learn"
        case .both: return "Both"
        }
    }

    var description: String {
        switch self {
        case .teach: return "I want to teach others"
        case .learn: return "I want to learn from others"
        case .both: return "I want to teach and learn"
        }
    }

    var iconName: String {
        switch self {
        case .teach: return "person.badge.plus"
        case .learn: return "book.fill"
        case .both: return "arrow.left.arrow.right"
        }
    }
}

// MARK: - Interest Match

/// A potential match based on complementary interests
struct InterestMatch: Codable, Sendable {
    let user: PublicProfile
    let matchingInterests: [MatchingInterest]
    let compatibilityScore: Double

    enum CodingKeys: String, CodingKey {
        case user
        case matchingInterests = "matching_interests"
        case compatibilityScore = "compatibility_score"
    }

    /// Overall match quality description
    var matchQuality: String {
        if compatibilityScore >= 0.8 { return "Excellent Match" }
        if compatibilityScore >= 0.6 { return "Great Match" }
        if compatibilityScore >= 0.4 { return "Good Match" }
        return "Potential Match"
    }
}

// MARK: - Matching Interest

/// Details about how two users' interests match
struct MatchingInterest: Codable, Sendable {
    let interestId: String
    let interestName: String
    let categoryName: String?
    let yourIntent: InterestIntent
    let yourLevel: SkillLevel
    let theirIntent: InterestIntent
    let theirLevel: SkillLevel
    let matchType: MatchType

    enum CodingKeys: String, CodingKey {
        case interestId = "interest_id"
        case interestName = "interest_name"
        case categoryName = "category_name"
        case yourIntent = "your_intent"
        case yourLevel = "your_level"
        case theirIntent = "their_intent"
        case theirLevel = "their_level"
        case matchType = "match_type"
    }
}

// MARK: - Match Type

enum MatchType: String, Codable, Sendable {
    case teachLearn = "teach_learn"
    case learnTeach = "learn_teach"
    case mutual

    var displayName: String {
        switch self {
        case .teachLearn: return "You can teach them"
        case .learnTeach: return "They can teach you"
        case .mutual: return "Learn together"
        }
    }

    var iconName: String {
        switch self {
        case .teachLearn: return "arrow.right"
        case .learnTeach: return "arrow.left"
        case .mutual: return "arrow.left.arrow.right"
        }
    }
}

// MARK: - Request Types

struct AddInterestRequest: Codable, Sendable {
    let interestId: String
    let skillLevel: SkillLevel
    let intent: InterestIntent
    let notes: String?

    enum CodingKeys: String, CodingKey {
        case interestId = "interest_id"
        case skillLevel = "skill_level"
        case intent
        case notes
    }
}

struct UpdateInterestRequest: Codable, Sendable {
    var skillLevel: SkillLevel?
    var intent: InterestIntent?
    var notes: String?

    enum CodingKeys: String, CodingKey {
        case skillLevel = "skill_level"
        case intent
        case notes
    }
}
