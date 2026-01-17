import Foundation

/// Main API client for BabySync backend
actor APIClient {
    static let shared = APIClient()

    private let baseURL: URL
    private let session: URLSession
    private let decoder: JSONDecoder
    private let encoder: JSONEncoder
    private var accessToken: String?
    private var refreshToken: String?

    init(
        baseURL: URL = URL(string: "http://localhost:8080/v1")!,
        session: URLSession = .shared
    ) {
        self.baseURL = baseURL
        self.session = session

        self.decoder = JSONDecoder()
        self.decoder.dateDecodingStrategy = .iso8601

        self.encoder = JSONEncoder()
        self.encoder.dateEncodingStrategy = .iso8601
    }

    // MARK: - Token Management

    func setTokens(access: String, refresh: String) {
        self.accessToken = access
        self.refreshToken = refresh
    }

    func clearTokens() {
        self.accessToken = nil
        self.refreshToken = nil
    }

    var isAuthenticated: Bool {
        accessToken != nil
    }

    // MARK: - Request Building

    private func buildRequest(
        path: String,
        method: String,
        body: (any Encodable)? = nil,
        queryItems: [URLQueryItem]? = nil,
        requiresAuth: Bool = true,
        idempotencyKey: String? = nil
    ) throws -> URLRequest {
        var components = URLComponents(url: baseURL.appendingPathComponent(path), resolvingAgainstBaseURL: true)!
        components.queryItems = queryItems

        guard let url = components.url else {
            throw APIError.invalidURL
        }

        var request = URLRequest(url: url)
        request.httpMethod = method
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.setValue("application/json", forHTTPHeaderField: "Accept")
        request.setValue("gzip, br", forHTTPHeaderField: "Accept-Encoding")

        if requiresAuth, let token = accessToken {
            request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }

        if let key = idempotencyKey {
            request.setValue(key, forHTTPHeaderField: "Idempotency-Key")
        }

        if let body = body {
            request.httpBody = try encoder.encode(body)
        }

        return request
    }

    // MARK: - Request Execution

    private func execute<T: Decodable & Sendable>(_ request: URLRequest) async throws -> T {
        let (data, response) = try await session.data(for: request)

        guard let httpResponse = response as? HTTPURLResponse else {
            throw APIError.unknown
        }

        switch httpResponse.statusCode {
        case 200...299:
            do {
                return try decoder.decode(T.self, from: data)
            } catch {
                throw APIError.decodingError(error)
            }

        case 401:
            // Try to refresh token
            if let newTokens = try? await refreshAccessToken() {
                setTokens(access: newTokens.accessToken, refresh: newTokens.refreshToken)
                // Retry request with new token
                var retryRequest = request
                retryRequest.setValue("Bearer \(newTokens.accessToken)", forHTTPHeaderField: "Authorization")
                return try await execute(retryRequest)
            }
            throw APIError.unauthorized

        case 403:
            throw APIError.forbidden

        case 404:
            throw APIError.notFound

        case 429:
            let retryAfter = Int(httpResponse.value(forHTTPHeaderField: "Retry-After") ?? "30") ?? 30
            throw APIError.rateLimited(retryAfter: retryAfter)

        case 500...599:
            throw APIError.serverError

        default:
            if let problem = try? decoder.decode(ProblemDetails.self, from: data) {
                throw APIError.problemDetails(problem)
            }
            throw APIError.httpError(statusCode: httpResponse.statusCode, data: data)
        }
    }

    private func executeNoContent(_ request: URLRequest) async throws {
        let (data, response) = try await session.data(for: request)

        guard let httpResponse = response as? HTTPURLResponse else {
            throw APIError.unknown
        }

        switch httpResponse.statusCode {
        case 200...299:
            return

        case 401:
            throw APIError.unauthorized

        case 403:
            throw APIError.forbidden

        case 404:
            throw APIError.notFound

        case 429:
            let retryAfter = Int(httpResponse.value(forHTTPHeaderField: "Retry-After") ?? "30") ?? 30
            throw APIError.rateLimited(retryAfter: retryAfter)

        case 500...599:
            throw APIError.serverError

        default:
            if let problem = try? decoder.decode(ProblemDetails.self, from: data) {
                throw APIError.problemDetails(problem)
            }
            throw APIError.httpError(statusCode: httpResponse.statusCode, data: data)
        }
    }

    // MARK: - Token Refresh

    private func refreshAccessToken() async throws -> TokenResponse {
        guard let token = refreshToken else {
            throw APIError.unauthorized
        }

        let request = try buildRequest(
            path: "auth/refresh",
            method: "POST",
            body: RefreshRequest(refreshToken: token),
            requiresAuth: false
        )

        return try await execute(request)
    }

    // MARK: - Auth API

    func register(_ request: RegisterRequest) async throws -> DataResponse<AuthResponse> {
        let req = try buildRequest(path: "auth/register", method: "POST", body: request, requiresAuth: false)
        return try await execute(req)
    }

    func login(_ request: LoginRequest) async throws -> DataResponse<AuthResponse> {
        let req = try buildRequest(path: "auth/login", method: "POST", body: request, requiresAuth: false)
        return try await execute(req)
    }

    func loginWithGoogle(_ request: OAuthRequest) async throws -> DataResponse<AuthResponse> {
        let req = try buildRequest(path: "auth/oauth/google", method: "POST", body: request, requiresAuth: false)
        return try await execute(req)
    }

    func loginWithApple(_ request: OAuthRequest) async throws -> DataResponse<AuthResponse> {
        let req = try buildRequest(path: "auth/oauth/apple", method: "POST", body: request, requiresAuth: false)
        return try await execute(req)
    }

    func logout() async throws {
        let req = try buildRequest(path: "auth/logout", method: "POST")
        try await executeNoContent(req)
        clearTokens()
    }

    func getCurrentUser() async throws -> DataResponse<UserWithIdentities> {
        let req = try buildRequest(path: "auth/me", method: "GET")
        return try await execute(req)
    }

    // MARK: - Families API

    func listFamilies() async throws -> CollectionResponse<Family> {
        let req = try buildRequest(path: "families", method: "GET")
        return try await execute(req)
    }

    func createFamily(name: String? = nil) async throws -> DataResponse<Family> {
        struct CreateFamilyRequest: Codable { var name: String? }
        let req = try buildRequest(
            path: "families",
            method: "POST",
            body: CreateFamilyRequest(name: name),
            idempotencyKey: UUID().uuidString
        )
        return try await execute(req)
    }

    func getFamily(id: String) async throws -> DataResponse<FamilyData> {
        let req = try buildRequest(path: "families/\(id)", method: "GET")
        return try await execute(req)
    }

    func updateFamily(id: String, name: String) async throws -> DataResponse<Family> {
        struct UpdateFamilyRequest: Codable { let name: String }
        let req = try buildRequest(path: "families/\(id)", method: "PATCH", body: UpdateFamilyRequest(name: name))
        return try await execute(req)
    }

    func deleteFamily(id: String) async throws {
        let req = try buildRequest(path: "families/\(id)", method: "DELETE")
        try await executeNoContent(req)
    }

    func joinFamily(id: String) async throws {
        let req = try buildRequest(path: "families/\(id)/join", method: "POST", idempotencyKey: UUID().uuidString)
        try await executeNoContent(req)
    }

    func leaveFamily(id: String) async throws {
        let req = try buildRequest(path: "families/\(id)/leave", method: "POST")
        try await executeNoContent(req)
    }

    // MARK: - Babies API

    func listBabies(familyId: String) async throws -> CollectionResponse<Baby> {
        let req = try buildRequest(path: "families/\(familyId)/babies", method: "GET")
        return try await execute(req)
    }

    func createBaby(familyId: String, request: CreateBabyRequest) async throws -> DataResponse<Baby> {
        let req = try buildRequest(
            path: "families/\(familyId)/babies",
            method: "POST",
            body: request,
            idempotencyKey: UUID().uuidString
        )
        return try await execute(req)
    }

    func getBaby(familyId: String, babyId: String) async throws -> DataResponse<Baby> {
        let req = try buildRequest(path: "families/\(familyId)/babies/\(babyId)", method: "GET")
        return try await execute(req)
    }

    func updateBaby(familyId: String, babyId: String, request: UpdateBabyRequest) async throws -> DataResponse<Baby> {
        let req = try buildRequest(path: "families/\(familyId)/babies/\(babyId)", method: "PATCH", body: request)
        return try await execute(req)
    }

    func deleteBaby(familyId: String, babyId: String) async throws {
        let req = try buildRequest(path: "families/\(familyId)/babies/\(babyId)", method: "DELETE")
        try await executeNoContent(req)
    }

    // MARK: - Timers API

    func listTimers(familyId: String, babyId: String) async throws -> CollectionResponse<BabyTimer> {
        let req = try buildRequest(path: "families/\(familyId)/babies/\(babyId)/timers", method: "GET")
        return try await execute(req)
    }

    func createTimer(familyId: String, babyId: String, request: CreateTimerRequest) async throws -> DataResponse<BabyTimer> {
        let req = try buildRequest(
            path: "families/\(familyId)/babies/\(babyId)/timers",
            method: "POST",
            body: request,
            idempotencyKey: UUID().uuidString
        )
        return try await execute(req)
    }

    func updateTimer(familyId: String, babyId: String, timerId: String, request: UpdateTimerRequest) async throws -> DataResponse<BabyTimer> {
        let req = try buildRequest(
            path: "families/\(familyId)/babies/\(babyId)/timers/\(timerId)",
            method: "PATCH",
            body: request
        )
        return try await execute(req)
    }

    func deleteTimer(familyId: String, babyId: String, timerId: String) async throws {
        let req = try buildRequest(path: "families/\(familyId)/babies/\(babyId)/timers/\(timerId)", method: "DELETE")
        try await executeNoContent(req)
    }

    func resetTimer(familyId: String, babyId: String, timerId: String) async throws -> DataResponse<BabyTimer> {
        let req = try buildRequest(
            path: "families/\(familyId)/babies/\(babyId)/timers/\(timerId)/reset",
            method: "POST",
            idempotencyKey: UUID().uuidString
        )
        return try await execute(req)
    }

    // MARK: - Activities API

    func listActivities(familyId: String) async throws -> CollectionResponse<Activity> {
        let req = try buildRequest(path: "families/\(familyId)/activities", method: "GET")
        return try await execute(req)
    }

    func createActivity(familyId: String, request: CreateActivityRequest) async throws -> DataResponse<Activity> {
        let req = try buildRequest(
            path: "families/\(familyId)/activities",
            method: "POST",
            body: request,
            idempotencyKey: UUID().uuidString
        )
        return try await execute(req)
    }

    func updateActivity(familyId: String, activityId: String, request: UpdateActivityRequest) async throws -> DataResponse<Activity> {
        let req = try buildRequest(
            path: "families/\(familyId)/activities/\(activityId)",
            method: "PATCH",
            body: request
        )
        return try await execute(req)
    }

    func deleteActivity(familyId: String, activityId: String) async throws {
        let req = try buildRequest(path: "families/\(familyId)/activities/\(activityId)", method: "DELETE")
        try await executeNoContent(req)
    }
}
