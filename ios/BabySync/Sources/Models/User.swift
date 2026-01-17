import Foundation

struct User: Codable, Identifiable, Equatable, Sendable {
    let id: String
    let email: String
    var username: String?
    var firstname: String?
    var lastname: String?
    let emailVerified: Bool
    let createdOn: Date
    let updatedOn: Date

    enum CodingKeys: String, CodingKey {
        case id, email, username, firstname, lastname
        case emailVerified = "email_verified"
        case createdOn = "created_on"
        case updatedOn = "updated_on"
    }

    var displayName: String {
        if let first = firstname, let last = lastname {
            return "\(first) \(last)"
        }
        return username ?? email
    }
}

struct Identity: Codable, Identifiable, Equatable, Sendable {
    let id: String
    let provider: AuthProvider
    let providerUserId: String
    var providerEmail: String?
    let emailVerifiedByProvider: Bool
    let createdOn: Date

    enum CodingKeys: String, CodingKey {
        case id, provider
        case providerUserId = "provider_user_id"
        case providerEmail = "provider_email"
        case emailVerifiedByProvider = "email_verified_by_provider"
        case createdOn = "created_on"
    }
}

enum AuthProvider: String, Codable, Sendable {
    case google
    case apple
}

struct Passkey: Codable, Identifiable, Equatable, Sendable {
    let id: String
    let name: String
    let createdOn: Date
    var lastUsedOn: Date?

    enum CodingKeys: String, CodingKey {
        case id, name
        case createdOn = "created_on"
        case lastUsedOn = "last_used_on"
    }
}

struct UserWithIdentities: Codable, Sendable {
    let user: User
    let identities: [Identity]
    let passkeys: [Passkey]
}
