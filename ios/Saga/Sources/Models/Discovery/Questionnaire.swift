import Foundation

// MARK: - Questionnaire

/// A questionnaire for compatibility matching
struct Questionnaire: Codable, Identifiable, Hashable, Sendable {
    let id: String
    let name: String
    let description: String?
    let category: QuestionnaireCategory
    let questions: [Question]?
    let isRequired: Bool
    let sortOrder: Int
    let createdOn: Date

    enum CodingKeys: String, CodingKey {
        case id
        case name
        case description
        case category
        case questions
        case isRequired = "is_required"
        case sortOrder = "sort_order"
        case createdOn = "created_on"
    }
}

// MARK: - Questionnaire Category

enum QuestionnaireCategory: String, Codable, Sendable, CaseIterable {
    case personality
    case values
    case lifestyle
    case social
    case interests

    var displayName: String {
        switch self {
        case .personality: return "Personality"
        case .values: return "Values"
        case .lifestyle: return "Lifestyle"
        case .social: return "Social Preferences"
        case .interests: return "Interests"
        }
    }

    var iconName: String {
        switch self {
        case .personality: return "brain.head.profile"
        case .values: return "heart.fill"
        case .lifestyle: return "house.fill"
        case .social: return "person.3.fill"
        case .interests: return "star.fill"
        }
    }
}

// MARK: - Question

/// A single question in a questionnaire
struct Question: Codable, Identifiable, Hashable, Sendable {
    let id: String
    let questionnaireId: String
    let text: String
    let questionType: QuestionType
    let options: [QuestionOption]?
    let minValue: Int?
    let maxValue: Int?
    let sortOrder: Int

    enum CodingKeys: String, CodingKey {
        case id
        case questionnaireId = "questionnaire_id"
        case text
        case questionType = "question_type"
        case options
        case minValue = "min_value"
        case maxValue = "max_value"
        case sortOrder = "sort_order"
    }
}

// MARK: - Question Type

enum QuestionType: String, Codable, Sendable {
    case multipleChoice = "multiple_choice"
    case scale
    case text
    case multiSelect = "multi_select"

    var displayName: String {
        switch self {
        case .multipleChoice: return "Multiple Choice"
        case .scale: return "Scale"
        case .text: return "Text"
        case .multiSelect: return "Multi-Select"
        }
    }
}

// MARK: - Question Option

/// An option for multiple choice questions
struct QuestionOption: Codable, Identifiable, Hashable, Sendable {
    let id: String
    let text: String
    let value: Int
    let sortOrder: Int

    enum CodingKeys: String, CodingKey {
        case id
        case text
        case value
        case sortOrder = "sort_order"
    }
}

// MARK: - Question Response

/// A user's response to a question
struct QuestionResponse: Codable, Identifiable, Hashable, Sendable {
    let id: String
    let userId: String
    let questionId: String
    let optionId: String?
    let scaleValue: Int?
    let textValue: String?
    let selectedOptions: [String]?
    let createdOn: Date
    let updatedOn: Date?

    enum CodingKeys: String, CodingKey {
        case id
        case userId = "user_id"
        case questionId = "question_id"
        case optionId = "option_id"
        case scaleValue = "scale_value"
        case textValue = "text_value"
        case selectedOptions = "selected_options"
        case createdOn = "created_on"
        case updatedOn = "updated_on"
    }
}

// MARK: - Questionnaire Progress

/// User's progress on questionnaires
struct QuestionnaireProgress: Codable, Sendable {
    let questionnaireId: String
    let questionnaireName: String
    let totalQuestions: Int
    let answeredQuestions: Int
    let isComplete: Bool

    enum CodingKeys: String, CodingKey {
        case questionnaireId = "questionnaire_id"
        case questionnaireName = "questionnaire_name"
        case totalQuestions = "total_questions"
        case answeredQuestions = "answered_questions"
        case isComplete = "is_complete"
    }

    var progressPercentage: Double {
        guard totalQuestions > 0 else { return 0 }
        return Double(answeredQuestions) / Double(totalQuestions)
    }
}

// MARK: - Compatibility Score

/// Compatibility score between two users
struct CompatibilityScore: Codable, Sendable {
    let userId: String
    let targetUserId: String
    let overallScore: Double
    let categoryScores: [CategoryScore]
    let calculatedOn: Date

    enum CodingKeys: String, CodingKey {
        case userId = "user_id"
        case targetUserId = "target_user_id"
        case overallScore = "overall_score"
        case categoryScores = "category_scores"
        case calculatedOn = "calculated_on"
    }

    /// Formatted overall score as percentage
    var formattedScore: String {
        "\(Int(overallScore * 100))%"
    }

    /// Description of compatibility level
    var compatibilityLevel: String {
        if overallScore >= 0.8 { return "Highly Compatible" }
        if overallScore >= 0.6 { return "Very Compatible" }
        if overallScore >= 0.4 { return "Compatible" }
        if overallScore >= 0.2 { return "Somewhat Compatible" }
        return "Low Compatibility"
    }
}

// MARK: - Category Score

/// Score for a specific questionnaire category
struct CategoryScore: Codable, Sendable {
    let category: QuestionnaireCategory
    let score: Double
    let weight: Double

    var formattedScore: String {
        "\(Int(score * 100))%"
    }
}

// MARK: - Discovery Result

/// A discovered user with compatibility info
struct DiscoveryResult: Codable, Sendable {
    let user: PublicProfile
    let compatibilityScore: Double?
    let distanceKm: Double?
    let sharedInterests: Int
    let mutualConnections: Int

    enum CodingKeys: String, CodingKey {
        case user
        case compatibilityScore = "compatibility_score"
        case distanceKm = "distance_km"
        case sharedInterests = "shared_interests"
        case mutualConnections = "mutual_connections"
    }

    var formattedCompatibility: String? {
        guard let score = compatibilityScore else { return nil }
        return "\(Int(score * 100))% match"
    }

    var formattedDistance: String? {
        guard let km = distanceKm else { return nil }
        if km < 1 { return "< 1 km" }
        return String(format: "%.1f km", km)
    }
}

// MARK: - Request Types

struct SubmitResponseRequest: Codable, Sendable {
    let questionId: String
    let optionId: String?
    let scaleValue: Int?
    let textValue: String?
    let selectedOptions: [String]?

    enum CodingKeys: String, CodingKey {
        case questionId = "question_id"
        case optionId = "option_id"
        case scaleValue = "scale_value"
        case textValue = "text_value"
        case selectedOptions = "selected_options"
    }
}

struct DiscoverPeopleRequest: Codable, Sendable {
    let lat: Double?
    let lng: Double?
    let radiusKm: Double?
    let minCompatibility: Double?
    let interestIds: [String]?
    let limit: Int

    enum CodingKeys: String, CodingKey {
        case lat
        case lng
        case radiusKm = "radius_km"
        case minCompatibility = "min_compatibility"
        case interestIds = "interest_ids"
        case limit
    }
}
