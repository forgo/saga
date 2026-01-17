import Foundation

/// Server-Sent Events client for real-time family updates
actor SSEClient {
    private let baseURL: URL
    private var task: Task<Void, Never>?
    private var accessToken: String?
    private var eventHandler: ((FamilyEvent) -> Void)?
    private var reconnectAttempts = 0
    private let maxReconnectAttempts = 10

    init(baseURL: URL = URL(string: "http://localhost:8080/v1")!) {
        self.baseURL = baseURL
    }

    func setAccessToken(_ token: String) {
        self.accessToken = token
    }

    func connect(familyId: String, onEvent: @escaping @Sendable (FamilyEvent) -> Void) {
        disconnect()
        eventHandler = onEvent
        reconnectAttempts = 0

        task = Task {
            await connectWithRetry(familyId: familyId)
        }
    }

    func disconnect() {
        task?.cancel()
        task = nil
        eventHandler = nil
    }

    private func connectWithRetry(familyId: String) async {
        while !Task.isCancelled && reconnectAttempts < maxReconnectAttempts {
            do {
                try await streamEvents(familyId: familyId)
            } catch {
                if Task.isCancelled { return }

                reconnectAttempts += 1
                let delay = calculateBackoff(attempt: reconnectAttempts)
                print("SSE connection failed, retrying in \(delay)s (attempt \(reconnectAttempts))")

                try? await Task.sleep(for: .seconds(delay))
            }
        }
    }

    private func calculateBackoff(attempt: Int) -> Double {
        let baseDelay = 1.0
        let maxDelay = 60.0
        let delay = baseDelay * pow(2.0, Double(attempt - 1))
        let jitter = Double.random(in: 0...0.3) * delay
        return min(delay + jitter, maxDelay)
    }

    private func streamEvents(familyId: String) async throws {
        guard let token = accessToken else {
            throw APIError.unauthorized
        }

        let url = baseURL.appendingPathComponent("families/\(familyId)/events")
        var request = URLRequest(url: url)
        request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        request.setValue("text/event-stream", forHTTPHeaderField: "Accept")

        let (bytes, response) = try await URLSession.shared.bytes(for: request)

        guard let httpResponse = response as? HTTPURLResponse,
              httpResponse.statusCode == 200 else {
            throw APIError.httpError(statusCode: (response as? HTTPURLResponse)?.statusCode ?? 0, data: nil)
        }

        // Reset reconnect attempts on successful connection
        reconnectAttempts = 0

        var eventType: String?
        var dataLines: [String] = []

        for try await line in bytes.lines {
            if Task.isCancelled { break }

            if line.isEmpty {
                // Empty line signals end of event
                if let type = eventType, !dataLines.isEmpty {
                    let data = dataLines.joined(separator: "\n")
                    await processEvent(type: type, data: data)
                }
                eventType = nil
                dataLines = []
            } else if line.hasPrefix("event:") {
                eventType = String(line.dropFirst(6)).trimmingCharacters(in: .whitespaces)
            } else if line.hasPrefix("data:") {
                dataLines.append(String(line.dropFirst(5)).trimmingCharacters(in: .whitespaces))
            }
            // Ignore id: and retry: fields for now
        }
    }

    private func processEvent(type: String, data: String) async {
        guard let handler = eventHandler,
              let jsonData = data.data(using: .utf8) else { return }

        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601

        do {
            let event: FamilyEvent

            switch type {
            case "timer.created":
                let timer = try decoder.decode(BabyTimer.self, from: jsonData)
                event = .timerCreated(timer)

            case "timer.reset":
                let timer = try decoder.decode(BabyTimer.self, from: jsonData)
                event = .timerReset(timer)

            case "timer.updated":
                let timer = try decoder.decode(BabyTimer.self, from: jsonData)
                event = .timerUpdated(timer)

            case "timer.deleted":
                let payload = try decoder.decode(DeletedPayload.self, from: jsonData)
                event = .timerDeleted(id: payload.id)

            case "timer.warn":
                let alert = try decoder.decode(TimerAlert.self, from: jsonData)
                event = .timerWarn(alert)

            case "timer.critical":
                let alert = try decoder.decode(TimerAlert.self, from: jsonData)
                event = .timerCritical(alert)

            case "baby.created":
                let baby = try decoder.decode(Baby.self, from: jsonData)
                event = .babyCreated(baby)

            case "baby.updated":
                let baby = try decoder.decode(Baby.self, from: jsonData)
                event = .babyUpdated(baby)

            case "baby.deleted":
                let payload = try decoder.decode(DeletedPayload.self, from: jsonData)
                event = .babyDeleted(id: payload.id)

            case "activity.created":
                let activity = try decoder.decode(Activity.self, from: jsonData)
                event = .activityCreated(activity)

            case "activity.updated":
                let activity = try decoder.decode(Activity.self, from: jsonData)
                event = .activityUpdated(activity)

            case "activity.deleted":
                let payload = try decoder.decode(DeletedPayload.self, from: jsonData)
                event = .activityDeleted(id: payload.id)

            case "family.member_joined":
                let member = try decoder.decode(MemberEvent.self, from: jsonData)
                event = .memberJoined(member)

            case "family.member_left":
                let member = try decoder.decode(MemberEvent.self, from: jsonData)
                event = .memberLeft(member)

            case "heartbeat":
                event = .heartbeat

            default:
                print("Unknown SSE event type: \(type)")
                return
            }

            handler(event)

        } catch {
            print("Failed to decode SSE event '\(type)': \(error)")
        }
    }
}

// MARK: - Event Types

enum FamilyEvent: Sendable {
    case timerCreated(BabyTimer)
    case timerReset(BabyTimer)
    case timerUpdated(BabyTimer)
    case timerDeleted(id: String)
    case timerWarn(TimerAlert)
    case timerCritical(TimerAlert)

    case babyCreated(Baby)
    case babyUpdated(Baby)
    case babyDeleted(id: String)

    case activityCreated(Activity)
    case activityUpdated(Activity)
    case activityDeleted(id: String)

    case memberJoined(MemberEvent)
    case memberLeft(MemberEvent)

    case heartbeat
}

struct DeletedPayload: Codable, Sendable {
    let id: String
}

struct TimerAlert: Codable, Sendable {
    let id: String
    let threshold: TimeInterval
    let babyName: String
    let activityName: String

    enum CodingKeys: String, CodingKey {
        case id, threshold
        case babyName = "baby_name"
        case activityName = "activity_name"
    }
}

struct MemberEvent: Codable, Sendable {
    let parentId: String
    let name: String

    enum CodingKeys: String, CodingKey {
        case parentId = "parent_id"
        case name
    }
}
