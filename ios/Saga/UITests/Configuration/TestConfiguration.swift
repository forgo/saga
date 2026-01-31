import Foundation

/// Configuration constants for UI tests
enum TestConfig {
    /// API base URL for testing
    static let apiBaseURL = "http://localhost:8080/v1"

    // MARK: - Timeouts

    /// Default wait timeout for element existence
    static let defaultTimeout: TimeInterval = 10

    /// Longer timeout for network operations
    static let longTimeout: TimeInterval = 30

    /// Short timeout for quick checks
    static let shortTimeout: TimeInterval = 3

    // MARK: - Delays

    /// Delay for network operations to complete
    static let networkDelay: TimeInterval = 0.5

    /// Delay for animations to complete
    static let animationDelay: TimeInterval = 0.3

    /// Delay for SSE sync between users
    static let sseDelay: TimeInterval = 2.0

    // MARK: - Launch Arguments

    /// Enable UI testing mode
    static let uiTestingArg = "--uitesting"

    /// Enable demo mode (auto-login)
    static let demoArg = "--demo"

    /// Mark as secondary instance (for multi-app testing)
    static let secondaryInstanceArg = "--secondary-instance"

    // MARK: - Environment Keys

    /// Override API base URL
    static let apiURLEnvKey = "API_BASE_URL"

    /// Test user email
    static let testUserEmailKey = "TEST_USER_EMAIL"

    /// Test user password
    static let testUserPasswordKey = "TEST_USER_PASSWORD"
}
