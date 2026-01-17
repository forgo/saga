import Foundation

// MARK: - User

/// Authenticated user
struct User: Codable, Sendable, Identifiable, Hashable {
    let id: String
    let email: String?
    let firstname: String?
    let lastname: String?
    let createdAt: Date?
    let updatedAt: Date?

    enum CodingKeys: String, CodingKey {
        case id, email, firstname, lastname
        case createdAt = "created_at"
        case updatedAt = "updated_at"
    }

    /// User's display name
    var displayName: String {
        if let first = firstname, let last = lastname {
            return "\(first) \(last)"
        }
        return firstname ?? lastname ?? email ?? "User"
    }

    /// User's initials for avatar
    var initials: String {
        let first = firstname?.prefix(1) ?? ""
        let last = lastname?.prefix(1) ?? ""
        if !first.isEmpty || !last.isEmpty {
            return "\(first)\(last)".uppercased()
        }
        return email?.prefix(1).uppercased() ?? "?"
    }
}

// MARK: - User with Identities

/// User with linked authentication identities
struct UserWithIdentities: Codable, Sendable {
    let user: User
    let identities: [Identity]
    let passkeys: [Passkey]?
}

// MARK: - Identity

/// Authentication identity (email, OAuth provider, etc.)
struct Identity: Codable, Sendable, Identifiable, Hashable {
    let id: String
    let provider: IdentityProvider
    let providerUserId: String?
    let email: String?
    let verified: Bool
    let createdAt: Date?

    enum CodingKeys: String, CodingKey {
        case id, provider, email, verified
        case providerUserId = "provider_user_id"
        case createdAt = "created_at"
    }
}

/// Authentication provider types
enum IdentityProvider: String, Codable, Sendable {
    case email
    case google
    case apple

    var displayName: String {
        switch self {
        case .email: return "Email"
        case .google: return "Google"
        case .apple: return "Apple"
        }
    }

    var iconName: String {
        switch self {
        case .email: return "envelope.fill"
        case .google: return "g.circle.fill"
        case .apple: return "apple.logo"
        }
    }
}

// MARK: - Passkey

/// Registered passkey for WebAuthn authentication
struct Passkey: Codable, Sendable, Identifiable, Hashable {
    let id: String
    let credentialId: String
    let displayName: String?
    let lastUsedAt: Date?
    let createdAt: Date?

    enum CodingKeys: String, CodingKey {
        case id
        case credentialId = "credential_id"
        case displayName = "display_name"
        case lastUsedAt = "last_used_at"
        case createdAt = "created_at"
    }

    /// Passkey display name with fallback
    var name: String {
        displayName ?? "Passkey"
    }
}
