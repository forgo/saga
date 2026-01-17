import Foundation
import SwiftUI

// MARK: - Timer

/// A timer tracking when an activity was last performed with a person
struct ActivityTimer: Codable, Sendable, Identifiable, Hashable {
    let id: String
    let resetDate: Date
    let enabled: Bool
    let push: Bool
    let createdOn: Date?
    let updatedOn: Date?

    enum CodingKeys: String, CodingKey {
        case id
        case resetDate = "reset_date"
        case enabled, push
        case createdOn = "created_on"
        case updatedOn = "updated_on"
    }

    /// Time elapsed since last reset
    var elapsed: TimeInterval {
        Date().timeIntervalSince(resetDate)
    }

    /// Elapsed time in seconds
    var elapsedSeconds: Int {
        Int(elapsed)
    }

    /// Human-readable elapsed time
    var elapsedDescription: String {
        formatElapsed(elapsedSeconds)
    }

    /// Timer status based on activity thresholds
    func status(for activity: Activity) -> TimerStatus {
        let elapsed = elapsedSeconds
        if elapsed >= activity.critical {
            return .critical
        } else if elapsed >= activity.warn {
            return .warning
        } else {
            return .good
        }
    }

    /// Progress towards warning threshold (0-1)
    func progressToWarning(for activity: Activity) -> Double {
        min(1.0, Double(elapsedSeconds) / Double(activity.warn))
    }

    /// Progress towards critical threshold (0-1)
    func progressToCritical(for activity: Activity) -> Double {
        min(1.0, Double(elapsedSeconds) / Double(activity.critical))
    }

    private func formatElapsed(_ seconds: Int) -> String {
        let days = seconds / 86400
        let hours = (seconds % 86400) / 3600
        let minutes = (seconds % 3600) / 60

        if days > 0 {
            return days == 1 ? "1 day ago" : "\(days) days ago"
        } else if hours > 0 {
            return hours == 1 ? "1 hour ago" : "\(hours) hours ago"
        } else if minutes > 0 {
            return minutes == 1 ? "1 minute ago" : "\(minutes) minutes ago"
        } else {
            return "Just now"
        }
    }
}

// MARK: - Timer Status

enum TimerStatus: String, Sendable {
    case good
    case warning
    case critical

    var color: Color {
        switch self {
        case .good: return .green
        case .warning: return .orange
        case .critical: return .red
        }
    }

    var iconName: String {
        switch self {
        case .good: return "checkmark.circle.fill"
        case .warning: return "exclamationmark.triangle.fill"
        case .critical: return "exclamationmark.octagon.fill"
        }
    }
}

// MARK: - Timer with Activity

/// Timer bundled with its associated activity
struct TimerWithActivity: Codable, Sendable, Identifiable {
    let timer: ActivityTimer
    let activity: Activity

    var id: String { timer.id }

    /// Convenience access to timer status
    var status: TimerStatus {
        timer.status(for: activity)
    }
}

// MARK: - Timer Requests

struct CreateTimerRequest: Codable, Sendable {
    let activityId: String
    let enabled: Bool
    let push: Bool

    enum CodingKeys: String, CodingKey {
        case activityId = "activity_id"
        case enabled, push
    }

    init(activityId: String, enabled: Bool = true, push: Bool = false) {
        self.activityId = activityId
        self.enabled = enabled
        self.push = push
    }
}

struct UpdateTimerRequest: Codable, Sendable {
    let enabled: Bool?
    let push: Bool?

    init(enabled: Bool? = nil, push: Bool? = nil) {
        self.enabled = enabled
        self.push = push
    }
}
