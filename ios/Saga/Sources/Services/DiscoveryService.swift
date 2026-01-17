import Foundation

/// Service for managing discovery, interests, and questionnaires
@Observable
final class DiscoveryService: @unchecked Sendable {
    // MARK: - Shared Instance

    static let shared = DiscoveryService()

    // MARK: - Interest State

    private(set) var interestCategories: [InterestCategory] = []
    private(set) var myInterests: [UserInterest] = []
    private(set) var interestMatches: [InterestMatch] = []
    private(set) var isLoadingInterests = false

    // MARK: - Questionnaire State

    private(set) var questionnaires: [Questionnaire] = []
    private(set) var questionnaireProgress: [QuestionnaireProgress] = []
    private(set) var currentQuestionnaire: Questionnaire?
    private(set) var currentResponses: [QuestionResponse] = []
    private(set) var isLoadingQuestionnaires = false

    // MARK: - Discovery State

    private(set) var discoveryResults: [DiscoveryResult] = []
    private(set) var nearbyEvents: [Event] = []
    private(set) var isLoadingDiscovery = false

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
        interestCategories = []
        myInterests = []
        interestMatches = []
        questionnaires = []
        questionnaireProgress = []
        currentQuestionnaire = nil
        currentResponses = []
        discoveryResults = []
        nearbyEvents = []
        error = nil
    }

    // MARK: - Interest Category Methods

    func loadInterestCategories() async {
        await MainActor.run { isLoadingInterests = true; error = nil }

        do {
            let categories = try await apiClient.getInterestCategories()
            await MainActor.run { interestCategories = categories }
        } catch {
            await MainActor.run { self.error = error }
        }

        await MainActor.run { isLoadingInterests = false }
    }

    func getInterests(categoryId: String) async throws -> [Interest] {
        return try await apiClient.getInterests(categoryId: categoryId)
    }

    func searchInterests(query: String) async throws -> [Interest] {
        return try await apiClient.searchInterests(query: query)
    }

    // MARK: - User Interest Methods

    func loadMyInterests() async {
        await MainActor.run { isLoadingInterests = true; error = nil }

        do {
            let interests = try await apiClient.getMyInterests()
            await MainActor.run { myInterests = interests }
        } catch {
            await MainActor.run { self.error = error }
        }

        await MainActor.run { isLoadingInterests = false }
    }

    func addInterest(_ request: AddInterestRequest) async throws -> UserInterest {
        let interest = try await apiClient.addInterest(request)
        await MainActor.run { myInterests.insert(interest, at: 0) }
        return interest
    }

    func updateInterest(userInterestId: String, _ request: UpdateInterestRequest) async throws -> UserInterest {
        let interest = try await apiClient.updateInterest(userInterestId: userInterestId, request)
        await MainActor.run {
            if let index = myInterests.firstIndex(where: { $0.id == userInterestId }) {
                myInterests[index] = interest
            }
        }
        return interest
    }

    func removeInterest(userInterestId: String) async throws {
        try await apiClient.removeInterest(userInterestId: userInterestId)
        await MainActor.run { myInterests.removeAll { $0.id == userInterestId } }
    }

    // MARK: - Interest Matching Methods

    func loadInterestMatches(interestId: String? = nil) async {
        await MainActor.run { isLoadingInterests = true; error = nil }

        do {
            let matches = try await apiClient.findInterestMatches(interestId: interestId)
            await MainActor.run { interestMatches = matches }
        } catch {
            await MainActor.run { self.error = error }
        }

        await MainActor.run { isLoadingInterests = false }
    }

    // MARK: - Questionnaire Methods

    func loadQuestionnaires() async {
        await MainActor.run { isLoadingQuestionnaires = true; error = nil }

        do {
            async let questionnairesResult = apiClient.getQuestionnaires()
            async let progressResult = apiClient.getQuestionnaireProgress()

            let (loadedQuestionnaires, loadedProgress) = try await (questionnairesResult, progressResult)
            await MainActor.run {
                questionnaires = loadedQuestionnaires
                questionnaireProgress = loadedProgress
            }
        } catch {
            await MainActor.run { self.error = error }
        }

        await MainActor.run { isLoadingQuestionnaires = false }
    }

    func loadQuestionnaire(questionnaireId: String) async {
        await MainActor.run { isLoadingQuestionnaires = true; error = nil }

        do {
            async let questionnaireResult = apiClient.getQuestionnaire(questionnaireId: questionnaireId)
            async let responsesResult = apiClient.getMyResponses(questionnaireId: questionnaireId)

            let (questionnaire, responses) = try await (questionnaireResult, responsesResult)
            await MainActor.run {
                currentQuestionnaire = questionnaire
                currentResponses = responses
            }
        } catch {
            await MainActor.run { self.error = error }
        }

        await MainActor.run { isLoadingQuestionnaires = false }
    }

    func submitResponse(_ request: SubmitResponseRequest) async throws -> QuestionResponse {
        let response = try await apiClient.submitResponse(request)
        await MainActor.run {
            // Update or add the response
            if let index = currentResponses.firstIndex(where: { $0.questionId == request.questionId }) {
                currentResponses[index] = response
            } else {
                currentResponses.append(response)
            }
        }
        return response
    }

    func submitResponses(_ requests: [SubmitResponseRequest]) async throws -> [QuestionResponse] {
        let responses = try await apiClient.submitResponses(requests)
        await MainActor.run {
            for response in responses {
                if let index = currentResponses.firstIndex(where: { $0.questionId == response.questionId }) {
                    currentResponses[index] = response
                } else {
                    currentResponses.append(response)
                }
            }
        }
        return responses
    }

    /// Get progress for a specific questionnaire
    func progress(for questionnaireId: String) -> QuestionnaireProgress? {
        questionnaireProgress.first { $0.questionnaireId == questionnaireId }
    }

    /// Overall questionnaire completion percentage
    var overallProgress: Double {
        guard !questionnaireProgress.isEmpty else { return 0 }
        let totalQuestions = questionnaireProgress.reduce(0) { $0 + $1.totalQuestions }
        let answeredQuestions = questionnaireProgress.reduce(0) { $0 + $1.answeredQuestions }
        guard totalQuestions > 0 else { return 0 }
        return Double(answeredQuestions) / Double(totalQuestions)
    }

    // MARK: - Compatibility Methods

    func getCompatibility(userId: String) async throws -> CompatibilityScore {
        return try await apiClient.getCompatibility(userId: userId)
    }

    // MARK: - Discovery Methods

    func discoverPeople(
        lat: Double? = nil,
        lng: Double? = nil,
        radiusKm: Double? = nil,
        minCompatibility: Double? = nil,
        interestIds: [String]? = nil
    ) async {
        await MainActor.run { isLoadingDiscovery = true; error = nil }

        do {
            let results = try await apiClient.discoverPeople(
                lat: lat,
                lng: lng,
                radiusKm: radiusKm,
                minCompatibility: minCompatibility,
                interestIds: interestIds
            )
            await MainActor.run { discoveryResults = results }
        } catch {
            await MainActor.run { self.error = error }
        }

        await MainActor.run { isLoadingDiscovery = false }
    }

    func discoverEventsNearby(lat: Double, lng: Double, radiusKm: Double = 50) async {
        await MainActor.run { isLoadingDiscovery = true; error = nil }

        do {
            let events = try await apiClient.discoverEventsNearby(lat: lat, lng: lng, radiusKm: radiusKm)
            await MainActor.run { nearbyEvents = events }
        } catch {
            await MainActor.run { self.error = error }
        }

        await MainActor.run { isLoadingDiscovery = false }
    }
}
