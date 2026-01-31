import Foundation

// MARK: - Guilds API Extension

extension APIClient {

    // MARK: - Guild CRUD

    /// List user's guilds
    func listGuilds() async throws -> CollectionResponse<Guild> {
        try await get(path: "guilds")
    }

    /// Create a new guild
    func createGuild(_ request: CreateGuildRequest) async throws -> DataResponse<Guild> {
        try await post(path: "guilds", body: request, idempotencyKey: UUID().uuidString)
    }

    /// Get guild with all data (members, people, activities)
    func getGuild(id: String) async throws -> DataResponse<GuildData> {
        try await get(path: "guilds/\(id)")
    }

    /// Update guild
    func updateGuild(id: String, _ request: UpdateGuildRequest) async throws -> DataResponse<Guild> {
        try await patch(path: "guilds/\(id)", body: request)
    }

    /// Delete guild
    func deleteGuild(id: String) async throws {
        try await delete(path: "guilds/\(id)")
    }

    // MARK: - Guild Membership

    /// Get guild members
    func getGuildMembers(guildId: String) async throws -> CollectionResponse<Member> {
        try await get(path: "guilds/\(guildId)/members")
    }

    /// Join a guild
    func joinGuild(id: String) async throws {
        try await postNoContent(path: "guilds/\(id)/join", idempotencyKey: UUID().uuidString)
    }

    /// Leave a guild
    func leaveGuild(id: String) async throws {
        try await postNoContent(path: "guilds/\(id)/leave")
    }

    /// Merge another guild into this one
    func mergeGuilds(targetId: String, sourceId: String) async throws -> DataResponse<GuildData> {
        let request = MergeGuildsRequest(sourceGuildId: sourceId)
        return try await post(path: "guilds/\(targetId)/merge", body: request)
    }

    // MARK: - Activities

    /// List activities in guild
    func listActivities(guildId: String) async throws -> CollectionResponse<Activity> {
        try await get(path: "guilds/\(guildId)/activities")
    }

    /// Create activity
    func createActivity(guildId: String, _ request: CreateActivityRequest) async throws -> DataResponse<Activity> {
        try await post(path: "guilds/\(guildId)/activities", body: request, idempotencyKey: UUID().uuidString)
    }

    /// Get activity details
    func getActivity(guildId: String, activityId: String) async throws -> DataResponse<Activity> {
        try await get(path: "guilds/\(guildId)/activities/\(activityId)")
    }

    /// Update activity
    func updateActivity(guildId: String, activityId: String, _ request: UpdateActivityRequest) async throws -> DataResponse<Activity> {
        try await patch(path: "guilds/\(guildId)/activities/\(activityId)", body: request)
    }

    /// Delete activity
    func deleteActivity(guildId: String, activityId: String) async throws {
        try await delete(path: "guilds/\(guildId)/activities/\(activityId)")
    }
}
