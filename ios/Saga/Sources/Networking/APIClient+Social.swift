import Foundation

// MARK: - Social API

extension APIClient {

    // MARK: - Profile

    /// Get my profile
    func getMyProfile() async throws -> Profile {
        let response: DataResponse<Profile> = try await get(path: "/profile")
        return response.data
    }

    /// Update my profile
    func updateProfile(_ request: UpdateProfileRequest) async throws -> Profile {
        let response: DataResponse<Profile> = try await patch(path: "/profile", body: request)
        return response.data
    }

    /// Get another user's public profile
    func getPublicProfile(userId: String) async throws -> PublicProfile {
        let response: DataResponse<PublicProfile> = try await get(path: "/profiles/\(userId)")
        return response.data
    }

    /// Search for nearby profiles
    func getNearbyProfiles(lat: Double, lng: Double, radiusKm: Double = 10, limit: Int = 20) async throws -> [NearbyProfile] {
        let queryItems = [
            URLQueryItem(name: "lat", value: String(lat)),
            URLQueryItem(name: "lng", value: String(lng)),
            URLQueryItem(name: "radius_km", value: String(radiusKm)),
            URLQueryItem(name: "limit", value: String(limit))
        ]
        let response: CollectionResponse<NearbyProfile> = try await get(path: "/profiles/nearby", queryItems: queryItems)
        return response.data
    }

    // MARK: - Availability

    /// Create an availability posting
    func createAvailability(_ request: CreateAvailabilityRequest) async throws -> Availability {
        let response: DataResponse<Availability> = try await post(path: "/availability", body: request)
        return response.data
    }

    /// Get my availabilities
    func getMyAvailabilities() async throws -> [Availability] {
        let response: CollectionResponse<Availability> = try await get(path: "/availability")
        return response.data
    }

    /// Update an availability
    func updateAvailability(availabilityId: String, _ request: UpdateAvailabilityRequest) async throws -> Availability {
        let response: DataResponse<Availability> = try await patch(path: "/availability/\(availabilityId)", body: request)
        return response.data
    }

    /// Cancel an availability
    func cancelAvailability(availabilityId: String) async throws {
        try await delete(path: "/availability/\(availabilityId)")
    }

    /// Find nearby availabilities
    func getNearbyAvailabilities(lat: Double, lng: Double, radiusKm: Double = 10, hangoutType: HangoutType? = nil) async throws -> [NearbyAvailability] {
        var queryItems = [
            URLQueryItem(name: "lat", value: String(lat)),
            URLQueryItem(name: "lng", value: String(lng)),
            URLQueryItem(name: "radius_km", value: String(radiusKm))
        ]
        if let hangoutType = hangoutType {
            queryItems.append(URLQueryItem(name: "hangout_type", value: hangoutType.rawValue))
        }
        let response: CollectionResponse<NearbyAvailability> = try await get(path: "/availability/nearby", queryItems: queryItems)
        return response.data
    }

    // MARK: - Hangout Requests

    /// Send a hangout request for an availability
    func createHangoutRequest(availabilityId: String, message: String? = nil) async throws -> HangoutRequest {
        let request = CreateHangoutRequestRequest(message: message)
        let response: DataResponse<HangoutRequest> = try await post(path: "/availability/\(availabilityId)/request", body: request)
        return response.data
    }

    /// Get incoming hangout requests for my availabilities
    func getIncomingHangoutRequests() async throws -> [HangoutRequestDisplay] {
        let response: CollectionResponse<HangoutRequestDisplay> = try await get(path: "/availability/requests/incoming")
        return response.data
    }

    /// Get my outgoing hangout requests
    func getOutgoingHangoutRequests() async throws -> [HangoutRequestDisplay] {
        let response: CollectionResponse<HangoutRequestDisplay> = try await get(path: "/availability/requests/outgoing")
        return response.data
    }

    /// Respond to a hangout request
    func respondToHangoutRequest(requestId: String, accept: Bool, message: String? = nil) async throws -> HangoutRequest {
        let request = RespondHangoutRequest(accept: accept, message: message)
        let response: DataResponse<HangoutRequest> = try await post(path: "/availability/requests/\(requestId)/respond", body: request)
        return response.data
    }

    // MARK: - Trust Grants

    /// Grant trust to another user
    func createTrustGrant(_ request: CreateTrustRequest) async throws -> TrustGrant {
        let response: DataResponse<TrustGrant> = try await post(path: "/trust/grants", body: request)
        return response.data
    }

