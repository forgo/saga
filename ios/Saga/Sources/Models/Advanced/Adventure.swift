import Foundation

// MARK: - Adventure

/// An adventure with admission control for group activities
struct Adventure: Codable, Identifiable, Hashable, Sendable {
    let id: String
    let guildId: String
    let hostId: String
    let title: String
    let description: String?
    let location: String?
    let locationLat: Double?
    let locationLng: Double?
    let startTime: Date
    let endTime: Date?
    let capacity: Int?
    let admissionType: AdmissionType
    let admissionCriteria: AdmissionCriteria?
    let status: AdventureStatus
    let participantCount: Int
    let createdOn: Date
    let updatedOn: Date

    enum CodingKeys: String, CodingKey {
        case id
        case guildId = "guild_id"
        case hostId = "host_id"
        case title
        case description
        case location
        case locationLat = "location_lat"
        case locationLng = "location_lng"
        case startTime = "start_time"
        case endTime = "end_time"
        case capacity
        case admissionType = "admission_type"
        case admissionCriteria = "admission_criteria"
        case status
        case participantCount = "participant_count"
        case createdOn = "created_on"
        case updatedOn = "updated_on"
    }

    /// Whether the adventure has available spots
    var hasAvailableSpots: Bool {
        guard let capacity = capacity else { return true }
        return participantCount < capacity
    }

    /// Spots remaining text
    var spotsRemainingText: String? {
        guard let capacity = capacity else { return nil }
        let remaining = capacity - participantCount
        if remaining <= 0 { return "Full" }
        return "\(remaining) spot\(remaining == 1 ? "" : "s") left"
    }
}

// MARK: - Admission Type

enum AdmissionType: String, Codable, Sendable, CaseIterable {
    case open
    case approval
    case invite
    case criteria

    var displayName: String {
        switch self {
        case .open: return "Open"
        case .approval: return "Approval Required"
        case .invite: return "Invite Only"
        case .criteria: return "Criteria-Based"
        }
    }

    var description: String {
        switch self {
        case .open: return "Anyone can join"
        case .approval: return "Host approves each request"
        case .invite: return "By invitation only"
        case .criteria: return "Must meet specific criteria"
        }
    }

    var iconName: String {
        switch self {
        case .open: return "door.left.hand.open"
        case .approval: return "hand.raised.fill"
        case .invite: return "envelope.fill"
        case .criteria: return "checklist"
        }
    }
}

// MARK: - Admission Criteria

/// Criteria for automatic admission to an adventure
struct AdmissionCriteria: Codable, Hashable, Sendable {
    let minTrustLevel: TrustLevel?
    let minCompatibility: Double?
    let requiredInterests: [String]?
    let minIrlConfirmations: Int?
    let guildMemberOnly: Bool

    enum CodingKeys: String, CodingKey {
        case minTrustLevel = "min_trust_level"
        case minCompatibility = "min_compatibility"
        case requiredInterests = "required_interests"
        case minIrlConfirmations = "min_irl_confirmations"
        case guildMemberOnly = "guild_member_only"
    }
}

// MARK: - Adventure Status

enum AdventureStatus: String, Codable, Sendable, CaseIterable {
    case draft
    case open
    case full
    case inProgress = "in_progress"
    case completed
    case cancelled

    var displayName: String {
        switch self {
        case .draft: return "Draft"
        case .open: return "Open"
        case .full: return "Full"
        case .inProgress: return "In Progress"
        case .completed: return "Completed"
        case .cancelled: return "Cancelled"
        }
    }

    var color: String {
        switch self {
        case .draft: return "gray"
        case .open: return "green"
        case .full: return "orange"
        case .inProgress: return "blue"
        case .completed: return "purple"
        case .cancelled: return "red"
        }
    }
}

// MARK: - Adventure Participant

/// A participant in an adventure
struct AdventureParticipant: Codable, Identifiable, Hashable, Sendable {
    let id: String
    let adventureId: String
    let userId: String
    let status: ParticipantStatus
    let role: String?
    let joinedOn: Date
    let approvedOn: Date?

    enum CodingKeys: String, CodingKey {
        case id
        case adventureId = "adventure_id"
        case userId = "user_id"
        case status
        case role
        case joinedOn = "joined_on"
        case approvedOn = "approved_on"
    }
}

// MARK: - Participant Status

enum ParticipantStatus: String, Codable, Sendable, CaseIterable {
    case pending
    case approved
    case rejected
    case withdrawn

    var displayName: String {
        switch self {
        case .pending: return "Pending"
        case .approved: return "Approved"
        case .rejected: return "Rejected"
        case .withdrawn: return "Withdrawn"
        }
    }
}

// MARK: - Adventure With Details

/// Adventure with participant list
struct AdventureWithDetails: Codable, Sendable {
    let adventure: Adventure
    let participants: [AdventureParticipantDisplay]
    let myStatus: ParticipantStatus?

    enum CodingKeys: String, CodingKey {
        case adventure
        case participants
        case myStatus = "my_status"
    }
}

// MARK: - Adventure Participant Display

/// Participant with public profile
struct AdventureParticipantDisplay: Codable, Sendable {
    let participant: AdventureParticipant
    let user: PublicProfile
}

// MARK: - Request Types

struct CreateAdventureRequest: Codable, Sendable {
    let guildId: String
    let title: String
    let description: String?
    let location: String?
    let locationLat: Double?
    let locationLng: Double?
    let startTime: Date
    let endTime: Date?
    let capacity: Int?
    let admissionType: AdmissionType
    let admissionCriteria: AdmissionCriteria?

    enum CodingKeys: String, CodingKey {
        case guildId = "guild_id"
        case title
        case description
        case location
        case locationLat = "location_lat"
        case locationLng = "location_lng"
        case startTime = "start_time"
        case endTime = "end_time"
        case capacity
        case admissionType = "admission_type"
        case admissionCriteria = "admission_criteria"
    }
}

struct UpdateAdventureRequest: Codable, Sendable {
    var title: String?
    var description: String?
    var location: String?
    var locationLat: Double?
    var locationLng: Double?
    var startTime: Date?
    var endTime: Date?
    var capacity: Int?
    var admissionType: AdmissionType?
    var admissionCriteria: AdmissionCriteria?
    var status: AdventureStatus?

    enum CodingKeys: String, CodingKey {
        case title
        case description
        case location
        case locationLat = "location_lat"
        case locationLng = "location_lng"
        case startTime = "start_time"
        case endTime = "end_time"
        case capacity
        case admissionType = "admission_type"
        case admissionCriteria = "admission_criteria"
        case status
    }
}

struct JoinAdventureRequest: Codable, Sendable {
    let message: String?
}

struct RespondToJoinRequest: Codable, Sendable {
    let approve: Bool
    let message: String?
}
