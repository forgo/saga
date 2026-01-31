import Foundation

/// Service for managing profile, availability, trust, and reviews
@Observable
final class ProfileService: @unchecked Sendable {
    // MARK: - Shared Instance

    static let shared = ProfileService()

    // MARK: - Profile State

    private(set) var myProfile: Profile?
    private(set) var isLoadingProfile = false

    // MARK: - Availability State

    private(set) var myAvailabilities: [Availability] = []
    private(set) var nearbyAvailabilities: [NearbyAvailability] = []
    private(set) var incomingRequests: [HangoutRequestDisplay] = []
    private(set) var outgoingRequests: [HangoutRequestDisplay] = []
    private(set) var isLoadingAvailability = false

    // MARK: - Trust State

    private(set) var myTrustGrants: [TrustGrant] = []
    private(set) var trustGrantsToMe: [TrustGrant] = []
    private(set) var pendingIRLRequests: [IRLConfirmation] = []
    private(set) var isLoadingTrust = false

    // MARK: - Review State

    private(set) var myReviewSummary: ReviewSummary?
    private(set) var isLoadingReviews = false

    // MARK: - Error State

    private(set) var error: Error?

    // MARK: - Dependencies

    private let apiClient: APIClient

    // MARK: - Init

    private init(apiClient: APIClient = .shared) {
        self.apiClient = apiClient
    }

    // MARK: - Clear

    @MainActor
    func clear() {
        myProfile = nil
        myAvailabilities = []
        nearbyAvailabilities = []
        incomingRequests = []
        outgoingRequests = []
        myTrustGrants = []
        trustGrantsToMe = []
        pendingIRLRequests = []
        myReviewSummary = nil
        error = nil
    }

    // MARK: - Profile Methods

    func loadMyProfile() async {
        await MainActor.run { isLoadingProfile = true; error = nil }

        do {
            let profile = try await apiClient.getMyProfile()
            await MainActor.run { myProfile = profile }
        } catch {
            await MainActor.run { self.error = error }
        }

        await MainActor.run { isLoadingProfile = false }
    }

    func updateProfile(_ request: UpdateProfileRequest) async throws -> Profile {
        let profile = try await apiClient.updateProfile(request)
        await MainActor.run { myProfile = profile }
        return profile
    }

    func getPublicProfile(userId: String) async throws -> PublicProfile {
        return try await apiClient.getPublicProfile(userId: userId)
    }

    func searchNearbyProfiles(lat: Double, lng: Double, radiusKm: Double = 10) async throws -> [NearbyProfile] {
        return try await apiClient.getNearbyProfiles(lat: lat, lng: lng, radiusKm: radiusKm)
    }

    // MARK: - Availability Methods

    func loadMyAvailabilities() async {
        await MainActor.run { isLoadingAvailability = true; error = nil }

        do {
            let availabilities = try await apiClient.getMyAvailabilities()
            await MainActor.run { myAvailabilities = availabilities }
        } catch {
            await MainActor.run { self.error = error }
        }

        await MainActor.run { isLoadingAvailability = false }
    }

    func createAvailability(_ request: CreateAvailabilityRequest) async throws -> Availability {
        let availability = try await apiClient.createAvailability(request)
        await MainActor.run { myAvailabilities.insert(availability, at: 0) }
        return availability
    }

    func updateAvailability(availabilityId: String, _ request: UpdateAvailabilityRequest) async throws -> Availability {
        let availability = try await apiClient.updateAvailability(availabilityId: availabilityId, request)
        await MainActor.run {
            if let index = myAvailabilities.firstIndex(where: { $0.id == availabilityId }) {
                myAvailabilities[index] = availability
            }
        }
        return availability
    }

    func cancelAvailability(availabilityId: String) async throws {
        try await apiClient.cancelAvailability(availabilityId: availabilityId)
        await MainActor.run { myAvailabilities.removeAll { $0.id == availabilityId } }
    }

    func searchNearbyAvailabilities(lat: Double, lng: Double, radiusKm: Double = 10, hangoutType: HangoutType? = nil) async {
        await MainActor.run { isLoadingAvailability = true; error = nil }

        do {
            let availabilities = try await apiClient.getNearbyAvailabilities(
                lat: lat,
                lng: lng,
                radiusKm: radiusKm,
                hangoutType: hangoutType
            )
            await MainActor.run { nearbyAvailabilities = availabilities }
        } catch {
            await MainActor.run { self.error = error }
        }

        await MainActor.run { isLoadingAvailability = false }
    }

    // MARK: - Hangout Request Methods

    func loadHangoutRequests() async {
        await MainActor.run { isLoadingAvailability = true; error = nil }

        do {
            async let incoming = apiClient.getIncomingHangoutRequests()
            async let outgoing = apiClient.getOutgoingHangoutRequests()

            let (incomingResult, outgoingResult) = try await (incoming, outgoing)
            await MainActor.run {
                incomingRequests = incomingResult
                outgoingRequests = outgoingResult
            }
        } catch {
            await MainActor.run { self.error = error }
        }

        await MainActor.run { isLoadingAvailability = false }
    }

