import Foundation

// MARK: - Profile

/// A user's profile with privacy settings
struct Profile: Codable, Identifiable, Hashable, Sendable {
    let id: String
    let userId: String
    let displayName: String?
    let bio: String?
    let avatarUrl: String?
    let location: String?
    let locationLat: Double?
    let locationLng: Double?
    let visibility: ProfileVisibility
    let showDistance: Bool
    let showOnline: Bool
    let createdOn: Date
    let updatedOn: Date

    enum CodingKeys: String, CodingKey {
        case id
        case userId = "user_id"
        case displayName = "display_name"
        case bio
        case avatarUrl = "avatar_url"
        case location
        case locationLat = "location_lat"
        case locationLng = "location_lng"
        case visibility
        case showDistance = "show_distance"
        case showOnline = "show_online"
        case createdOn = "created_on"
        case updatedOn = "updated_on"
    }

    /// Display name or fallback
    var name: String {
        displayName ?? "Anonymous"
    }

    /// Whether the profile has location set
    var hasLocation: Bool {
        locationLat != nil && locationLng != nil
    }
}

// MARK: - Profile Visibility

enum ProfileVisibility: String, Codable, Sendable, CaseIterable {
    case `public`
    case guildsOnly = "guilds_only"
    case `private`

    var displayName: String {
        switch self {
        case .public: return "Public"
        case .guildsOnly: return "Guild Members Only"
        case .private: return "Private"
        }
    }

    var description: String {
        switch self {
        case .public: return "Anyone can see your profile"
        case .guildsOnly: return "Only members of your guilds can see your profile"
        case .private: return "Only you can see your profile"
        }
    }

    var iconName: String {
        switch self {
        case .public: return "globe"
        case .guildsOnly: return "person.3"
        case .private: return "lock"
        }
    }
}

// MARK: - Public Profile

/// Privacy-filtered profile for other users to see
struct PublicProfile: Codable, Identifiable, Hashable, Sendable {
    let userId: String
    let displayName: String
    let bio: String?
    let avatarUrl: String?
    let distanceKm: Double?
    let isOnline: Bool?
    let mutualGuilds: [String]?

    var id: String { userId }

    enum CodingKeys: String, CodingKey {
        case userId = "user_id"
        case displayName = "display_name"
        case bio
        case avatarUrl = "avatar_url"
        case distanceKm = "distance_km"
        case isOnline = "is_online"
        case mutualGuilds = "mutual_guilds"
    }

    /// Initials for avatar placeholder
    var initials: String {
        let parts = displayName.split(separator: " ")
        if parts.count >= 2 {
            return String(parts[0].prefix(1) + parts[1].prefix(1)).uppercased()
        }
        return String(displayName.prefix(2)).uppercased()
    }

    /// Formatted distance
    var formattedDistance: String? {
        guard let km = distanceKm else { return nil }
        if km < 1 {
            return "< 1 km away"
        }
        return String(format: "%.1f km away", km)
    }
}

// MARK: - Nearby Profile

/// A profile found through discovery with distance info
struct NearbyProfile: Codable, Sendable {
    let profile: PublicProfile
    let distanceKm: Double
    let compatibilityScore: Double?

    enum CodingKeys: String, CodingKey {
        case profile
        case distanceKm = "distance_km"
        case compatibilityScore = "compatibility_score"
    }
}

// MARK: - Request Types

struct UpdateProfileRequest: Codable, Sendable {
    var displayName: String?
    var bio: String?
    var avatarUrl: String?
    var location: String?
    var locationLat: Double?
    var locationLng: Double?
    var visibility: ProfileVisibility?
    var showDistance: Bool?
    var showOnline: Bool?

    enum CodingKeys: String, CodingKey {
        case displayName = "display_name"
        case bio
        case avatarUrl = "avatar_url"
        case location
        case locationLat = "location_lat"
        case locationLng = "location_lng"
        case visibility
        case showDistance = "show_distance"
        case showOnline = "show_online"
    }
}