    /// Get users I've granted trust to
    func getMyTrustGrants() async throws -> [TrustGrant] {
        let response: CollectionResponse<TrustGrant> = try await get(path: "/trust/grants")
        return response.data
    }

    /// Get users who have granted trust to me
    func getTrustGrantsToMe() async throws -> [TrustGrant] {
        let response: CollectionResponse<TrustGrant> = try await get(path: "/trust/grants/received")
        return response.data
    }

    /// Update a trust grant
    func updateTrustGrant(grantId: String, _ request: UpdateTrustRequest) async throws -> TrustGrant {
        let response: DataResponse<TrustGrant> = try await patch(path: "/trust/grants/\(grantId)", body: request)
        return response.data
    }

    /// Revoke a trust grant
    func revokeTrustGrant(grantId: String) async throws {
        try await delete(path: "/trust/grants/\(grantId)")
    }

    /// Get trust summary for a user
    func getTrustSummary(userId: String) async throws -> TrustSummary {
        let response: DataResponse<TrustSummary> = try await get(path: "/trust/summary/\(userId)")
        return response.data
    }

    // MARK: - IRL Confirmations

    /// Request IRL confirmation from another user
    func requestIRLConfirmation(_ request: RequestIRLRequest) async throws -> IRLConfirmation {
        let response: DataResponse<IRLConfirmation> = try await post(path: "/trust/irl", body: request)
        return response.data
    }

    /// Get pending IRL confirmation requests
    func getPendingIRLRequests() async throws -> [IRLConfirmation] {
        let response: CollectionResponse<IRLConfirmation> = try await get(path: "/trust/irl/pending")
        return response.data
    }

    /// Respond to an IRL confirmation request
    func respondToIRLRequest(confirmationId: String, confirm: Bool) async throws -> IRLConfirmation {
        let request = AcceptIRLRequest(confirm: confirm)
        let response: DataResponse<IRLConfirmation> = try await post(path: "/trust/irl/\(confirmationId)/respond", body: request)
        return response.data
    }

    // MARK: - Trust Ratings

    /// Create a trust rating for someone
    func createTrustRating(_ request: CreateTrustRatingRequest) async throws -> TrustRating {
        let response: DataResponse<TrustRating> = try await post(path: "/trust/ratings", body: request)
        return response.data
    }

    /// Get trust ratings for a user
    func getTrustRatings(userId: String) async throws -> [TrustRatingWithEndorsements] {
        let response: CollectionResponse<TrustRatingWithEndorsements> = try await get(path: "/trust/ratings/\(userId)")
        return response.data
    }

    /// Get trust aggregate for a user
    func getTrustAggregate(userId: String) async throws -> TrustAggregate {
        let response: DataResponse<TrustAggregate> = try await get(path: "/trust/aggregate/\(userId)")
        return response.data
    }

    // MARK: - Trust Endorsements

    /// Endorse (agree/disagree) with a trust rating
    func createTrustEndorsement(ratingId: String, _ request: CreateEndorsementRequest) async throws -> TrustEndorsement {
        let response: DataResponse<TrustEndorsement> = try await post(path: "/trust/ratings/\(ratingId)/endorse", body: request)
        return response.data
    }

    // MARK: - Reviews

    /// Create a review for another user
    func createReview(_ request: CreateReviewRequest) async throws -> Review {
        let response: DataResponse<Review> = try await post(path: "/reviews", body: request)
        return response.data
    }

    /// Get reviews for a user
    func getReviews(userId: String) async throws -> [ReviewWithAuthor] {
        let response: CollectionResponse<ReviewWithAuthor> = try await get(path: "/reviews/\(userId)")
        return response.data
    }

    /// Get my review summary
    func getMyReviewSummary() async throws -> ReviewSummary {
        let response: DataResponse<ReviewSummary> = try await get(path: "/reviews/summary")
        return response.data
    }

    /// Update my review
    func updateReview(reviewId: String, _ request: UpdateReviewRequest) async throws -> Review {
        let response: DataResponse<Review> = try await patch(path: "/reviews/\(reviewId)", body: request)
        return response.data
    }

    /// Delete my review
    func deleteReview(reviewId: String) async throws {
        try await delete(path: "/reviews/\(reviewId)")
    }
}
