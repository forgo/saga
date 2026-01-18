import Foundation

/// Manages guild data and real-time updates
@Observable
@MainActor
final class GuildService {
    static let shared = GuildService()

    private(set) var guilds: [Guild] = []
    private(set) var currentGuild: GuildData?
    private(set) var isLoading = false
    private(set) var isConnected = false
    private(set) var error: Error?

    private let apiClient = APIClient.shared
    private var sseClient: SSEClient?
    private var currentGuildId: String?

    init() {}

    // MARK: - Guild List

    /// Fetch all guilds for current user
    func fetchGuilds() async throws {
        isLoading = true
        defer { isLoading = false }

        let response = try await apiClient.listGuilds()
        self.guilds = response.data
        self.error = nil
    }

    /// Create a new guild
    func createGuild(name: String, description: String? = nil, icon: String? = nil, color: String? = nil) async throws -> Guild {
        isLoading = true
        defer { isLoading = false }

        let request = CreateGuildRequest(name: name, description: description, icon: icon, color: color)
        let response = try await apiClient.createGuild(request)

        // Use full assignment instead of append for more reliable @Observable updates
        var updatedGuilds = self.guilds
        updatedGuilds.append(response.data)
        self.guilds = updatedGuilds

        return response.data
    }

    /// Delete a guild
    func deleteGuild(id: String) async throws {
        try await apiClient.deleteGuild(id: id)
        self.guilds.removeAll { $0.id == id }
        if self.currentGuildId == id {
            self.currentGuild = nil
            self.currentGuildId = nil
        }
    }

    // MARK: - Guild Detail

    /// Fetch full guild data and connect to SSE
    func selectGuild(id: String) async throws {
        // Disconnect from previous guild
        await disconnectSSE()

        isLoading = true
        defer { isLoading = false }

        let response = try await apiClient.getGuild(id: id)
        self.currentGuild = response.data
        self.currentGuildId = id
        self.error = nil

        // Connect to SSE for real-time updates
        await connectSSE(guildId: id)
    }

    /// Update current guild
    func updateGuild(name: String? = nil, description: String? = nil, icon: String? = nil, color: String? = nil) async throws {
        guard let guildId = currentGuildId else { return }

        let request = UpdateGuildRequest(name: name, description: description, icon: icon, color: color)
        let response = try await apiClient.updateGuild(id: guildId, request)

        // Update in guilds list
        if let index = self.guilds.firstIndex(where: { $0.id == guildId }) {
            self.guilds[index] = response.data
        }
        // Update current guild
        if let current = self.currentGuild {
            self.currentGuild = GuildData(
                guild: response.data,
                members: current.members,
                people: current.people,
                activities: current.activities
            )
        }
    }

    /// Join a guild
    func joinGuild(id: String) async throws {
        try await apiClient.joinGuild(id: id)
        try await fetchGuilds()
    }

    /// Leave current guild
    func leaveGuild() async throws {
        guard let guildId = currentGuildId else { return }

        try await apiClient.leaveGuild(id: guildId)
        await disconnectSSE()

        self.guilds.removeAll { $0.id == guildId }
        self.currentGuild = nil
        self.currentGuildId = nil
    }

    // MARK: - People

    /// Add a person to current guild
    func createPerson(name: String, nickname: String? = nil, birthday: String? = nil, notes: String? = nil) async throws -> Person {
        guard let guildId = currentGuildId else {
            throw APIError.notFound
        }

        let request = CreatePersonRequest(name: name, nickname: nickname, birthday: birthday, notes: notes)
        let response = try await apiClient.createPerson(guildId: guildId, request)

        if let current = self.currentGuild {
            var people = current.people
            people.append(response.data)
            self.currentGuild = GuildData(
                guild: current.guild,
                members: current.members,
                people: people,
                activities: current.activities
            )
        }

        return response.data
    }

    /// Update a person
    func updatePerson(id: String, name: String? = nil, nickname: String? = nil, birthday: String? = nil, notes: String? = nil) async throws {
        guard let guildId = currentGuildId else { return }

        let request = UpdatePersonRequest(name: name, nickname: nickname, birthday: birthday, notes: notes)
        let response = try await apiClient.updatePerson(guildId: guildId, personId: id, request)

        if let current = self.currentGuild {
            var people = current.people
            if let index = people.firstIndex(where: { $0.id == id }) {
                people[index] = response.data
            }
            self.currentGuild = GuildData(
                guild: current.guild,
                members: current.members,
                people: people,
                activities: current.activities
            )
        }
    }

    /// Delete a person
    func deletePerson(id: String) async throws {
        guard let guildId = currentGuildId else { return }

        try await apiClient.deletePerson(guildId: guildId, personId: id)

        if let current = self.currentGuild {
            let people = current.people.filter { $0.id != id }
            self.currentGuild = GuildData(
                guild: current.guild,
                members: current.members,
                people: people,
                activities: current.activities
            )
        }
    }

    // MARK: - Activities

