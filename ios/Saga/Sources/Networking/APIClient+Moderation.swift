import Foundation

// MARK: - Moderation & Gamification API

extension APIClient {

    // MARK: - Reports

    /// Submit a report
    func submitReport(_ request: CreateReportRequest) async throws -> Report {
        let response: DataResponse<Report> = try await post(path: "/reports", body: request)
        return response.data
    }

    /// Get my reports (submitted and against me)
    func getMyReports() async throws -> MyReportsResponse {
        return try await get(path: "/reports/mine")
    }

    /// Get a specific report
    func getReport(reportId: String) async throws -> Report {
        let response: DataResponse<Report> = try await get(path: "/reports/\(reportId)")
        return response.data
    }

    // MARK: - Blocks

    /// Get my blocked users
    func getBlockedUsers() async throws -> [BlockDisplay] {
        let response: CollectionResponse<BlockDisplay> = try await get(path: "/blocks")
        return response.data
    }

    /// Block a user
    func blockUser(_ request: CreateBlockRequest) async throws -> Block {
        let response: DataResponse<Block> = try await post(path: "/blocks", body: request)
        return response.data
    }

    /// Unblock a user
    func unblockUser(blockedId: String) async throws {
        try await delete(path: "/blocks/\(blockedId)")
    }

    /// Check if a user is blocked
    func isUserBlocked(userId: String) async throws -> Bool {
        struct BlockedResponse: Codable {
            let isBlocked: Bool

            enum CodingKeys: String, CodingKey {
                case isBlocked = "is_blocked"
            }
        }
        let response: BlockedResponse = try await get(path: "/blocks/check/\(userId)")
        return response.isBlocked
    }

    // MARK: - Moderation Status

    /// Get my moderation status
    func getModerationStatus() async throws -> ModerationStatus {
        return try await get(path: "/moderation/status")
    }

    // MARK: - Resonance

    /// Get my resonance score
    func getMyResonance() async throws -> Resonance {
        return try await get(path: "/resonance")
    }

    /// Get resonance for a specific user
    func getResonance(userId: String) async throws -> Resonance {
        return try await get(path: "/resonance/\(userId)")
    }

    /// Get resonance leaderboard
    func getResonanceLeaderboard(guildId: String? = nil, limit: Int = 10) async throws -> [ResonanceRanking] {
        var queryItems: [URLQueryItem] = [
            URLQueryItem(name: "limit", value: String(limit))
        ]
        if let guildId = guildId {
            queryItems.append(URLQueryItem(name: "guild_id", value: guildId))
        }
        let response: CollectionResponse<ResonanceRanking> = try await get(path: "/resonance/leaderboard", queryItems: queryItems)
        return response.data
    }

    // MARK: - Devices

    /// Get my registered devices
    func getMyDevices() async throws -> [Device] {
        let response: CollectionResponse<Device> = try await get(path: "/devices")
        return response.data
    }

    /// Register a device for push notifications
    func registerDevice(_ request: RegisterDeviceRequest) async throws -> Device {
        let response: DataResponse<Device> = try await post(path: "/devices", body: request)
        return response.data
    }

    /// Update a device
    func updateDevice(deviceId: String, _ request: UpdateDeviceRequest) async throws -> Device {
        let response: DataResponse<Device> = try await patch(path: "/devices/\(deviceId)", body: request)
        return response.data
    }

    /// Unregister a device
    func unregisterDevice(deviceId: String) async throws {
        try await delete(path: "/devices/\(deviceId)")
    }
}

// MARK: - Resonance Ranking

/// A user's position on the resonance leaderboard
struct ResonanceRanking: Codable, Identifiable, Sendable {
    let rank: Int
    let userId: String
    let user: PublicProfile
    let level: ResonanceLevel
    let totalScore: Int

    var id: String { userId }

    enum CodingKeys: String, CodingKey {
        case rank
        case userId = "user_id"
        case user
        case level
        case totalScore = "total_score"
    }
}
