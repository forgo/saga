import Foundation

// MARK: - Trust Grant

/// A trust grant from one user to another
struct TrustGrant: Codable, Identifiable, Hashable, Sendable {
    let id: String
    let grantorId: String
    let granteeId: String
    let trustLevel: TrustLevel
    let permissions: [TrustPermission]?
    let notes: String?
    let createdOn: Date
    let updatedOn: Date

    enum CodingKeys: String, CodingKey {
        case id
        case grantorId = "grantor_id"
        case granteeId = "grantee_id"
        case trustLevel = "trust_level"
        case permissions
        case notes
        case createdOn = "created_on"
        case updatedOn = "updated_on"
    }
}

// MARK: - Trust Level

enum TrustLevel: String, Codable, Sendable, CaseIterable {
    case basic
    case elevated
    case full

    var displayName: String {
        switch self {
        case .basic: return "Basic"
        case .elevated: return "Elevated"
        case .full: return "Full"
        }
    }

    var description: String {
        switch self {
        case .basic: return "Basic trust - limited access"
        case .elevated: return "Elevated trust - more access"
        case .full: return "Full trust - complete access"
        }
    }

    var iconName: String {
        switch self {
        case .basic: return "shield"
        case .elevated: return "shield.lefthalf.filled"
        case .full: return "shield.fill"
        }
    }

    /// Default permissions for this trust level
    var defaultPermissions: [TrustPermission] {
        switch self {
        case .basic: return []
        case .elevated: return [.viewLocation]
        case .full: return [.viewLocation, .viewSchedule, .contactDirectly]
        }
    }
}

// MARK: - Trust Permission

enum TrustPermission: String, Codable, Sendable, CaseIterable {
    case viewLocation = "view_location"
    case viewSchedule = "view_schedule"
    case contactDirectly = "contact_directly"
    case seeCommute = "see_commute"

    var displayName: String {
        switch self {
        case .viewLocation: return "View Location"
        case .viewSchedule: return "View Schedule"
        case .contactDirectly: return "Contact Directly"
        case .seeCommute: return "See Commute"
        }
    }

    var description: String {
        switch self {
        case .viewLocation: return "Can see your approximate location"
        case .viewSchedule: return "Can see your availability schedule"
        case .contactDirectly: return "Can message you directly"
        case .seeCommute: return "Can see your commute patterns"
        }
    }

    var iconName: String {
        switch self {
        case .viewLocation: return "location"
        case .viewSchedule: return "calendar"
        case .contactDirectly: return "message"
        case .seeCommute: return "car"
        }
    }
}

// MARK: - Trust Summary

/// Summary of trust relationships for a user
struct TrustSummary: Codable, Sendable {
    let userId: String
    let trustedByCount: Int
    let trustsCount: Int
    let mutualTrustCount: Int
    let irlConfirmedCount: Int

    enum CodingKeys: String, CodingKey {
        case userId = "user_id"
        case trustedByCount = "trusted_by_count"
        case trustsCount = "trusts_count"
        case mutualTrustCount = "mutual_trust_count"
        case irlConfirmedCount = "irl_confirmed_count"
    }
}

// MARK: - IRL Confirmation

/// A request to confirm meeting someone in real life
struct IRLConfirmation: Codable, Identifiable, Hashable, Sendable {
    let id: String
    let requesterId: String
    let targetId: String
    let context: IRLContext
    let contextId: String?
    let location: String?
    let status: IRLConfirmationStatus
    let createdOn: Date
    let confirmedOn: Date?

    enum CodingKeys: String, CodingKey {
        case id
        case requesterId = "requester_id"
        case targetId = "target_id"
        case context
        case contextId = "context_id"
        case location
        case status
        case createdOn = "created_on"
        case confirmedOn = "confirmed_on"
    }
}

// MARK: - IRL Context

enum IRLContext: String, Codable, Sendable, CaseIterable {
    case event
    case hangout
    case spontaneous

    var displayName: String {
        switch self {
        case .event: return "Event"
        case .hangout: return "Hangout"
        case .spontaneous: return "Spontaneous"
        }
    }
}

// MARK: - IRL Confirmation Status

enum IRLConfirmationStatus: String, Codable, Sendable, CaseIterable {
    case pending
    case confirmed
    case rejected
    case expired

    var displayName: String {
        switch self {
        case .pending: return "Pending"
        case .confirmed: return "Confirmed"
        case .rejected: return "Rejected"
        case .expired: return "Expired"
        }
    }
}

// MARK: - Trust Rating

/// A trust/distrust rating given after an interaction
struct TrustRating: Codable, Identifiable, Hashable, Sendable {
    let id: String
    let raterId: String
    let rateeId: String
    let anchorType: TrustAnchorType
    let anchorId: String
    let trustLevel: TrustRatingLevel
    let trustReview: String
    let reviewVisibility: ReviewVisibility?
    let createdOn: Date
    let updatedOn: Date?

