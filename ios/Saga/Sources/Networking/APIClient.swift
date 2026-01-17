import Foundation

/// Main API client for Saga backend
actor APIClient {
    static let shared = APIClient()

    private let baseURL: URL
    private let session: URLSession
    private let decoder: JSONDecoder
    private let encoder: JSONEncoder
    private var accessToken: String?
    private var refreshToken: String?
    private var isRefreshing = false
    private var pendingRequests: [CheckedContinuation<Void, Never>] = []

    init(
        baseURL: URL = currentEnvironment.baseURL,
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

    func getAccessToken() -> String? {
        accessToken
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
        components.queryItems = queryItems?.isEmpty == false ? queryItems : nil

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

    func execute<T: Decodable & Sendable>(_ request: URLRequest) async throws -> T {
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

        case 409:
            if let problem = try? decoder.decode(ProblemDetails.self, from: data) {
                throw APIError.problemDetails(problem)
            }
            throw APIError.conflict

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

    func executeNoContent(_ request: URLRequest) async throws {
        let (data, response) = try await session.data(for: request)

        guard let httpResponse = response as? HTTPURLResponse else {
            throw APIError.unknown
        }

        switch httpResponse.statusCode {
        case 200...299:
            return

        case 401:
            // Try to refresh token
            if let newTokens = try? await refreshAccessToken() {
                setTokens(access: newTokens.accessToken, refresh: newTokens.refreshToken)
                // Retry request with new token
                var retryRequest = request
                retryRequest.setValue("Bearer \(newTokens.accessToken)", forHTTPHeaderField: "Authorization")
                return try await executeNoContent(retryRequest)
            }
            throw APIError.unauthorized

        case 403:
            throw APIError.forbidden

        case 404:
            throw APIError.notFound

        case 409:
            if let problem = try? decoder.decode(ProblemDetails.self, from: data) {
                throw APIError.problemDetails(problem)
            }
            throw APIError.conflict

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

        // Prevent multiple simultaneous refresh requests
        if isRefreshing {
            await withCheckedContinuation { continuation in
                pendingRequests.append(continuation)
            }
            // After waiting, check if we now have a valid token
            guard let newToken = accessToken else {
                throw APIError.unauthorized
            }
            return TokenResponse(
                accessToken: newToken,
                refreshToken: refreshToken ?? "",
                tokenType: "Bearer",
                expiresIn: 3600
            )
        }

        isRefreshing = true
        defer {
            isRefreshing = false
            // Resume all pending requests
            for continuation in pendingRequests {
                continuation.resume()
            }
            pendingRequests.removeAll()
        }

        let request = try buildRequest(
            path: "auth/refresh",
            method: "POST",
            body: RefreshRequest(refreshToken: token),
            requiresAuth: false
        )

        let (data, response) = try await session.data(for: request)

        guard let httpResponse = response as? HTTPURLResponse else {
            throw APIError.unknown
        }

        guard httpResponse.statusCode == 200 else {
            throw APIError.unauthorized
        }

        let tokenResponse: DataResponse<TokenResponse> = try decoder.decode(DataResponse<TokenResponse>.self, from: data)
        return tokenResponse.data
    }

    // MARK: - Public Request Helpers

    func get<T: Decodable & Sendable>(
        path: String,
        queryItems: [URLQueryItem]? = nil,
        requiresAuth: Bool = true
    ) async throws -> T {
        let request = try buildRequest(
            path: path,
            method: "GET",
            queryItems: queryItems,
            requiresAuth: requiresAuth
        )
        return try await execute(request)
    }

    func post<T: Decodable & Sendable>(
        path: String,
        body: (any Encodable)? = nil,
        requiresAuth: Bool = true,
        idempotencyKey: String? = nil
    ) async throws -> T {
        let request = try buildRequest(
            path: path,
            method: "POST",
            body: body,
            requiresAuth: requiresAuth,
            idempotencyKey: idempotencyKey
        )
        return try await execute(request)
    }

    func postNoContent(
        path: String,
        body: (any Encodable)? = nil,
        requiresAuth: Bool = true,
        idempotencyKey: String? = nil
    ) async throws {
        let request = try buildRequest(
            path: path,
            method: "POST",
            body: body,
            requiresAuth: requiresAuth,
            idempotencyKey: idempotencyKey
        )
        try await executeNoContent(request)
    }

    func patch<T: Decodable & Sendable>(
        path: String,
        body: (any Encodable)? = nil,
        requiresAuth: Bool = true
    ) async throws -> T {
        let request = try buildRequest(
            path: path,
            method: "PATCH",
            body: body,
            requiresAuth: requiresAuth
        )
        return try await execute(request)
    }

    func put<T: Decodable & Sendable>(
        path: String,
        body: (any Encodable)? = nil,
        requiresAuth: Bool = true
    ) async throws -> T {
        let request = try buildRequest(
            path: path,
            method: "PUT",
            body: body,
            requiresAuth: requiresAuth
        )
        return try await execute(request)
    }

    func delete(
        path: String,
        requiresAuth: Bool = true
    ) async throws {
        let request = try buildRequest(
            path: path,
            method: "DELETE",
            requiresAuth: requiresAuth
        )
        try await executeNoContent(request)
    }
}
