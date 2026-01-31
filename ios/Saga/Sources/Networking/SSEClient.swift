import Foundation

/// Server-Sent Events client for real-time guild updates
actor SSEClient {
    private var task: URLSessionDataTask?
    private var session: URLSession?
    private let baseURL: URL
    private var eventHandler: (@Sendable @MainActor (GuildEvent) -> Void)?
    private var connectionHandler: (@Sendable @MainActor (Bool) -> Void)?
    private var buffer = ""

    init(baseURL: URL = currentEnvironment.baseURL) {
        self.baseURL = baseURL
    }

    /// Connect to guild events stream
    func connect(
        guildId: String,
        accessToken: String,
        onEvent: @escaping @Sendable @MainActor (GuildEvent) -> Void,
        onConnectionChange: @escaping @Sendable @MainActor (Bool) -> Void
    ) {
        disconnect()

        self.eventHandler = onEvent
        self.connectionHandler = onConnectionChange

        let url = baseURL.appendingPathComponent("guilds/\(guildId)/events")
        var request = URLRequest(url: url)
        request.setValue("Bearer \(accessToken)", forHTTPHeaderField: "Authorization")
        request.setValue("text/event-stream", forHTTPHeaderField: "Accept")
        request.timeoutInterval = TimeInterval(Int.max) // Long-lived connection

        let configuration = URLSessionConfiguration.default
        configuration.timeoutIntervalForRequest = TimeInterval(Int.max)
        configuration.timeoutIntervalForResource = TimeInterval(Int.max)

        let delegate = SSEDelegate { [weak self] data in
            Task { [weak self] in
                await self?.handleData(data)
            }
        } onComplete: { [weak self] error in
            Task { [weak self] in
                await self?.handleDisconnect(error: error)
            }
        }

        session = URLSession(configuration: configuration, delegate: delegate, delegateQueue: nil)
        task = session?.dataTask(with: request)
        task?.resume()

        if let handler = connectionHandler {
            Task { @MainActor in
                handler(true)
            }
        }
    }

    /// Disconnect from the events stream
    func disconnect() {
        task?.cancel()
        task = nil
        session?.invalidateAndCancel()
        session = nil
        buffer = ""
        if let handler = connectionHandler {
            Task { @MainActor in
                handler(false)
            }
        }
    }

    private func handleData(_ data: Data) {
        guard let string = String(data: data, encoding: .utf8) else { return }
        buffer += string

        // Parse SSE format: lines separated by \n\n
        while let range = buffer.range(of: "\n\n") {
            let eventString = String(buffer[..<range.lowerBound])
            buffer = String(buffer[range.upperBound...])

            if let event = parseEvent(eventString), let handler = eventHandler {
                Task { @MainActor in
                    handler(event)
                }
            }
        }
    }

    private func parseEvent(_ string: String) -> GuildEvent? {
        var eventType: String?
        var eventData: String?

        for line in string.split(separator: "\n", omittingEmptySubsequences: false) {
            let lineStr = String(line)
            if lineStr.hasPrefix("event:") {
                eventType = String(lineStr.dropFirst(6)).trimmingCharacters(in: .whitespaces)
            } else if lineStr.hasPrefix("data:") {
                eventData = String(lineStr.dropFirst(5)).trimmingCharacters(in: .whitespaces)
            }
        }

        guard let type = eventType else { return nil }

        // Handle heartbeat
        if type == "heartbeat" {
            return .heartbeat
        }

        guard let data = eventData,
              let jsonData = data.data(using: .utf8) else { return nil }

        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601

        do {
            switch type {
            // Timer events
            case "timer.created":
                let timer = try decoder.decode(ActivityTimer.self, from: jsonData)
                return .timerCreated(timer)
            case "timer.reset":
                let timer = try decoder.decode(ActivityTimer.self, from: jsonData)
                return .timerReset(timer)
            case "timer.updated":
                let timer = try decoder.decode(ActivityTimer.self, from: jsonData)
                return .timerUpdated(timer)
            case "timer.deleted":
                let payload = try decoder.decode(DeletedPayload.self, from: jsonData)
                return .timerDeleted(id: payload.id)
            case "timer.warn":
                let alert = try decoder.decode(TimerAlert.self, from: jsonData)
                return .timerWarn(alert)
            case "timer.critical":
                let alert = try decoder.decode(TimerAlert.self, from: jsonData)
                return .timerCritical(alert)

            // Person events
            case "person.created":
                let person = try decoder.decode(Person.self, from: jsonData)
                return .personCreated(person)
            case "person.updated":
                let person = try decoder.decode(Person.self, from: jsonData)
                return .personUpdated(person)
            case "person.deleted":
                let payload = try decoder.decode(DeletedPayload.self, from: jsonData)
                return .personDeleted(id: payload.id)

            // Activity events
            case "activity.created":
                let activity = try decoder.decode(Activity.self, from: jsonData)
                return .activityCreated(activity)
            case "activity.updated":
                let activity = try decoder.decode(Activity.self, from: jsonData)
                return .activityUpdated(activity)
            case "activity.deleted":
                let payload = try decoder.decode(DeletedPayload.self, from: jsonData)
                return .activityDeleted(id: payload.id)

            // Guild events
            case "guild.member_joined":
                let member = try decoder.decode(Member.self, from: jsonData)
                return .memberJoined(member)
            case "guild.member_left":
                let payload = try decoder.decode(DeletedPayload.self, from: jsonData)
                return .memberLeft(id: payload.id)

            default:
                return nil
            }
        } catch {
            print("SSE parse error for \(type): \(error)")
            return nil
        }
    }

    private func handleDisconnect(error: Error?) {
        if let handler = connectionHandler {
            Task { @MainActor in
                handler(false)
            }
        }

        // Auto-reconnect after delay if not cancelled
        if let error = error as NSError?, error.code != NSURLErrorCancelled {
            // Could implement reconnection logic here
        }
    }
}

