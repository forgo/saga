import Foundation

// MARK: - Event Role

/// A role that can be assigned to attendees at an event
struct EventRole: Codable, Identifiable, Hashable, Sendable {
    let id: String
    let eventId: String
    let catalogRoleId: String?
    let name: String
    let description: String?
    let maxSlots: Int
    let filledSlots: Int

    enum CodingKeys: String, CodingKey {
        case id
        case eventId = "event_id"
        case catalogRoleId = "catalog_role_id"
        case name
        case description
        case maxSlots = "max_slots"
        case filledSlots = "filled_slots"
    }

    /// Whether this role has all slots filled
    var isFull: Bool {
        filledSlots >= maxSlots
    }

    /// Number of spots remaining
    var spotsLeft: Int {
        max(0, maxSlots - filledSlots)
    }

    /// Availability text like "2/5 filled"
    var availabilityText: String {
        "\(filledSlots)/\(maxSlots) filled"
    }
}

// MARK: - Role Assignment

/// An assignment of a user to a role
struct RoleAssignment: Codable, Identifiable, Hashable, Sendable {
    let id: String
    let eventId: String
    let roleId: String
    let userId: String
    let note: String?
    let status: RoleAssignmentStatus
    let assignedOn: Date

    enum CodingKeys: String, CodingKey {
        case id
        case eventId = "event_id"
        case roleId = "role_id"
        case userId = "user_id"
        case note
        case status
        case assignedOn = "assigned_on"
    }
}

// MARK: - Role Assignment Status

enum RoleAssignmentStatus: String, Codable, Sendable, CaseIterable {
    case assigned
    case confirmed
    case declined

    var displayName: String {
        switch self {
        case .assigned: return "Assigned"
        case .confirmed: return "Confirmed"
        case .declined: return "Declined"
        }
    }
}

// MARK: - Event Role With Assignments

/// A role bundled with its assignments
struct EventRoleWithAssignments: Codable, Sendable {
    let role: EventRole
    let assignments: [RoleAssignment]
    let isFull: Bool
    let spotsLeft: Int

    enum CodingKeys: String, CodingKey {
        case role
        case assignments
        case isFull = "is_full"
        case spotsLeft = "spots_left"
    }
}

// MARK: - Event Roles Overview

/// Overview of all roles for an event
struct EventRolesOverview: Codable, Sendable {
    let data: [EventRoleWithAssignments]
}

// MARK: - Role Suggestion

/// A suggested role based on user interests
struct RoleSuggestion: Codable, Identifiable, Sendable {
    let role: RoleCatalogEntry
    let reason: String
    let usageCount: Int

    var id: String { role.id }

    enum CodingKeys: String, CodingKey {
        case role
        case reason
        case usageCount = "usage_count"
    }
}

// MARK: - Role Catalog Entry

/// A reusable role from the catalog
struct RoleCatalogEntry: Codable, Identifiable, Hashable, Sendable {
    let id: String
    let scopeType: String
    let scopeId: String
    let roleType: String
    let name: String
    let description: String?
    let icon: String?
    let isActive: Bool
    let createdBy: String?
    let createdOn: Date

    enum CodingKeys: String, CodingKey {
        case id
        case scopeType = "scope_type"
        case scopeId = "scope_id"
        case roleType = "role_type"
        case name
        case description
        case icon
        case isActive = "is_active"
        case createdBy = "created_by"
        case createdOn = "created_on"
    }
}

// MARK: - Request Types

struct CreateEventRoleRequest: Codable, Sendable {
    let catalogRoleId: String?
    let name: String
    let description: String?
    let maxSlots: Int

    enum CodingKeys: String, CodingKey {
        case catalogRoleId = "catalog_role_id"
        case name
        case description
        case maxSlots = "max_slots"
    }

    init(name: String, description: String? = nil, maxSlots: Int = 1, catalogRoleId: String? = nil) {
        self.name = name
        self.description = description
        self.maxSlots = maxSlots
        self.catalogRoleId = catalogRoleId
    }
}

struct UpdateEventRoleRequest: Codable, Sendable {
    var name: String?
    var description: String?
    var maxSlots: Int?

    enum CodingKeys: String, CodingKey {
        case name
        case description
        case maxSlots = "max_slots"
    }
}

struct AssignRoleRequest: Codable, Sendable {
    let roleId: String
    let userId: String?
    let note: String?

    enum CodingKeys: String, CodingKey {
        case roleId = "role_id"
        case userId = "user_id"
        case note
    }

    init(roleId: String, userId: String? = nil, note: String? = nil) {
        self.roleId = roleId
        self.userId = userId
        self.note = note
    }
}
