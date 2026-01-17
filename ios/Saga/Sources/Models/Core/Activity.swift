import Foundation

// MARK: - Activity

/// An activity type that can be tracked with timers (e.g., "Hung out", "Called")
struct Activity: Codable, Sendable, Identifiable, Hashable {
    let id: String
    let name: String
    let icon: String
    let warn: Int      // Seconds until warning threshold
    let critical: Int  // Seconds until critical threshold
    let createdOn: Date?
    let updatedOn: Date?

    enum CodingKeys: String, CodingKey {
        case id, name, icon, warn, critical
        case createdOn = "created_on"
        case updatedOn = "updated_on"
    }

    /// Warning threshold as TimeInterval
    var warnInterval: TimeInterval {
        TimeInterval(warn)
    }

    /// Critical threshold as TimeInterval
    var criticalInterval: TimeInterval {
        TimeInterval(critical)
    }

    /// Human-readable warning duration
    var warnDescription: String {
        formatDuration(warn)
    }

    /// Human-readable critical duration
    var criticalDescription: String {
        formatDuration(critical)
    }

    private func formatDuration(_ seconds: Int) -> String {
        let days = seconds / 86400
        if days >= 7 {
            let weeks = days / 7
            return weeks == 1 ? "1 week" : "\(weeks) weeks"
        } else if days >= 1 {
            return days == 1 ? "1 day" : "\(days) days"
        } else {
            let hours = seconds / 3600
            return hours == 1 ? "1 hour" : "\(hours) hours"
        }
    }
}

// MARK: - Activity Requests

struct CreateActivityRequest: Codable, Sendable {
    let name: String
    let icon: String
    let warn: Int
    let critical: Int

    init(name: String, icon: String, warn: Int, critical: Int) {
        self.name = name
        self.icon = icon
        self.warn = warn
        self.critical = critical
    }

    /// Create with duration components
    static func create(
        name: String,
        icon: String,
        warnDays: Int,
        criticalDays: Int
    ) -> CreateActivityRequest {
        CreateActivityRequest(
            name: name,
            icon: icon,
            warn: warnDays * 86400,
            critical: criticalDays * 86400
        )
    }
}

struct UpdateActivityRequest: Codable, Sendable {
    let name: String?
    let icon: String?
    let warn: Int?
    let critical: Int?

    init(name: String? = nil, icon: String? = nil, warn: Int? = nil, critical: Int? = nil) {
        self.name = name
        self.icon = icon
        self.warn = warn
        self.critical = critical
    }
}

// MARK: - Default Activities

extension Activity {
    /// Suggested default activities for new guilds
    static let defaults: [(name: String, icon: String, warnDays: Int, criticalDays: Int)] = [
        ("Hung out", "person.2.fill", 14, 30),
        ("Called", "phone.fill", 7, 14),
        ("Texted", "message.fill", 3, 7),
        ("Checked in", "hand.wave.fill", 30, 60)
    ]
}
