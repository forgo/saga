import XCTest
@testable import Saga

/// Tests for the AuthService authentication management
final class AuthServiceTests: XCTestCase {

    // MARK: - Initial State Tests

    func testAuthService_InitialState_IsNotAuthenticated() {
        // Create a fresh instance without going through shared singleton
        let authService = AuthService()

        XCTAssertFalse(authService.isAuthenticated)
        XCTAssertNil(authService.currentUser)
        XCTAssertTrue(authService.identities.isEmpty)
        XCTAssertTrue(authService.passkeys.isEmpty)
        XCTAssertFalse(authService.isLoading)
        XCTAssertNil(authService.error)
    }

    // MARK: - Provider Linking Tests

    func testHasLinkedProvider_ReturnsFalse_WhenNoIdentities() {
        let authService = AuthService()

        XCTAssertFalse(authService.hasLinkedProvider(.google))
        XCTAssertFalse(authService.hasLinkedProvider(.apple))
    }

    func testIdentityForProvider_ReturnsNil_WhenNoIdentities() {
        let authService = AuthService()

        XCTAssertNil(authService.identity(for: .google))
        XCTAssertNil(authService.identity(for: .apple))
    }

    // MARK: - User Model Tests

    func testUser_DisplayName_WithFirstAndLastName() throws {
        let json = """
        {
            "id": "user:123",
            "email": "test@example.com",
            "firstname": "John",
            "lastname": "Doe"
        }
        """
        let user = try JSONDecoder().decode(User.self, from: json.data(using: .utf8)!)

        XCTAssertEqual(user.displayName, "John Doe")
    }

    func testUser_DisplayName_WithOnlyFirstName() throws {
        let json = """
        {
            "id": "user:123",
            "email": "test@example.com",
            "firstname": "John"
        }
        """
        let user = try JSONDecoder().decode(User.self, from: json.data(using: .utf8)!)

        XCTAssertEqual(user.displayName, "John")
    }

    func testUser_DisplayName_FallsBackToEmail() throws {
        let json = """
        {
            "id": "user:123",
            "email": "test@example.com"
        }
        """
        let user = try JSONDecoder().decode(User.self, from: json.data(using: .utf8)!)

        XCTAssertEqual(user.displayName, "test@example.com")
    }

    // MARK: - User Initials Tests

    func testUser_Initials_TwoNames() throws {
        let json = """
        {
            "id": "user:123",
            "email": "test@example.com",
            "firstname": "John",
            "lastname": "Doe"
        }
        """
        let user = try JSONDecoder().decode(User.self, from: json.data(using: .utf8)!)

        XCTAssertEqual(user.initials, "JD")
    }

    func testUser_Initials_SingleName() throws {
        let json = """
        {
            "id": "user:123",
            "email": "test@example.com",
            "firstname": "John"
        }
        """
        let user = try JSONDecoder().decode(User.self, from: json.data(using: .utf8)!)

        XCTAssertEqual(user.initials, "J")
    }

    func testUser_Initials_FromEmail() throws {
        let json = """
        {
            "id": "user:123",
            "email": "test@example.com"
        }
        """
        let user = try JSONDecoder().decode(User.self, from: json.data(using: .utf8)!)

        XCTAssertEqual(user.initials, "T")
    }

    // MARK: - Identity Model Tests

    func testIdentity_Provider_Decoding() throws {
        let json = """
        {
            "id": "identity:123",
            "provider": "google",
            "email": "user@gmail.com",
            "verified": true,
            "created_at": "2024-01-15T10:30:45Z"
        }
        """

        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601

        let identity = try decoder.decode(Identity.self, from: json.data(using: .utf8)!)

        XCTAssertEqual(identity.id, "identity:123")
        XCTAssertEqual(identity.provider, .google)
        XCTAssertEqual(identity.email, "user@gmail.com")
        XCTAssertTrue(identity.verified)
    }

    func testIdentityProvider_AllCases() {
        XCTAssertEqual(IdentityProvider.google.rawValue, "google")
        XCTAssertEqual(IdentityProvider.apple.rawValue, "apple")
        XCTAssertEqual(IdentityProvider.email.rawValue, "email")
    }

    func testIdentityProvider_DisplayName() {
        XCTAssertEqual(IdentityProvider.google.displayName, "Google")
        XCTAssertEqual(IdentityProvider.apple.displayName, "Apple")
        XCTAssertEqual(IdentityProvider.email.displayName, "Email")
    }

    // MARK: - Passkey Model Tests

    func testPasskey_Decoding() throws {
        let json = """
        {
            "id": "passkey:123",
            "credential_id": "cred_abc123",
            "display_name": "iPhone 15 Pro",
            "created_at": "2024-01-15T10:30:45Z",
            "last_used_at": "2024-01-15T12:00:00Z"
        }
        """

        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601

        let passkey = try decoder.decode(Passkey.self, from: json.data(using: .utf8)!)

        XCTAssertEqual(passkey.id, "passkey:123")
        XCTAssertEqual(passkey.name, "iPhone 15 Pro")
        XCTAssertNotNil(passkey.lastUsedAt)
    }

    func testPasskey_Decoding_WithoutOptionalFields() throws {
        let json = """
        {
            "id": "passkey:456",
            "credential_id": "cred_def456"
        }
        """

        let passkey = try JSONDecoder().decode(Passkey.self, from: json.data(using: .utf8)!)

        XCTAssertEqual(passkey.id, "passkey:456")
        XCTAssertEqual(passkey.name, "Passkey")  // Default fallback
        XCTAssertNil(passkey.lastUsedAt)
    }

