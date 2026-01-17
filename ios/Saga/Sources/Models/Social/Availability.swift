import Foundation

// MARK: - Availability

/// A user's posted availability for hangouts
struct Availability: Codable, Identifiable, Hashable, Sendable {
    let id: String
    let userId: String
    let hangoutType: HangoutType
    let title: String?
    let description: String?
    let startTime: Date
    let endTime: Date
    let location: String?
    let locationLat: Double?
    let locationLng: Double?
    let radiusKm: Double?
    let status: AvailabilityStatus
    let createdOn: Date

    enum CodingKeys: String, CodingKey {
        case id
        case userId = "user_id"
        case hangoutType = "hangout_type"
        case title
        case description
        case startTime = "start_time"
        case endTime = "end_time"
        case location
        case locationLat = "location_lat"
        case locationLng = "location_lng"
        case radiusKm = "radius_km"
        case status
        case createdOn = "created_on"
    }

    /// Whether the availability is currently active
    var isActive: Bool {
        status == .active && endTime > Date()
    }

    /// Time range display
    var timeRange: String {
        let formatter = DateFormatter()
        formatter.dateStyle = .short
        formatter.timeStyle = .short
        return "\(formatter.string(from: startTime)) - \(formatter.string(from: endTime))"
    }
}

// MARK: - Hangout Type

enum HangoutType: String, Codable, Sendable, CaseIterable {
    case talkItOut = "talk_it_out"
    case hereToListen = "here_to_listen"
    case concreteActivity = "concrete_activity"
    case mutualInterest = "mutual_interest"
    case meetAnyone = "meet_anyone"

    var displayName: String {
        switch self {
        case .talkItOut: return "Talk It Out"
        case .hereToListen: return "Here to Listen"
        case .concreteActivity: return "Concrete Activity"
        case .mutualInterest: return "Mutual Interest"
        case .meetAnyone: return "Meet Anyone"
        }
    }

    var description: String {
        switch self {
        case .talkItOut: return "I want to talk about something on my mind"
        case .hereToListen: return "I'm available to listen and support"
        case .concreteActivity: return "I have a specific activity in mind"
        case .mutualInterest: return "Looking to connect over shared interests"
        case .meetAnyone: return "Open to meeting new people"
        }
    }

    var iconName: String {
        switch self {
        case .talkItOut: return "bubble.left.and.bubble.right"
        case .hereToListen: return "ear"
        case .concreteActivity: return "figure.walk"
        case .mutualInterest: return "heart.circle"
        case .meetAnyone: return "person.wave.2"
        }
    }
}

// MARK: - Availability Status

enum AvailabilityStatus: String, Codable, Sendable, CaseIterable {
    case active
    case matched
    case expired
    case cancelled

    var displayName: String {
        switch self {
        case .active: return "Active"
        case .matched: return "Matched"
        case .expired: return "Expired"
        case .cancelled: return "Cancelled"
        }
    }
}

// MARK: - Nearby Availability

/// An availability found nearby with user info
struct NearbyAvailability: Codable, Sendable {
    let availability: Availability
    let user: PublicProfile
    let distanceKm: Double
    let compatibilityScore: Double?

    enum CodingKeys: String, CodingKey {
        case availability
        case user
        case distanceKm = "distance_km"
        case compatibilityScore = "compatibility_score"
    }
}

// MARK: - Hangout Request

/// A request to hang out based on an availability posting
struct HangoutRequest: Codable, Identifiable, Hashable, Sendable {
    let id: String
    let availabilityId: String
    let requesterId: String
    let message: String?
    let status: HangoutRequestStatus
    let createdOn: Date
    let respondedOn: Date?

    enum CodingKeys: String, CodingKey {
        case id
        case availabilityId = "availability_id"
        case requesterId = "requester_id"
        case message
        case status
        case createdOn = "created_on"
        case respondedOn = "responded_on"
    }
}

// MARK: - Hangout Request Status

enum HangoutRequestStatus: String, Codable, Sendable, CaseIterable {
    case pending
    case accepted
    case declined
    case cancelled

    var displayName: String {
        switch self {
        case .pending: return "Pending"
        case .accepted: return "Accepted"
        case .declined: return "Declined"
        case .cancelled: return "Cancelled"
        }
    }

    var iconName: String {
        switch self {
        case .pending: return "clock"
        case .accepted: return "checkmark.circle.fill"
        case .declined: return "xmark.circle.fill"
        case .cancelled: return "minus.circle.fill"
        }
    }
}

// MARK: - Hangout Request Display

/// A hangout request with full user and availability info
struct HangoutRequestDisplay: Codable, Sendable {
    let request: HangoutRequest
    let requester: PublicProfile
    let availability: Availability
}

// MARK: - Hangout

/// A confirmed hangout between users
struct Hangout: Codable, Identifiable, Hashable, Sendable {
    let id: String
    let availabilityId: String
    let participants: [String]
    let scheduledTime: Date?
    let location: String?
    let status: HangoutStatus
    let createdOn: Date

    enum CodingKeys: String, CodingKey {
        case id
        case availabilityId = "availability_id"
        case participants
        case scheduledTime = "scheduled_time"
        case location
        case status
        case createdOn = "created_on"
    }
}

// MARK: - Hangout Status

enum HangoutStatus: String, Codable, Sendable, CaseIterable {
    case scheduled
    case completed
    case cancelled
    case noShow = "no_show"

    var displayName: String {
        switch self {
        case .scheduled: return "Scheduled"
        case .completed: return "Completed"
        case .cancelled: return "Cancelled"
        case .noShow: return "No Show"
        }
    }
}

// MARK: - Request Types

struct CreateAvailabilityRequest: Codable, Sendable {
    let hangoutType: HangoutType
    let title: String?
    let description: String?
    let startTime: Date
    let endTime: Date
    let location: String?
    let locationLat: Double?
    let locationLng: Double?
    let radiusKm: Double?

    enum CodingKeys: String, CodingKey {
        case hangoutType = "hangout_type"
        case title
        case description
        case startTime = "start_time"
        case endTime = "end_time"
        case location
        case locationLat = "location_lat"
        case locationLng = "location_lng"
        case radiusKm = "radius_km"
    }
}

struct UpdateAvailabilityRequest: Codable, Sendable {
    var title: String?
    var description: String?
    var startTime: Date?
    var endTime: Date?
    var location: String?
    var locationLat: Double?
    var locationLng: Double?
    var radiusKm: Double?
    var status: AvailabilityStatus?

    enum CodingKeys: String, CodingKey {
        case title
        case description
        case startTime = "start_time"
        case endTime = "end_time"
        case location
        case locationLat = "location_lat"
        case locationLng = "location_lng"
        case radiusKm = "radius_km"
        case status
    }
}

struct CreateHangoutRequestRequest: Codable, Sendable {
    let message: String?
}

struct RespondHangoutRequest: Codable, Sendable {
    let accept: Bool
    let message: String?
}
