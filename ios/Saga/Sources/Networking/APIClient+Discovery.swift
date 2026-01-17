import Foundation

// MARK: - Discovery API

extension APIClient {

    // MARK: - Interest Categories

    /// Get all interest categories
    func getInterestCategories() async throws -> [InterestCategory] {
        let response: CollectionResponse<InterestCategory> = try await get(path: "/interests/categories")
        return response.data
    }

    /// Get interests for a category
    func getInterests(categoryId: String) async throws -> [Interest] {
        let response: CollectionResponse<Interest> = try await get(path: "/interests/categories/\(categoryId)/interests")
        return response.data
    }

    /// Search interests
    func searchInterests(query: String) async throws -> [Interest] {
        let queryItems = [URLQueryItem(name: "q", value: query)]
        let response: CollectionResponse<Interest> = try await get(path: "/interests/search", queryItems: queryItems)
        return response.data
    }

    // MARK: - User Interests

    /// Get my interests
    func getMyInterests() async throws -> [UserInterest] {
        let response: CollectionResponse<UserInterest> = try await get(path: "/interests/mine")
        return response.data
    }

    /// Add an interest
    func addInterest(_ request: AddInterestRequest) async throws -> UserInterest {
        let response: DataResponse<UserInterest> = try await post(path: "/interests/mine", body: request)
        return response.data
    }

    /// Update an interest
    func updateInterest(userInterestId: String, _ request: UpdateInterestRequest) async throws -> UserInterest {
        let response: DataResponse<UserInterest> = try await patch(path: "/interests/mine/\(userInterestId)", body: request)
        return response.data
    }

    /// Remove an interest
    func removeInterest(userInterestId: String) async throws {
        try await delete(path: "/interests/mine/\(userInterestId)")
    }

    // MARK: - Interest Matching

    /// Find users with matching interests
    func findInterestMatches(interestId: String? = nil, limit: Int = 20) async throws -> [InterestMatch] {
        var queryItems = [URLQueryItem(name: "limit", value: String(limit))]
        if let interestId = interestId {
            queryItems.append(URLQueryItem(name: "interest_id", value: interestId))
        }
        let response: CollectionResponse<InterestMatch> = try await get(path: "/interests/matches", queryItems: queryItems)
        return response.data
    }

    // MARK: - Questionnaires

    /// Get all questionnaires
    func getQuestionnaires() async throws -> [Questionnaire] {
        let response: CollectionResponse<Questionnaire> = try await get(path: "/questionnaires")
        return response.data
    }

    /// Get a specific questionnaire with questions
    func getQuestionnaire(questionnaireId: String) async throws -> Questionnaire {
        let response: DataResponse<Questionnaire> = try await get(path: "/questionnaires/\(questionnaireId)")
        return response.data
    }

    /// Get my questionnaire progress
    func getQuestionnaireProgress() async throws -> [QuestionnaireProgress] {
        let response: CollectionResponse<QuestionnaireProgress> = try await get(path: "/questionnaires/progress")
        return response.data
    }

    /// Get my responses for a questionnaire
    func getMyResponses(questionnaireId: String) async throws -> [QuestionResponse] {
        let response: CollectionResponse<QuestionResponse> = try await get(path: "/questionnaires/\(questionnaireId)/responses")
        return response.data
    }

    /// Submit a response to a question
    func submitResponse(_ request: SubmitResponseRequest) async throws -> QuestionResponse {
        let response: DataResponse<QuestionResponse> = try await post(path: "/questionnaires/responses", body: request)
        return response.data
    }

    /// Submit multiple responses at once
    func submitResponses(_ requests: [SubmitResponseRequest]) async throws -> [QuestionResponse] {
        let response: CollectionResponse<QuestionResponse> = try await post(path: "/questionnaires/responses/batch", body: requests)
        return response.data
    }

    // MARK: - Compatibility

    /// Get compatibility score with another user
    func getCompatibility(userId: String) async throws -> CompatibilityScore {
        let response: DataResponse<CompatibilityScore> = try await get(path: "/compatibility/\(userId)")
        return response.data
    }

    // MARK: - Discovery

    /// Discover people nearby
    func discoverPeople(
        lat: Double? = nil,
        lng: Double? = nil,
        radiusKm: Double? = nil,
        minCompatibility: Double? = nil,
        interestIds: [String]? = nil,
        limit: Int = 20
    ) async throws -> [DiscoveryResult] {
        var queryItems = [URLQueryItem(name: "limit", value: String(limit))]
        if let lat = lat { queryItems.append(URLQueryItem(name: "lat", value: String(lat))) }
        if let lng = lng { queryItems.append(URLQueryItem(name: "lng", value: String(lng))) }
        if let radius = radiusKm { queryItems.append(URLQueryItem(name: "radius_km", value: String(radius))) }
        if let minCompat = minCompatibility { queryItems.append(URLQueryItem(name: "min_compatibility", value: String(minCompat))) }
        if let ids = interestIds, !ids.isEmpty {
            queryItems.append(URLQueryItem(name: "interest_ids", value: ids.joined(separator: ",")))
        }

        let response: CollectionResponse<DiscoveryResult> = try await get(path: "/discover/people", queryItems: queryItems)
        return response.data
    }

    /// Discover events (complementing the existing discoverEvents in APIClient+Events)
    func discoverEventsNearby(
        lat: Double,
        lng: Double,
        radiusKm: Double = 50,
        limit: Int = 20
    ) async throws -> [Event] {
        let queryItems = [
            URLQueryItem(name: "lat", value: String(lat)),
            URLQueryItem(name: "lng", value: String(lng)),
            URLQueryItem(name: "radius_km", value: String(radiusKm)),
            URLQueryItem(name: "limit", value: String(limit))
        ]
        let response: CollectionResponse<Event> = try await get(path: "/discover/events/nearby", queryItems: queryItems)
        return response.data
    }
}
