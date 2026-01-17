import Foundation

// MARK: - Advanced API

extension APIClient {

    // MARK: - Adventures

    /// Get adventures for a guild
    func getAdventures(guildId: String) async throws -> [Adventure] {
        let response: CollectionResponse<Adventure> = try await get(path: "/guilds/\(guildId)/adventures")
        return response.data
    }

    /// Get adventure details
    func getAdventure(adventureId: String) async throws -> AdventureWithDetails {
        return try await get(path: "/adventures/\(adventureId)")
    }

    /// Create an adventure
    func createAdventure(_ request: CreateAdventureRequest) async throws -> Adventure {
        let response: DataResponse<Adventure> = try await post(path: "/adventures", body: request)
        return response.data
    }

    /// Update an adventure
    func updateAdventure(adventureId: String, _ request: UpdateAdventureRequest) async throws -> Adventure {
        let response: DataResponse<Adventure> = try await patch(path: "/adventures/\(adventureId)", body: request)
        return response.data
    }

    /// Cancel an adventure
    func cancelAdventure(adventureId: String) async throws {
        try await delete(path: "/adventures/\(adventureId)")
    }

    /// Request to join an adventure
    func joinAdventure(adventureId: String, message: String? = nil) async throws -> AdventureParticipant {
        let request = JoinAdventureRequest(message: message)
        let response: DataResponse<AdventureParticipant> = try await post(path: "/adventures/\(adventureId)/join", body: request)
        return response.data
    }

    /// Leave an adventure
    func leaveAdventure(adventureId: String) async throws {
        try await delete(path: "/adventures/\(adventureId)/leave")
    }

    /// Get pending join requests (host only)
    func getAdventureJoinRequests(adventureId: String) async throws -> [AdventureParticipantDisplay] {
        let response: CollectionResponse<AdventureParticipantDisplay> = try await get(path: "/adventures/\(adventureId)/requests")
        return response.data
    }

    /// Respond to a join request (host only)
    func respondToAdventureJoinRequest(adventureId: String, participantId: String, approve: Bool, message: String? = nil) async throws -> AdventureParticipant {
        let request = RespondToJoinRequest(approve: approve, message: message)
        let response: DataResponse<AdventureParticipant> = try await post(path: "/adventures/\(adventureId)/requests/\(participantId)/respond", body: request)
        return response.data
    }

    // MARK: - Pools

    /// Get pools for a guild
    func getPools(guildId: String) async throws -> [Pool] {
        let response: CollectionResponse<Pool> = try await get(path: "/guilds/\(guildId)/pools")
        return response.data
    }

    /// Get pool details
    func getPool(poolId: String) async throws -> PoolWithDetails {
        return try await get(path: "/pools/\(poolId)")
    }

    /// Create a pool
    func createPool(_ request: CreatePoolRequest) async throws -> Pool {
        let response: DataResponse<Pool> = try await post(path: "/pools", body: request)
        return response.data
    }

    /// Update a pool
    func updatePool(poolId: String, _ request: UpdatePoolRequest) async throws -> Pool {
        let response: DataResponse<Pool> = try await patch(path: "/pools/\(poolId)", body: request)
        return response.data
    }

    /// Delete a pool
    func deletePool(poolId: String) async throws {
        try await delete(path: "/pools/\(poolId)")
    }

    /// Join a pool
    func joinPool(poolId: String, preferences: PoolPreferences? = nil) async throws -> PoolMember {
        let request = JoinPoolRequest(preferences: preferences)
        let response: DataResponse<PoolMember> = try await post(path: "/pools/\(poolId)/join", body: request)
        return response.data
    }

    /// Leave a pool
    func leavePool(poolId: String) async throws {
        try await delete(path: "/pools/\(poolId)/leave")
    }

    /// Update pool preferences
    func updatePoolPreferences(poolId: String, preferences: PoolPreferences) async throws -> PoolMember {
        let request = UpdatePoolPreferencesRequest(preferences: preferences)
        let response: DataResponse<PoolMember> = try await patch(path: "/pools/\(poolId)/preferences", body: request)
        return response.data
    }