    // MARK: - User Decoding Tests

    func testUser_Decoding_MinimalFields() throws {
        let json = """
        {
            "id": "user:123"
        }
        """

        let user = try JSONDecoder().decode(User.self, from: json.data(using: .utf8)!)

        XCTAssertEqual(user.id, "user:123")
        XCTAssertNil(user.email)
        XCTAssertNil(user.firstname)
        XCTAssertNil(user.lastname)
    }

    func testUser_Decoding_AllFields() throws {
        let json = """
        {
            "id": "user:123",
            "email": "test@example.com",
            "firstname": "John",
            "lastname": "Doe",
            "created_at": "2024-01-15T10:30:45Z",
            "updated_at": "2024-01-15T12:00:00Z"
        }
        """

        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601

        let user = try decoder.decode(User.self, from: json.data(using: .utf8)!)

        XCTAssertEqual(user.id, "user:123")
        XCTAssertEqual(user.email, "test@example.com")
        XCTAssertEqual(user.firstname, "John")
        XCTAssertEqual(user.lastname, "Doe")
        XCTAssertNotNil(user.createdAt)
        XCTAssertNotNil(user.updatedAt)
    }

    // MARK: - UserWithIdentities Tests

    func testUserWithIdentities_Decoding() throws {
        let json = """
        {
            "user": {
                "id": "user:123",
                "email": "test@example.com"
            },
            "identities": [
                {
                    "id": "identity:1",
                    "provider": "google",
                    "verified": true
                }
            ],
            "passkeys": [
                {
                    "id": "passkey:1",
                    "credential_id": "cred_1"
                }
            ]
        }
        """

        let response = try JSONDecoder().decode(UserWithIdentities.self, from: json.data(using: .utf8)!)

        XCTAssertEqual(response.user.id, "user:123")
        XCTAssertEqual(response.identities.count, 1)
        XCTAssertEqual(response.passkeys?.count, 1)
    }

    // MARK: - Demo Credentials Tests

    #if DEBUG
    func testDemoCredentials_HaveExpectedValues() {
        // These match the seed data in the API
        XCTAssertEqual(AuthService.DemoCredentials.email, "demo@forgo.software")
        XCTAssertEqual(AuthService.DemoCredentials.password, "password123")
    }
    #endif

    // MARK: - State Observation Tests

    func testAuthService_IdentitiesStartEmpty() {
        let authService = AuthService()
        XCTAssertTrue(authService.identities.isEmpty)
    }

    func testAuthService_PasskeysStartEmpty() {
        let authService = AuthService()
        XCTAssertTrue(authService.passkeys.isEmpty)
    }

    func testAuthService_CurrentUserStartsNil() {
        let authService = AuthService()
        XCTAssertNil(authService.currentUser)
    }

    func testAuthService_IsLoadingStartsFalse() {
        let authService = AuthService()
        XCTAssertFalse(authService.isLoading)
    }

    func testAuthService_ErrorStartsNil() {
        let authService = AuthService()
        XCTAssertNil(authService.error)
    }

    // MARK: - User Hashable/Equatable Tests

    func testUser_Hashable() throws {
        let json1 = """
        {"id": "user:123", "email": "test@example.com"}
        """
        let json2 = """
        {"id": "user:123", "email": "test@example.com"}
        """
        let json3 = """
        {"id": "user:456", "email": "other@example.com"}
        """

        let user1 = try JSONDecoder().decode(User.self, from: json1.data(using: .utf8)!)
        let user2 = try JSONDecoder().decode(User.self, from: json2.data(using: .utf8)!)
        let user3 = try JSONDecoder().decode(User.self, from: json3.data(using: .utf8)!)

        // Same users should be equal
        XCTAssertEqual(user1, user2)

        // Different users should not be equal
        XCTAssertNotEqual(user1, user3)

        // Test using in Set
        var userSet = Set<User>()
        userSet.insert(user1)
        userSet.insert(user2)  // Should not add duplicate
        userSet.insert(user3)

        XCTAssertEqual(userSet.count, 2)
    }

    // MARK: - Identity Hashable Tests

    func testIdentity_Hashable() throws {
        let json1 = """
        {"id": "identity:1", "provider": "google", "verified": true}
        """
        let json2 = """
        {"id": "identity:1", "provider": "google", "verified": true}
        """

        let identity1 = try JSONDecoder().decode(Identity.self, from: json1.data(using: .utf8)!)
        let identity2 = try JSONDecoder().decode(Identity.self, from: json2.data(using: .utf8)!)

        XCTAssertEqual(identity1, identity2)
    }

    // MARK: - Passkey Hashable Tests

    func testPasskey_Hashable() throws {
        let json1 = """
        {"id": "passkey:1", "credential_id": "cred_1"}
        """
        let json2 = """
        {"id": "passkey:1", "credential_id": "cred_1"}
        """

        let passkey1 = try JSONDecoder().decode(Passkey.self, from: json1.data(using: .utf8)!)
        let passkey2 = try JSONDecoder().decode(Passkey.self, from: json2.data(using: .utf8)!)

        XCTAssertEqual(passkey1, passkey2)
    }
}
