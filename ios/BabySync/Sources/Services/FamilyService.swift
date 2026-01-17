import Foundation

/// Manages family data and real-time synchronization
@Observable
final class FamilyService: @unchecked Sendable {
    static let shared = FamilyService()

    // Current family state
    private(set) var families: [Family] = []
    private(set) var currentFamily: FamilyData?
    private(set) var isLoading = false
    private(set) var error: APIError?

    // Computed properties for current family data
    var babies: [Baby] { currentFamily?.babies ?? [] }
    var activities: [Activity] { currentFamily?.activities ?? [] }
    var parents: [Parent] { currentFamily?.parents ?? [] }

    private let apiClient = APIClient.shared
    private let sseClient = SSEClient()
    private var sseToken: String?

    // MARK: - Families

    func loadFamilies() async throws {
        await MainActor.run { isLoading = true }
        defer { Task { @MainActor in isLoading = false } }

        let response = try await apiClient.listFamilies()
        await MainActor.run {
            self.families = response.data
        }
    }

    func selectFamily(id: String) async throws {
        await MainActor.run { isLoading = true }
        defer { Task { @MainActor in isLoading = false } }

        let response = try await apiClient.getFamily(id: id)
        await MainActor.run {
            self.currentFamily = response.data
        }

        // Connect to SSE for real-time updates
        if let token = sseToken {
            await sseClient.setAccessToken(token)
            await sseClient.connect(familyId: id) { [weak self] event in
                Task { @MainActor in
                    self?.handleEvent(event)
                }
            }
        }
    }

    func createFamily(name: String? = nil) async throws -> Family {
        let response = try await apiClient.createFamily(name: name)
        await MainActor.run {
            self.families.append(response.data)
        }
        return response.data
    }

    func updateFamily(id: String, name: String) async throws {
        let response = try await apiClient.updateFamily(id: id, name: name)
        await MainActor.run {
            if let index = self.families.firstIndex(where: { $0.id == id }) {
                self.families[index] = response.data
            }
            if self.currentFamily?.family.id == id {
                self.currentFamily?.family = response.data
            }
        }
    }

    func joinFamily(id: String) async throws {
        try await apiClient.joinFamily(id: id)
        try await loadFamilies()
    }

    func leaveFamily(id: String) async throws {
        try await apiClient.leaveFamily(id: id)
        await MainActor.run {
            self.families.removeAll { $0.id == id }
            if self.currentFamily?.family.id == id {
                self.currentFamily = nil
                Task { await self.sseClient.disconnect() }
            }
        }
    }

    // MARK: - Babies

    func createBaby(name: String) async throws {
        guard let familyId = currentFamily?.family.id else { return }
        let request = CreateBabyRequest(name: name)
        let response = try await apiClient.createBaby(familyId: familyId, request: request)
        await MainActor.run {
            self.currentFamily?.babies.append(response.data)
        }
    }

    func updateBaby(id: String, name: String) async throws {
        guard let familyId = currentFamily?.family.id else { return }
        let request = UpdateBabyRequest(name: name)
        let response = try await apiClient.updateBaby(familyId: familyId, babyId: id, request: request)
        await MainActor.run {
            if let index = self.currentFamily?.babies.firstIndex(where: { $0.id == id }) {
                self.currentFamily?.babies[index] = response.data
            }
        }
    }

    func deleteBaby(id: String) async throws {
        guard let familyId = currentFamily?.family.id else { return }
        try await apiClient.deleteBaby(familyId: familyId, babyId: id)
        await MainActor.run {
            self.currentFamily?.babies.removeAll { $0.id == id }
        }
    }

    // MARK: - Activities

    func createActivity(name: String, icon: String, warn: TimeInterval, critical: TimeInterval) async throws {
        guard let familyId = currentFamily?.family.id else { return }
        let request = CreateActivityRequest(name: name, icon: icon, warn: warn, critical: critical)
        let response = try await apiClient.createActivity(familyId: familyId, request: request)
        await MainActor.run {
            self.currentFamily?.activities.append(response.data)
        }
    }

    func deleteActivity(id: String) async throws {
        guard let familyId = currentFamily?.family.id else { return }
        try await apiClient.deleteActivity(familyId: familyId, activityId: id)
        await MainActor.run {
            self.currentFamily?.activities.removeAll { $0.id == id }
        }
    }

    // MARK: - SSE Setup

    func setSSEToken(_ token: String) {
        sseToken = token
    }

    func disconnectSSE() async {
        await sseClient.disconnect()
    }

    // MARK: - Event Handling

    @MainActor
    private func handleEvent(_ event: FamilyEvent) {
        switch event {
        case .timerCreated(let timer):
            // Timer events are handled by TimerViewModel
            NotificationCenter.default.post(name: .timerCreated, object: timer)

        case .timerReset(let timer):
            NotificationCenter.default.post(name: .timerReset, object: timer)

        case .timerUpdated(let timer):
            NotificationCenter.default.post(name: .timerUpdated, object: timer)

        case .timerDeleted(let id):
            NotificationCenter.default.post(name: .timerDeleted, object: id)

        case .timerWarn(let alert):
            NotificationCenter.default.post(name: .timerWarn, object: alert)

        case .timerCritical(let alert):
            NotificationCenter.default.post(name: .timerCritical, object: alert)

        case .babyCreated(let baby):
            currentFamily?.babies.append(baby)

        case .babyUpdated(let baby):
            if let index = currentFamily?.babies.firstIndex(where: { $0.id == baby.id }) {
                currentFamily?.babies[index] = baby
            }

        case .babyDeleted(let id):
            currentFamily?.babies.removeAll { $0.id == id }

        case .activityCreated(let activity):
            currentFamily?.activities.append(activity)

        case .activityUpdated(let activity):
            if let index = currentFamily?.activities.firstIndex(where: { $0.id == activity.id }) {
                currentFamily?.activities[index] = activity
            }

        case .activityDeleted(let id):
            currentFamily?.activities.removeAll { $0.id == id }

        case .memberJoined(let member):
            let parent = Parent(id: member.parentId, name: member.name, email: "")
            currentFamily?.parents.append(parent)

        case .memberLeft(let member):
            currentFamily?.parents.removeAll { $0.id == member.parentId }

        case .heartbeat:
            break
        }
    }
}

// MARK: - Notification Names

extension Notification.Name {
    static let timerCreated = Notification.Name("timerCreated")
    static let timerReset = Notification.Name("timerReset")
    static let timerUpdated = Notification.Name("timerUpdated")
    static let timerDeleted = Notification.Name("timerDeleted")
    static let timerWarn = Notification.Name("timerWarn")
    static let timerCritical = Notification.Name("timerCritical")
}