// MARK: - Guild Events

enum GuildEvent: Sendable {
    case heartbeat

    // Timer events
    case timerCreated(ActivityTimer)
    case timerReset(ActivityTimer)
    case timerUpdated(ActivityTimer)
    case timerDeleted(id: String)
    case timerWarn(TimerAlert)
    case timerCritical(TimerAlert)

    // Person events
    case personCreated(Person)
    case personUpdated(Person)
    case personDeleted(id: String)

    // Activity events
    case activityCreated(Activity)
    case activityUpdated(Activity)
    case activityDeleted(id: String)

    // Member events
    case memberJoined(Member)
    case memberLeft(id: String)
}

/// Timer threshold alert
struct TimerAlert: Codable, Sendable {
    let timerId: String
    let personId: String
    let activityId: String
    let personName: String
    let activityName: String
    let elapsed: Int

    enum CodingKeys: String, CodingKey {
        case timerId = "timer_id"
        case personId = "person_id"
        case activityId = "activity_id"
        case personName = "person_name"
        case activityName = "activity_name"
        case elapsed
    }
}

/// Payload for delete events
private struct DeletedPayload: Codable {
    let id: String
}

// MARK: - SSE Delegate

private final class SSEDelegate: NSObject, URLSessionDataDelegate, @unchecked Sendable {
    private let onData: (Data) -> Void
    private let onComplete: (Error?) -> Void

    init(onData: @escaping (Data) -> Void, onComplete: @escaping (Error?) -> Void) {
        self.onData = onData
        self.onComplete = onComplete
    }

    func urlSession(_ session: URLSession, dataTask: URLSessionDataTask, didReceive data: Data) {
        onData(data)
    }

    func urlSession(_ session: URLSession, task: URLSessionTask, didCompleteWithError error: Error?) {
        onComplete(error)
    }
}