    enum CodingKeys: String, CodingKey {
        case id
        case raterId = "rater_id"
        case rateeId = "ratee_id"
        case anchorType = "anchor_type"
        case anchorId = "anchor_id"
        case trustLevel = "trust_level"
        case trustReview = "trust_review"
        case reviewVisibility = "review_visibility"
        case createdOn = "created_on"
        case updatedOn = "updated_on"
    }
}

// MARK: - Trust Anchor Type

enum TrustAnchorType: String, Codable, Sendable, CaseIterable {
    case event
    case rideshare

    var displayName: String {
        switch self {
        case .event: return "Event"
        case .rideshare: return "Rideshare"
        }
    }
}

// MARK: - Trust Rating Level

enum TrustRatingLevel: String, Codable, Sendable, CaseIterable {
    case trust
    case distrust

    var displayName: String {
        switch self {
        case .trust: return "Trust"
        case .distrust: return "Distrust"
        }
    }

    var iconName: String {
        switch self {
        case .trust: return "hand.thumbsup.fill"
        case .distrust: return "hand.thumbsdown.fill"
        }
    }
}

// MARK: - Review Visibility

enum ReviewVisibility: String, Codable, Sendable, CaseIterable {
    case `public`
    case `private`

    var displayName: String {
        switch self {
        case .public: return "Public"
        case .private: return "Private"
        }
    }
}

// MARK: - Trust Rating With Endorsements

/// A trust rating with endorsement counts
struct TrustRatingWithEndorsements: Codable, Sendable {
    let id: String
    let raterId: String
    let rateeId: String
    let anchorType: TrustAnchorType
    let anchorId: String
    let trustLevel: TrustRatingLevel
    let trustReview: String
    let agreeCount: Int
    let disagreeCount: Int
    let cooldownUntil: Date?
    let createdOn: Date

    enum CodingKeys: String, CodingKey {
        case id
        case raterId = "rater_id"
        case rateeId = "ratee_id"
        case anchorType = "anchor_type"
        case anchorId = "anchor_id"
        case trustLevel = "trust_level"
        case trustReview = "trust_review"
        case agreeCount = "agree_count"
        case disagreeCount = "disagree_count"
        case cooldownUntil = "cooldown_until"
        case createdOn = "created_on"
    }
}

// MARK: - Trust Endorsement

/// An endorsement (agree/disagree) on a trust rating
struct TrustEndorsement: Codable, Identifiable, Hashable, Sendable {
    let id: String
    let trustRatingId: String
    let endorserId: String
    let endorsementType: EndorsementType
    let note: String?
    let createdOn: Date

    enum CodingKeys: String, CodingKey {
        case id
        case trustRatingId = "trust_rating_id"
        case endorserId = "endorser_id"
        case endorsementType = "endorsement_type"
        case note
        case createdOn = "created_on"
    }
}

// MARK: - Endorsement Type

enum EndorsementType: String, Codable, Sendable, CaseIterable {
    case agree
    case disagree

    var displayName: String {
        switch self {
        case .agree: return "Agree"
        case .disagree: return "Disagree"
        }
    }
}

// MARK: - Trust Aggregate

/// Aggregated trust statistics for a user
struct TrustAggregate: Codable, Sendable {
    let userId: String
    let trustCount: Int
    let distrustCount: Int
    let totalEndorsements: Int

    enum CodingKeys: String, CodingKey {
        case userId = "user_id"
        case trustCount = "trust_count"
        case distrustCount = "distrust_count"
        case totalEndorsements = "total_endorsements"
    }
}

// MARK: - Request Types

struct CreateTrustRequest: Codable, Sendable {
    let granteeId: String
    let trustLevel: TrustLevel
    let permissions: [TrustPermission]?
    let notes: String?

    enum CodingKeys: String, CodingKey {
        case granteeId = "grantee_id"
        case trustLevel = "trust_level"
        case permissions
        case notes
    }
}

struct UpdateTrustRequest: Codable, Sendable {
    var trustLevel: TrustLevel?
    var permissions: [TrustPermission]?
    var notes: String?

    enum CodingKeys: String, CodingKey {
        case trustLevel = "trust_level"
        case permissions
        case notes
    }
}

struct RequestIRLRequest: Codable, Sendable {
    let targetId: String
    let context: IRLContext
    let contextId: String?
    let location: String?

    enum CodingKeys: String, CodingKey {
        case targetId = "target_id"
        case context
        case contextId = "context_id"
        case location
    }
}

struct AcceptIRLRequest: Codable, Sendable {
    let confirm: Bool
}

struct CreateTrustRatingRequest: Codable, Sendable {
    let rateeId: String
    let anchorType: TrustAnchorType
    let anchorId: String
    let trustLevel: TrustRatingLevel
    let trustReview: String

    enum CodingKeys: String, CodingKey {
        case rateeId = "ratee_id"
        case anchorType = "anchor_type"
        case anchorId = "anchor_id"
        case trustLevel = "trust_level"
        case trustReview = "trust_review"
    }
}

struct CreateEndorsementRequest: Codable, Sendable {
    let endorsementType: EndorsementType
    let note: String?

    enum CodingKeys: String, CodingKey {
        case endorsementType = "endorsement_type"
        case note
    }
}
