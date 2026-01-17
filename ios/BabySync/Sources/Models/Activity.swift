import Foundation

struct Activity: Codable, Identifiable, Equatable, Sendable {
    let id: String
    let familyId: String
    var name: String
    var icon: String
    var warn: TimeInterval
    var critical: TimeInterval
    let createdOn: Date
    let updatedOn: Date

    enum CodingKeys: String, CodingKey {
        case id, name, icon, warn, critical
        case familyId = "family_id"
        case createdOn = "created_on"
        case updatedOn = "updated_on"
    }

    var warnFormatted: String {
        formatDuration(warn)
    }

    var criticalFormatted: String {
        formatDuration(critical)
    }

    private func formatDuration(_ seconds: TimeInterval) -> String {
        let hours = Int(seconds) / 3600
        let minutes = (Int(seconds) % 3600) / 60
        if hours > 0 {
            return "\(hours)h \(minutes)m"
        }
        return "\(minutes)m"
    }
}

struct CreateActivityRequest: Codable, Sendable {
    let name: String
    let icon: String
    let warn: TimeInterval
    let critical: TimeInterval
}

struct UpdateActivityRequest: Codable, Sendable {
    var name: String?
    var icon: String?
    var warn: TimeInterval?
    var critical: TimeInterval?
}