    /// Create activity in current guild
    func createActivity(name: String, icon: String, warn: Int, critical: Int) async throws -> Activity {
        guard let guildId = currentGuildId else {
            throw APIError.notFound
        }

        let request = CreateActivityRequest(name: name, icon: icon, warn: warn, critical: critical)
        let response = try await apiClient.createActivity(guildId: guildId, request)

        if let current = self.currentGuild {
            var activities = current.activities
            activities.append(response.data)
            self.currentGuild = GuildData(
                guild: current.guild,
                members: current.members,
                people: current.people,
                activities: activities
            )
        }

        return response.data
    }

    /// Delete activity
    func deleteActivity(id: String) async throws {
        guard let guildId = currentGuildId else { return }

        try await apiClient.deleteActivity(guildId: guildId, activityId: id)

        if let current = self.currentGuild {
            let activities = current.activities.filter { $0.id != id }
            self.currentGuild = GuildData(
                guild: current.guild,
                members: current.members,
                people: current.people,
                activities: activities
            )
        }
    }

    // MARK: - Timers

    /// Create timer for person
    func createTimer(personId: String, activityId: String, push: Bool = false) async throws -> ActivityTimer {
        guard let guildId = currentGuildId else {
            throw APIError.notFound
        }

        let request = CreateTimerRequest(activityId: activityId, enabled: true, push: push)
        let response = try await apiClient.createTimer(guildId: guildId, personId: personId, request)
        return response.data
    }

    /// Reset timer
    func resetTimer(personId: String, timerId: String) async throws -> ActivityTimer {
        guard let guildId = currentGuildId else {
            throw APIError.notFound
        }

        let response = try await apiClient.resetTimer(guildId: guildId, personId: personId, timerId: timerId)
        return response.data
    }

    /// Delete timer
    func deleteTimer(personId: String, timerId: String) async throws {
        guard let guildId = currentGuildId else { return }

        try await apiClient.deleteTimer(guildId: guildId, personId: personId, timerId: timerId)
    }

    // MARK: - SSE Connection

    private func connectSSE(guildId: String) async {
        guard let token = await apiClient.getAccessToken() else { return }

        sseClient = SSEClient()
        await sseClient?.connect(
            guildId: guildId,
            accessToken: token,
            onEvent: { [weak self] event in
                self?.handleSSEEvent(event)
            },
            onConnectionChange: { [weak self] connected in
                self?.isConnected = connected
            }
        )
    }

    private func disconnectSSE() async {
        await sseClient?.disconnect()
        sseClient = nil
        self.isConnected = false
    }

    @MainActor
    private func handleSSEEvent(_ event: GuildEvent) {
        guard var current = currentGuild else { return }

        switch event {
        case .heartbeat:
            break // Keep-alive, no action needed

        case .personCreated(let person):
            var people = current.people
            if !people.contains(where: { $0.id == person.id }) {
                people.append(person)
                currentGuild = GuildData(guild: current.guild, members: current.members, people: people, activities: current.activities)
            }

        case .personUpdated(let person):
            var people = current.people
            if let index = people.firstIndex(where: { $0.id == person.id }) {
                people[index] = person
                currentGuild = GuildData(guild: current.guild, members: current.members, people: people, activities: current.activities)
            }

        case .personDeleted(let id):
            let people = current.people.filter { $0.id != id }
            currentGuild = GuildData(guild: current.guild, members: current.members, people: people, activities: current.activities)

        case .activityCreated(let activity):
            var activities = current.activities
            if !activities.contains(where: { $0.id == activity.id }) {
                activities.append(activity)
                currentGuild = GuildData(guild: current.guild, members: current.members, people: current.people, activities: activities)
            }

        case .activityUpdated(let activity):
            var activities = current.activities
            if let index = activities.firstIndex(where: { $0.id == activity.id }) {
                activities[index] = activity
                currentGuild = GuildData(guild: current.guild, members: current.members, people: current.people, activities: activities)
            }

        case .activityDeleted(let id):
            let activities = current.activities.filter { $0.id != id }
            currentGuild = GuildData(guild: current.guild, members: current.members, people: current.people, activities: activities)

        case .memberJoined(let member):
            var members = current.members
            if !members.contains(where: { $0.id == member.id }) {
                members.append(member)
                currentGuild = GuildData(guild: current.guild, members: members, people: current.people, activities: current.activities)
            }

        case .memberLeft(let id):
            let members = current.members.filter { $0.id != id }
            currentGuild = GuildData(guild: current.guild, members: members, people: current.people, activities: current.activities)

        case .timerCreated, .timerReset, .timerUpdated, .timerDeleted, .timerWarn, .timerCritical:
            // Timer events are handled at the person detail level
            // Post a notification or use a callback for views that need this
            NotificationCenter.default.post(name: .guildTimerEvent, object: event)
        }
    }

    // MARK: - Helpers

    /// Clear all state (for logout)
    func clear() async {
        await disconnectSSE()
        self.guilds = []
        self.currentGuild = nil
        self.currentGuildId = nil
    }
}

// MARK: - Notifications

extension Notification.Name {
    static let guildTimerEvent = Notification.Name("guildTimerEvent")
}
