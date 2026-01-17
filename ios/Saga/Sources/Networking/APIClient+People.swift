import Foundation

// MARK: - People API Extension

extension APIClient {

    // MARK: - People CRUD

    /// List people in guild
    func listPeople(guildId: String) async throws -> CollectionResponse<Person> {
        try await get(path: "guilds/\(guildId)/people")
    }

    /// Create person
    func createPerson(guildId: String, _ request: CreatePersonRequest) async throws -> DataResponse<Person> {
        try await post(path: "guilds/\(guildId)/people", body: request, idempotencyKey: UUID().uuidString)
    }

    /// Get person with timers
    func getPerson(guildId: String, personId: String) async throws -> DataResponse<PersonWithTimers> {
        try await get(path: "guilds/\(guildId)/people/\(personId)")
    }

    /// Update person
    func updatePerson(guildId: String, personId: String, _ request: UpdatePersonRequest) async throws -> DataResponse<Person> {
        try await patch(path: "guilds/\(guildId)/people/\(personId)", body: request)
    }

    /// Delete person
    func deletePerson(guildId: String, personId: String) async throws {
        try await delete(path: "guilds/\(guildId)/people/\(personId)")
    }

    // MARK: - Timers

    /// List timers for person
    func listTimers(guildId: String, personId: String) async throws -> CollectionResponse<TimerWithActivity> {
        try await get(path: "guilds/\(guildId)/people/\(personId)/timers")
    }

    /// Create timer
    func createTimer(guildId: String, personId: String, _ request: CreateTimerRequest) async throws -> DataResponse<ActivityTimer> {
        try await post(
            path: "guilds/\(guildId)/people/\(personId)/timers",
            body: request,
            idempotencyKey: UUID().uuidString
        )
    }

    /// Get timer details
    func getTimer(guildId: String, personId: String, timerId: String) async throws -> DataResponse<ActivityTimer> {
        try await get(path: "guilds/\(guildId)/people/\(personId)/timers/\(timerId)")
    }

    /// Update timer
    func updateTimer(guildId: String, personId: String, timerId: String, _ request: UpdateTimerRequest) async throws -> DataResponse<ActivityTimer> {
        try await patch(path: "guilds/\(guildId)/people/\(personId)/timers/\(timerId)", body: request)
    }

    /// Delete timer
    func deleteTimer(guildId: String, personId: String, timerId: String) async throws {
        try await delete(path: "guilds/\(guildId)/people/\(personId)/timers/\(timerId)")
    }

    /// Reset timer to current time
    func resetTimer(guildId: String, personId: String, timerId: String) async throws -> DataResponse<ActivityTimer> {
        try await post(
            path: "guilds/\(guildId)/people/\(personId)/timers/\(timerId)/reset",
            idempotencyKey: UUID().uuidString
        )
    }
}
