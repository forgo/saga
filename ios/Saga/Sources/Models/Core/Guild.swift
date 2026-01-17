import Foundation
import SwiftUI

// MARK: - Guild

/// A guild represents a group/circle of people
struct Guild: Codable, Sendable, Identifiable, Hashable {
    let id: String
    let name: String
    let description: String?
    let icon: String?
    let color: String?
    let createdOn: Date?
    let updatedOn: Date?

    enum CodingKeys: String, CodingKey {
        case id, name, description, icon, color
        case createdOn = "created_on"
        case updatedOn = "updated_on"
    }

    /// SF Symbol name for the guild icon
    var iconName: String {
        icon ?? "person.3.fill"
    }

    /// SwiftUI color from hex string
    var displayColor: Color {
        guard let hex = color else { return .blue }
        return Color(hex: hex) ?? .blue
    }
}

// MARK: - Guild Data (Full details)

/// Guild with all related data (members, people, activities)
struct GuildData: Codable, Sendable {
    let guild: Guild
    let members: [Member]
    let people: [Person]
    let activities: [Activity]
}

// MARK: - Member

/// A user who is a member of a guild
struct Member: Codable, Sendable, Identifiable, Hashable {
    let id: String
    let name: String
    let email: String
    let user: String?
}

// MARK: - Guild Requests

struct CreateGuildRequest: Codable, Sendable {
    let name: String
    let description: String?
    let icon: String?
    let color: String?

    init(name: String, description: String? = nil, icon: String? = nil, color: String? = nil) {
        self.name = name
        self.description = description
        self.icon = icon
        self.color = color
    }
}

struct UpdateGuildRequest: Codable, Sendable {
    let name: String?
    let description: String?
    let icon: String?
    let color: String?

    init(name: String? = nil, description: String? = nil, icon: String? = nil, color: String? = nil) {
        self.name = name
        self.description = description
        self.icon = icon
        self.color = color
    }
}

struct MergeGuildsRequest: Codable, Sendable {
    let sourceGuildId: String

    enum CodingKeys: String, CodingKey {
        case sourceGuildId = "source_guild_id"
    }
}

// MARK: - Color Extension

extension Color {
    init?(hex: String) {
        var hexString = hex.trimmingCharacters(in: .whitespacesAndNewlines)
        if hexString.hasPrefix("#") {
            hexString.removeFirst()
        }

        guard hexString.count == 6 else { return nil }

        var rgbValue: UInt64 = 0
        Scanner(string: hexString).scanHexInt64(&rgbValue)

        let red = Double((rgbValue & 0xFF0000) >> 16) / 255.0
        let green = Double((rgbValue & 0x00FF00) >> 8) / 255.0
        let blue = Double(rgbValue & 0x0000FF) / 255.0

        self.init(red: red, green: green, blue: blue)
    }

    /// Convert Color to hex string
    var hexString: String? {
        #if canImport(UIKit)
        guard let components = UIColor(self).cgColor.components, components.count >= 3 else {
            return nil
        }
        let r = Int(components[0] * 255)
        let g = Int(components[1] * 255)
        let b = Int(components[2] * 255)
        return String(format: "#%02X%02X%02X", r, g, b)
        #else
        return nil
        #endif
    }
}
