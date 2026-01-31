import Foundation

// MARK: - Event

/// An event within a guild
struct Event: Codable, Identifiable, Hashable, Sendable {
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
    let rsvpCount: Int
    let status: EventStatus
    let visibility: EventVisibility
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
        case rsvpCount = "rsvp_count"
        case status
        case visibility
        case createdOn = "created_on"
        case updatedOn = "updated_on"
    }

    /// Whether the event has a location with coordinates
    var hasCoordinates: Bool {
        locationLat != nil && locationLng != nil
    }

    /// Whether the event is in the past
    var isPast: Bool {
        startTime < Date()
    }

    /// Whether the event is happening today
    var isToday: Bool {
        Calendar.current.isDateInToday(startTime)
    }

    /// Whether the event is full (capacity reached)
    var isFull: Bool {
        guard let capacity = capacity else { return false }
        return rsvpCount >= capacity
    }

    /// Spots remaining (nil if unlimited)
    var spotsRemaining: Int? {
        guard let capacity = capacity else { return nil }
        return max(0, capacity - rsvpCount)
    }

    /// Formatted time range
    var timeRange: String {
        let formatter = DateFormatter()
        formatter.dateStyle = .none
        formatter.timeStyle = .short

        let start = formatter.string(from: startTime)
        if let end = endTime {
            let endStr = formatter.string(from: end)
            return "\(start) - \(endStr)"
        }
        return start
    }

    /// Formatted date
    var formattedDate: String {
        let formatter = DateFormatter()
        formatter.dateStyle = .medium
        formatter.timeStyle = .none
        return formatter.string(from: startTime)
    }
}

// MARK: - Event Status

enum EventStatus: String, Codable, Sendable, CaseIterable {
    case draft
    case published
    case cancelled
    case completed

    var displayName: String {
        switch self {
        case .draft: return "Draft"
        case .published: return "Published"
        case .cancelled: return "Cancelled"
        case .completed: return "Completed"
        }
    }

    var iconName: String {
        switch self {
        case .draft: return "doc.badge.ellipsis"
        case .published: return "checkmark.circle"
        case .cancelled: return "xmark.circle"
        case .completed: return "checkmark.seal"
        }
    }
}

// MARK: - Event Visibility

enum EventVisibility: String, Codable, Sendable, CaseIterable {
    case `public`
    case guild
    case `private`

    var displayName: String {
        switch self {
        case .public: return "Public"
        case .guild: return "Guild Only"
        case .private: return "Private"
        }
    }

    var iconName: String {
        switch self {
        case .public: return "globe"
        case .guild: return "person.3"
        case .private: return "lock"
        }
    }
}

// MARK: - RSVP

/// A user's RSVP to an event
struct RSVP: Codable, Identifiable, Hashable, Sendable {
    let id: String
    let eventId: String
    let userId: String
    let status: RSVPStatus
    let note: String?
    let createdOn: Date
    let updatedOn: Date

    enum CodingKeys: String, CodingKey {
        case id
        case eventId = "event_id"
        case userId = "user_id"
        case status
        case note
        case createdOn = "created_on"
        case updatedOn = "updated_on"
    }
}

// MARK: - RSVP Status

enum RSVPStatus: String, Codable, Sendable, CaseIterable {
    case going
    case maybe
    case notGoing = "not_going"

    var displayName: String {
        switch self {
        case .going: return "Going"
        case .maybe: return "Maybe"
        case .notGoing: return "Not Going"
        }
    }

    var iconName: String {
        switch self {
        case .going: return "checkmark.circle.fill"
        case .maybe: return "questionmark.circle.fill"
        case .notGoing: return "xmark.circle.fill"
        }
    }

    var color: String {
        switch self {
        case .going: return "green"
        case .maybe: return "orange"
        case .notGoing: return "red"
        }
    }
}

// MARK: - Event Host

/// A host or co-host of an event
struct EventHost: Codable, Identifiable, Hashable, Sendable {
    let id: String
    let eventId: String
    let userId: String
    let createdOn: Date

    enum CodingKeys: String, CodingKey {
        case id
        case eventId = "event_id"
        case userId = "user_id"
        case createdOn = "created_on"
    }
}

// MARK: - Event With Details

/// Event with RSVPs, roles, and permissions
struct EventWithDetails: Codable, Sendable {
    let event: Event
    let rsvps: [RSVP]
    let roles: [EventRole]
    let myRsvp: RSVP?
    let canManage: Bool

    enum CodingKeys: String, CodingKey {
        case event
        case rsvps
        case roles
        case myRsvp = "my_rsvp"
        case canManage = "can_manage"
    }

    /// Count of "going" RSVPs
    var goingCount: Int {
        rsvps.filter { $0.status == .going }.count
    }

    /// Count of "maybe" RSVPs
    var maybeCount: Int {
        rsvps.filter { $0.status == .maybe }.count
    }

    /// RSVPs with "going" status
    var goingRsvps: [RSVP] {
        rsvps.filter { $0.status == .going }
    }

    /// RSVPs with "maybe" status
    var maybeRsvps: [RSVP] {
        rsvps.filter { $0.status == .maybe }
    }
}

// MARK: - Request Types

struct CreateEventRequest: Codable, Sendable {
    let guildId: String
    let title: String
    let description: String?
    let location: String?
    let locationLat: Double?
    let locationLng: Double?
    let startTime: Date
    let endTime: Date?
    let capacity: Int?
    let visibility: EventVisibility

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
        case visibility
    }
}

struct UpdateEventRequest: Codable, Sendable {
    var title: String?
    var description: String?
    var location: String?
    var locationLat: Double?
    var locationLng: Double?
    var startTime: Date?
    var endTime: Date?
    var capacity: Int?
    var status: EventStatus?
    var visibility: EventVisibility?

    enum CodingKeys: String, CodingKey {
        case title
        case description
        case location
        case locationLat = "location_lat"
        case locationLng = "location_lng"
        case startTime = "start_time"
        case endTime = "end_time"
        case capacity
        case status
        case visibility
    }
}

struct RSVPRequest: Codable, Sendable {
    let status: RSVPStatus
    let note: String?
}

struct RespondToRSVPRequest: Codable, Sendable {
    let accept: Bool
    let message: String?
}

struct EventFeedbackRequest: Codable, Sendable {
    let attended: Bool
    let rating: Int?
    let comment: String?
}

struct ConfirmCompletionRequest: Codable, Sendable {
    let completed: Bool
}

struct AddHostRequest: Codable, Sendable {
    let userId: String

    enum CodingKeys: String, CodingKey {
        case userId = "user_id"
    }
}
