import Foundation

// MARK: - Events API

extension APIClient {

    // MARK: - Event CRUD

    /// Create a new event
    func createEvent(_ request: CreateEventRequest) async throws -> Event {
        let response: DataResponse<Event> = try await post(path: "/events", body: request)
        return response.data
    }

    /// Get event details with RSVPs and roles
    func getEvent(eventId: String) async throws -> EventWithDetails {
        return try await get(path: "/events/\(eventId)")
    }

    /// Update an event
    func updateEvent(eventId: String, _ request: UpdateEventRequest) async throws -> Event {
        let response: DataResponse<Event> = try await patch(path: "/events/\(eventId)", body: request)
        return response.data
    }

    /// Cancel an event
    func cancelEvent(eventId: String) async throws {
        try await delete(path: "/events/\(eventId)")
    }

    // MARK: - Event Discovery

    /// Discover public events
    func discoverEvents(template: String? = nil, city: String? = nil, limit: Int = 20) async throws -> [Event] {
        var queryItems: [URLQueryItem] = [URLQueryItem(name: "limit", value: String(limit))]
        if let template = template { queryItems.append(URLQueryItem(name: "template", value: template)) }
        if let city = city { queryItems.append(URLQueryItem(name: "city", value: city)) }

        let response: CollectionResponse<Event> = try await get(path: "/discover/events", queryItems: queryItems)
        return response.data
    }

    /// Get events for a guild
    func getGuildEvents(guildId: String) async throws -> [Event] {
        let response: CollectionResponse<Event> = try await get(path: "/guilds/\(guildId)/events")
        return response.data
    }

    // MARK: - RSVP

    /// Create or update RSVP to an event
    func createRSVP(eventId: String, status: RSVPStatus, note: String? = nil) async throws -> RSVP {
        let request = RSVPRequest(status: status, note: note)
        let response: DataResponse<RSVP> = try await post(path: "/events/\(eventId)/rsvp", body: request)
        return response.data
    }

    /// Cancel own RSVP
    func cancelRSVP(eventId: String) async throws {
        try await delete(path: "/events/\(eventId)/rsvp")
    }

    /// Get pending RSVPs (host only)
    func getPendingRSVPs(eventId: String) async throws -> [RSVP] {
        let response: CollectionResponse<RSVP> = try await get(path: "/events/\(eventId)/rsvps/pending")
        return response.data
    }

    /// Respond to an RSVP (host only)
    func respondToRSVP(eventId: String, userId: String, accept: Bool, message: String? = nil) async throws -> RSVP {
        let request = RespondToRSVPRequest(accept: accept, message: message)
        let response: DataResponse<RSVP> = try await post(path: "/events/\(eventId)/rsvps/\(userId)/respond", body: request)
        return response.data
    }

    // MARK: - Hosts

    /// Add a co-host to an event
    func addHost(eventId: String, userId: String) async throws -> EventHost {
        let request = AddHostRequest(userId: userId)
        let response: DataResponse<EventHost> = try await post(path: "/events/\(eventId)/hosts", body: request)
        return response.data
    }

    // MARK: - Check-in & Feedback

    /// Check in to an event
    func checkinEvent(eventId: String) async throws {
        try await postNoContent(path: "/events/\(eventId)/checkin", body: EmptyRequest())
    }

    /// Submit event feedback
    func submitEventFeedback(eventId: String, attended: Bool, rating: Int? = nil, comment: String? = nil) async throws {
        let request = EventFeedbackRequest(attended: attended, rating: rating, comment: comment)
        try await postNoContent(path: "/events/\(eventId)/feedback", body: request)
    }

    /// Confirm event completion
    func confirmEventCompletion(eventId: String, completed: Bool) async throws {
        let request = ConfirmCompletionRequest(completed: completed)
        try await postNoContent(path: "/events/\(eventId)/confirm", body: request)
    }

    // MARK: - Event Roles

    /// List roles for an event
    func getEventRoles(eventId: String) async throws -> [EventRole] {
        let response: CollectionResponse<EventRole> = try await get(path: "/events/\(eventId)/roles")
        return response.data
    }

    /// Get roles with assignments overview
    func getEventRolesOverview(eventId: String) async throws -> [EventRoleWithAssignments] {
        let response: EventRolesOverview = try await get(path: "/events/\(eventId)/roles/overview")
        return response.data
    }

    /// Create a new role (host only)
    func createEventRole(eventId: String, _ request: CreateEventRoleRequest) async throws -> EventRole {
        let response: DataResponse<EventRole> = try await post(path: "/events/\(eventId)/roles", body: request)
        return response.data
    }

    /// Update a role (host only)
    func updateEventRole(eventId: String, roleId: String, _ request: UpdateEventRoleRequest) async throws -> EventRole {
        let response: DataResponse<EventRole> = try await patch(path: "/events/\(eventId)/roles/\(roleId)", body: request)
        return response.data
    }

    /// Delete a role (host only)
    func deleteEventRole(eventId: String, roleId: String) async throws {
        try await delete(path: "/events/\(eventId)/roles/\(roleId)")
    }

    /// Assign self to a role
    func assignEventRole(eventId: String, roleId: String, note: String? = nil) async throws -> RoleAssignment {
        let request = AssignRoleRequest(roleId: roleId, note: note)
        let response: DataResponse<RoleAssignment> = try await post(path: "/events/\(eventId)/roles/assign", body: request)
        return response.data
    }

    /// Get my roles for an event
    func getMyEventRoles(eventId: String) async throws -> [RoleAssignment] {
        let response: CollectionResponse<RoleAssignment> = try await get(path: "/events/\(eventId)/roles/mine")
        return response.data
    }

    /// Get role suggestions based on interests
    func getEventRoleSuggestions(eventId: String) async throws -> [RoleSuggestion] {
        let response: CollectionResponse<RoleSuggestion> = try await get(path: "/events/\(eventId)/roles/suggestions")
        return response.data
    }

    /// Cancel own role assignment
    func cancelEventRoleAssignment(eventId: String, assignmentId: String) async throws {
        try await delete(path: "/events/\(eventId)/roles/assignments/\(assignmentId)")
    }
}

// MARK: - Helper Types

private struct EmptyRequest: Codable, Sendable {}
