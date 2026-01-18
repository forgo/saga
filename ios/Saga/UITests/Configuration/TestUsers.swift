import Foundation

/// Test user credentials from API seed data
enum TestUsers {

    /// A test user with credentials
    struct User {
        let email: String
        let password: String
        let displayName: String

        /// Primary demo user with guilds and data
        static let demo = User(
            email: "demo@forgo.software",
            password: "password123",
            displayName: "Demo User"
        )

        /// Secondary user for multi-user testing
        static let second = User(
            email: "second@forgo.software",
            password: "password123",
            displayName: "Second User"
        )

        /// User with pending invites for testing
        static let pending = User(
            email: "pending@forgo.software",
            password: "password123",
            displayName: "Pending User"
        )
    }

    /// All available test users
    static let all: [User] = [.demo, .second, .pending]
}