    func sendHangoutRequest(availabilityId: String, message: String? = nil) async throws -> HangoutRequest {
        return try await apiClient.createHangoutRequest(availabilityId: availabilityId, message: message)
    }

    func respondToHangoutRequest(requestId: String, accept: Bool, message: String? = nil) async throws -> HangoutRequest {
        let response = try await apiClient.respondToHangoutRequest(requestId: requestId, accept: accept, message: message)
        await MainActor.run { incomingRequests.removeAll { $0.request.id == requestId } }
        return response
    }

    // MARK: - Trust Grant Methods

    func loadTrustGrants() async {
        await MainActor.run { isLoadingTrust = true; error = nil }

        do {
            async let grants = apiClient.getMyTrustGrants()
            async let received = apiClient.getTrustGrantsToMe()

            let (grantsResult, receivedResult) = try await (grants, received)
            await MainActor.run {
                myTrustGrants = grantsResult
                trustGrantsToMe = receivedResult
            }
        } catch {
            await MainActor.run { self.error = error }
        }

        await MainActor.run { isLoadingTrust = false }
    }

    func grantTrust(_ request: CreateTrustRequest) async throws -> TrustGrant {
        let grant = try await apiClient.createTrustGrant(request)
        await MainActor.run { myTrustGrants.insert(grant, at: 0) }
        return grant
    }

    func updateTrustGrant(grantId: String, _ request: UpdateTrustRequest) async throws -> TrustGrant {
        let grant = try await apiClient.updateTrustGrant(grantId: grantId, request)
        await MainActor.run {
            if let index = myTrustGrants.firstIndex(where: { $0.id == grantId }) {
                myTrustGrants[index] = grant
            }
        }
        return grant
    }

    func revokeTrustGrant(grantId: String) async throws {
        try await apiClient.revokeTrustGrant(grantId: grantId)
        await MainActor.run { myTrustGrants.removeAll { $0.id == grantId } }
    }

    func getTrustSummary(userId: String) async throws -> TrustSummary {
        return try await apiClient.getTrustSummary(userId: userId)
    }

    // MARK: - IRL Confirmation Methods

    func loadPendingIRLRequests() async {
        await MainActor.run { isLoadingTrust = true; error = nil }

        do {
            let requests = try await apiClient.getPendingIRLRequests()
            await MainActor.run { pendingIRLRequests = requests }
        } catch {
            await MainActor.run { self.error = error }
        }

        await MainActor.run { isLoadingTrust = false }
    }

    func requestIRLConfirmation(_ request: RequestIRLRequest) async throws -> IRLConfirmation {
        return try await apiClient.requestIRLConfirmation(request)
    }

    func respondToIRLRequest(confirmationId: String, confirm: Bool) async throws -> IRLConfirmation {
        let response = try await apiClient.respondToIRLRequest(confirmationId: confirmationId, confirm: confirm)
        await MainActor.run { pendingIRLRequests.removeAll { $0.id == confirmationId } }
        return response
    }

    // MARK: - Trust Rating Methods

    func createTrustRating(_ request: CreateTrustRatingRequest) async throws -> TrustRating {
        return try await apiClient.createTrustRating(request)
    }

    func getTrustRatings(userId: String) async throws -> [TrustRatingWithEndorsements] {
        return try await apiClient.getTrustRatings(userId: userId)
    }

    func getTrustAggregate(userId: String) async throws -> TrustAggregate {
        return try await apiClient.getTrustAggregate(userId: userId)
    }

    func endorseTrustRating(ratingId: String, _ request: CreateEndorsementRequest) async throws -> TrustEndorsement {
        return try await apiClient.createTrustEndorsement(ratingId: ratingId, request)
    }

    // MARK: - Review Methods

    func loadMyReviewSummary() async {
        await MainActor.run { isLoadingReviews = true; error = nil }

        do {
            let summary = try await apiClient.getMyReviewSummary()
            await MainActor.run { myReviewSummary = summary }
        } catch {
            await MainActor.run { self.error = error }
        }

        await MainActor.run { isLoadingReviews = false }
    }

    func createReview(_ request: CreateReviewRequest) async throws -> Review {
        return try await apiClient.createReview(request)
    }

    func getReviews(userId: String) async throws -> [ReviewWithAuthor] {
        return try await apiClient.getReviews(userId: userId)
    }

    func updateReview(reviewId: String, _ request: UpdateReviewRequest) async throws -> Review {
        return try await apiClient.updateReview(reviewId: reviewId, request)
    }

    func deleteReview(reviewId: String) async throws {
        try await apiClient.deleteReview(reviewId: reviewId)
    }
}
