import Foundation

/// Service for managing events and RSVPs
@Observable
final class EventService: @unchecked Sendable {
    static let shared = EventService()

    private let apiClient = APIClient.shared

    // MARK: - State

    /// Events for the current guild
    private(set) var guildEvents: [Event] = []

    /// Discovered public events
    private(set) var discoveredEvents: [Event] = []

    /// Currently selected event with details
    private(set) var currentEventDetails: EventWithDetails?

    /// Loading state
    private(set) var isLoading = false

    /// Error message
    private(set) var errorMessage: String?

    private init() {}

    // MARK: - Guild Events

    /// Load events for a guild
    @MainActor
    func loadGuildEvents(guildId: String) async {
        isLoading = true
        errorMessage = nil

        do {
            let events = try await apiClient.getGuildEvents(guildId: guildId)
            guildEvents = events.sorted { $0.startTime < $1.startTime }
        } catch {
            errorMessage = error.localizedDescription
        }

        isLoading = false
    }

    /// Create a new event
    @MainActor
    func createEvent(
        guildId: String,
        title: String,
        description: String? = nil,
        location: String? = nil,
        locationLat: Double? = nil,
        locationLng: Double? = nil,
        startTime: Date,
        endTime: Date? = nil,
        capacity: Int? = nil,
        visibility: EventVisibility = .guild
    ) async throws -> Event {
        let request = CreateEventRequest(
            guildId: guildId,
            title: title,
            description: description,
            location: location,
            locationLat: locationLat,
            locationLng: locationLng,
            startTime: startTime,
            endTime: endTime,
            capacity: capacity,
            visibility: visibility
        )

        let event = try await apiClient.createEvent(request)
        guildEvents.insert(event, at: 0)
        guildEvents.sort { $0.startTime < $1.startTime }
        return event
    }

    /// Update an event
    @MainActor
    func updateEvent(eventId: String, _ request: UpdateEventRequest) async throws -> Event {
        let event = try await apiClient.updateEvent(eventId: eventId, request)

        if let index = guildEvents.firstIndex(where: { $0.id == eventId }) {
            guildEvents[index] = event
        }

        if currentEventDetails?.event.id == eventId {
            currentEventDetails = EventWithDetails(
                event: event,
                rsvps: currentEventDetails?.rsvps ?? [],
                roles: currentEventDetails?.roles ?? [],
                myRsvp: currentEventDetails?.myRsvp,
                canManage: currentEventDetails?.canManage ?? false
            )
        }

        return event
    }

    /// Cancel an event
    @MainActor
    func cancelEvent(eventId: String) async throws {
        try await apiClient.cancelEvent(eventId: eventId)
        guildEvents.removeAll { $0.id == eventId }

        if currentEventDetails?.event.id == eventId {
            currentEventDetails = nil
        }
    }

    // MARK: - Event Details

    /// Load event details
    @MainActor
    func loadEventDetails(eventId: String) async throws {
        isLoading = true
        errorMessage = nil

        do {
            currentEventDetails = try await apiClient.getEvent(eventId: eventId)
        } catch {
            errorMessage = error.localizedDescription
            throw error
        }

        isLoading = false
    }

    /// Clear current event details
    @MainActor
    func clearCurrentEvent() {
        currentEventDetails = nil
    }

    // MARK: - Discovery

    /// Discover public events
    @MainActor
    func discoverEvents(template: String? = nil, city: String? = nil) async {
        isLoading = true
        errorMessage = nil

        do {
            discoveredEvents = try await apiClient.discoverEvents(template: template, city: city)
        } catch {
            errorMessage = error.localizedDescription
        }

        isLoading = false
    }

    // MARK: - RSVP

    /// RSVP to an event
    @MainActor
    func rsvp(eventId: String, status: RSVPStatus, note: String? = nil) async throws -> RSVP {
        let rsvp = try await apiClient.createRSVP(eventId: eventId, status: status, note: note)

        // Update current event details if viewing this event
        if var details = currentEventDetails, details.event.id == eventId {
            // Update my RSVP
            var rsvps = details.rsvps
            if let index = rsvps.firstIndex(where: { $0.userId == rsvp.userId }) {
                rsvps[index] = rsvp
            } else {
                rsvps.append(rsvp)
            }

            currentEventDetails = EventWithDetails(
                event: details.event,
                rsvps: rsvps,
                roles: details.roles,
                myRsvp: rsvp,
                canManage: details.canManage
            )
        }

        // Update event in guild list
        if let index = guildEvents.firstIndex(where: { $0.id == eventId }) {
            // Reload to get updated rsvp_count
            await loadGuildEvents(guildId: guildEvents[index].guildId)
        }

        return rsvp
    }

    /// Cancel RSVP
    @MainActor
    func cancelRSVP(eventId: String) async throws {
        try await apiClient.cancelRSVP(eventId: eventId)

        // Update current event details
        if var details = currentEventDetails, details.event.id == eventId {
            let rsvps = details.rsvps.filter { $0.id != details.myRsvp?.id }

            currentEventDetails = EventWithDetails(
                event: details.event,
                rsvps: rsvps,
                roles: details.roles,
                myRsvp: nil,
                canManage: details.canManage
            )
        }
    }

    // MARK: - Check-in & Feedback

    /// Check in to an event
    @MainActor
    func checkin(eventId: String) async throws {
        try await apiClient.checkinEvent(eventId: eventId)
    }

    /// Submit feedback
    @MainActor
    func submitFeedback(eventId: String, attended: Bool, rating: Int? = nil, comment: String? = nil) async throws {
        try await apiClient.submitEventFeedback(eventId: eventId, attended: attended, rating: rating, comment: comment)
    }

    // MARK: - Roles

    /// Create a role for an event
    @MainActor
    func createRole(eventId: String, name: String, description: String? = nil, maxSlots: Int = 1) async throws -> EventRole {
        let request = CreateEventRoleRequest(name: name, description: description, maxSlots: maxSlots)
        let role = try await apiClient.createEventRole(eventId: eventId, request)

        // Update current event details
        if var details = currentEventDetails, details.event.id == eventId {
            var roles = details.roles
            roles.append(role)

            currentEventDetails = EventWithDetails(
                event: details.event,
                rsvps: details.rsvps,
                roles: roles,
                myRsvp: details.myRsvp,
                canManage: details.canManage
            )
        }

        return role
    }

    /// Assign self to a role
    @MainActor
    func assignRole(eventId: String, roleId: String, note: String? = nil) async throws -> RoleAssignment {
        return try await apiClient.assignEventRole(eventId: eventId, roleId: roleId, note: note)
    }

    /// Get role suggestions
    @MainActor
    func getRoleSuggestions(eventId: String) async throws -> [RoleSuggestion] {
        return try await apiClient.getEventRoleSuggestions(eventId: eventId)
    }

    // MARK: - Helpers

    /// Filter events by upcoming (future events)
    var upcomingEvents: [Event] {
        guildEvents.filter { !$0.isPast && $0.status == .published }
    }

    /// Filter events by past
    var pastEvents: [Event] {
        guildEvents.filter { $0.isPast || $0.status == .completed }
    }

    /// Events happening today
    var todayEvents: [Event] {
        guildEvents.filter { $0.isToday && $0.status == .published }
    }

    /// Clear all state
    @MainActor
    func clear() {
        guildEvents = []
        discoveredEvents = []
        currentEventDetails = nil
        errorMessage = nil
    }
}
