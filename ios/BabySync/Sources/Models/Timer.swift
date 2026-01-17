import Foundation

struct BabyTimer: Codable, Identifiable, Equatable, Sendable {
    let id: String
    let babyId: String
    let activityId: String
    var resetDate: Date
    var enabled: Bool
    var push: Bool
    let createdOn: Date
    let updatedOn: Date

    enum CodingKeys: String, CodingKey {
        case id, enabled, push
        case babyId = "baby_id"
        case activityId = "activity_id"
        case resetDate = "reset_date"
        case createdOn = "created_on"
        case updatedOn = "updated_on"
    }

    /// Elapsed time since last reset (computed client-side)
    var elapsed: TimeInterval {
        guard enabled else { return 0 }
        return Date().timeIntervalSince(resetDate)
    }

    /// Formatted elapsed time (HH:MM:SS)
    var elapsedFormatted: String {
        let totalSeconds = Int(elapsed)
        let hours = totalSeconds / 3600
        let minutes = (totalSeconds % 3600) / 60
        let seconds = totalSeconds % 60
        return String(format: "%02d:%02d:%02d", hours, minutes, seconds)
    }

    /// Check if timer is in warning state
    func isWarning(threshold: TimeInterval) -> Bool {
        elapsed >= threshold
    }

    /// Check if timer is in critical state
    func isCritical(threshold: TimeInterval) -> Bool {
        elapsed >= threshold
    }
}

struct CreateTimerRequest: Codable, Sendable {
    let activityId: String
    var enabled: Bool = true
    var push: Bool = false

    enum CodingKeys: String, CodingKey {
        case activityId = "activity_id"
        case enabled, push
    }
}

struct UpdateTimerRequest: Codable, Sendable {
    var enabled: Bool?
    var push: Bool?
}