    /// Get my pool matches
    func getMyPoolMatches(poolId: String? = nil) async throws -> [PoolMatchDisplay] {
        var path = "/pools/matches"
        if let poolId = poolId {
            path = "/pools/\(poolId)/matches"
        }
        let response: CollectionResponse<PoolMatchDisplay> = try await get(path: path)
        return response.data
    }

    /// Respond to a pool match
    func respondToPoolMatch(matchId: String, accept: Bool) async throws -> PoolMatch {
        let request = RespondToMatchRequest(accept: accept)
        let response: DataResponse<PoolMatch> = try await post(path: "/pools/matches/\(matchId)/respond", body: request)
        return response.data
    }

    /// Schedule a match meeting
    func schedulePoolMatch(matchId: String, meetingTime: Date) async throws -> PoolMatch {
        let request = ScheduleMatchRequest(meetingTime: meetingTime)
        let response: DataResponse<PoolMatch> = try await post(path: "/pools/matches/\(matchId)/schedule", body: request)
        return response.data
    }

    /// Submit match feedback
    func submitMatchFeedback(matchId: String, metUp: Bool, rating: Int?, notes: String?) async throws {
        let request = MatchFeedbackRequest(metUp: metUp, rating: rating, notes: notes)
        try await postNoContent(path: "/pools/matches/\(matchId)/feedback", body: request)
    }

    // MARK: - Votes

    /// Get votes for a guild
    func getVotes(guildId: String, status: VoteStatus? = nil) async throws -> [Vote] {
        var queryItems: [URLQueryItem] = []
        if let status = status {
            queryItems.append(URLQueryItem(name: "status", value: status.rawValue))
        }
        let response: CollectionResponse<Vote> = try await get(path: "/guilds/\(guildId)/votes", queryItems: queryItems)
        return response.data
    }

    /// Get vote details
    func getVote(voteId: String) async throws -> VoteWithDetails {
        return try await get(path: "/votes/\(voteId)")
    }

    /// Create a vote
    func createVote(_ request: CreateVoteRequest) async throws -> Vote {
        let response: DataResponse<Vote> = try await post(path: "/votes", body: request)
        return response.data
    }

    /// Update a vote
    func updateVote(voteId: String, _ request: UpdateVoteRequest) async throws -> Vote {
        let response: DataResponse<Vote> = try await patch(path: "/votes/\(voteId)", body: request)
        return response.data
    }

    /// End a vote early
    func endVote(voteId: String) async throws -> Vote {
        let response: DataResponse<Vote> = try await post(path: "/votes/\(voteId)/end", body: EmptyBody())
        return response.data
    }

    /// Cast a ballot
    func castBallot(voteId: String, selections: [BallotSelection], abstain: Bool = false) async throws -> Ballot {
        let request = CastBallotRequest(selections: selections, abstain: abstain)
        let response: DataResponse<Ballot> = try await post(path: "/votes/\(voteId)/ballot", body: request)
        return response.data
    }

    /// Get vote results
    func getVoteResults(voteId: String) async throws -> VoteResults {
        let response: DataResponse<VoteResults> = try await get(path: "/votes/\(voteId)/results")
        return response.data
    }

    // MARK: - Role Catalogs

    /// Get role catalogs for a guild
    func getRoleCatalogs(guildId: String) async throws -> [RoleCatalog] {
        let response: CollectionResponse<RoleCatalog> = try await get(path: "/guilds/\(guildId)/role-catalogs")
        return response.data
    }

    /// Create a role catalog entry
    func createRoleCatalog(_ request: CreateRoleCatalogRequest) async throws -> RoleCatalog {
        let response: DataResponse<RoleCatalog> = try await post(path: "/role-catalogs", body: request)
        return response.data
    }

    /// Update a role catalog entry
    func updateRoleCatalog(catalogId: String, _ request: UpdateRoleCatalogRequest) async throws -> RoleCatalog {
        let response: DataResponse<RoleCatalog> = try await patch(path: "/role-catalogs/\(catalogId)", body: request)
        return response.data
    }

    /// Delete a role catalog entry
    func deleteRoleCatalog(catalogId: String) async throws {
        try await delete(path: "/role-catalogs/\(catalogId)")
    }
}

// MARK: - Empty Body

private struct EmptyBody: Codable, Sendable {}
