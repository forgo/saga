import Foundation

// MARK: - Role Catalog

/// A catalog of reusable roles for events
struct RoleCatalog: Codable, Identifiable, Hashable, Sendable {
    let id: String
    let guildId: String
    let name: String
    let description: String?
    let iconName: String?
    let category: RoleCatalogCategory
    let suggestedSkills: [String]?
    let isActive: Bool
    let usageCount: Int
    let createdOn: Date
    let updatedOn: Date

    enum CodingKeys: String, CodingKey {
        case id
        case guildId = "guild_id"
        case name
        case description
        case iconName = "icon_name"
        case category
        case suggestedSkills = "suggested_skills"
        case isActive = "is_active"
        case usageCount = "usage_count"
        case createdOn = "created_on"
        case updatedOn = "updated_on"
    }
}

// MARK: - Role Catalog Category

enum RoleCatalogCategory: String, Codable, Sendable, CaseIterable {
    case logistics
    case hospitality
    case technical
    case creative
    case leadership
    case support
    case other

    var displayName: String {
        switch self {
        case .logistics: return "Logistics"
        case .hospitality: return "Hospitality"
        case .technical: return "Technical"
        case .creative: return "Creative"
        case .leadership: return "Leadership"
        case .support: return "Support"
        case .other: return "Other"
        }
    }

    var description: String {
        switch self {
        case .logistics: return "Planning, organization, and coordination"
        case .hospitality: return "Welcoming and caring for guests"
        case .technical: return "AV, equipment, and tech support"
        case .creative: return "Design, decoration, and entertainment"
        case .leadership: return "Leading and facilitating"
        case .support: return "General assistance and backup"
        case .other: return "Other roles"
        }
    }

    var iconName: String {
        switch self {
        case .logistics: return "list.clipboard"
        case .hospitality: return "hand.wave"
        case .technical: return "wrench.and.screwdriver"
        case .creative: return "paintbrush"
        case .leadership: return "star"
        case .support: return "hands.sparkles"
        case .other: return "ellipsis.circle"
        }
    }
}

// MARK: - Role Template

/// A template for quickly creating event roles from catalog
struct RoleTemplate: Codable, Identifiable, Hashable, Sendable {
    let id: String
    let catalogId: String
    let name: String
    let description: String?
    let defaultCapacity: Int
    let responsibilities: [String]?
    let requirements: [String]?

    enum CodingKeys: String, CodingKey {
        case id
        case catalogId = "catalog_id"
        case name
        case description
        case defaultCapacity = "default_capacity"
        case responsibilities
        case requirements
    }
}

// MARK: - Request Types

struct CreateRoleCatalogRequest: Codable, Sendable {
    let guildId: String
    let name: String
    let description: String?
    let iconName: String?
    let category: RoleCatalogCategory
    let suggestedSkills: [String]?

    enum CodingKeys: String, CodingKey {
        case guildId = "guild_id"
        case name
        case description
        case iconName = "icon_name"
        case category
        case suggestedSkills = "suggested_skills"
    }
}

struct UpdateRoleCatalogRequest: Codable, Sendable {
    var name: String?
    var description: String?
    var iconName: String?
    var category: RoleCatalogCategory?
    var suggestedSkills: [String]?
    var isActive: Bool?

    enum CodingKeys: String, CodingKey {
        case name
        case description
        case iconName = "icon_name"
        case category
        case suggestedSkills = "suggested_skills"
        case isActive = "is_active"
    }
}
