import XCTest

/// Helper for multi-user testing scenarios
/// Supports both sequential user switching and dual-app instances
@MainActor
class MultiAppHelper {

    /// The primary app instance
    private(set) var primaryApp: XCUIApplication

    /// The secondary app instance (for dual-app testing)
    private(set) var secondaryApp: XCUIApplication?

    /// Login helper for primary app
    private(set) var primaryLogin: LoginHelper

    /// Login helper for secondary app
    private(set) var secondaryLogin: LoginHelper?

    /// Current user logged into primary app
    private(set) var primaryUser: TestUsers.User?

    /// Current user logged into secondary app
    private(set) var secondaryUser: TestUsers.User?

    init() {
        primaryApp = XCUIApplication()
        primaryApp.launchArguments = [TestConfig.uiTestingArg]
        primaryApp.launchEnvironment = [
            TestConfig.apiURLEnvKey: TestConfig.apiBaseURL
        ]
        primaryLogin = LoginHelper(app: primaryApp)
    }

    // MARK: - Primary App Operations

    /// Launch primary app and login as specified user
    @discardableResult
    func launchPrimaryApp(asUser user: TestUsers.User) -> Bool {
        primaryApp.launch()
        primaryUser = user
        return primaryLogin.loginAs(user)
    }

    /// Launch primary app without logging in
    func launchPrimaryApp() {
        primaryApp.launch()
    }

    // MARK: - Sequential User Switching

    /// Switch to a different user in the primary app
    /// Performs logout then login
    @discardableResult
    func switchPrimaryTo(_ user: TestUsers.User) -> Bool {
        let success = primaryLogin.switchTo(user)
        if success {
            primaryUser = user
        }
        return success
    }

    /// Logout from primary app
    func logoutPrimary() {
        primaryLogin.logout()
        primaryUser = nil
    }

    // MARK: - Dual App Operations

    /// Launch a secondary app instance for real-time sync testing
    /// Note: This requires the app to support multiple instances
    func launchSecondaryApp(asUser user: TestUsers.User) -> Bool {
        // Create secondary app instance with different bundle ID suffix
        // This is a simplified version - real dual-app testing may need special setup
        secondaryApp = XCUIApplication()
        secondaryApp?.launchArguments = [
            TestConfig.uiTestingArg,
            "--secondary-instance"
        ]
        secondaryApp?.launchEnvironment = [
            TestConfig.apiURLEnvKey: TestConfig.apiBaseURL
        ]

        guard let secondary = secondaryApp else { return false }

        secondaryLogin = LoginHelper(app: secondary)
        secondary.launch()

        let success = secondaryLogin?.loginAs(user) ?? false
        if success {
            secondaryUser = user
        }
        return success
    }

    /// Terminate secondary app
    func terminateSecondary() {
        secondaryApp?.terminate()
        secondaryApp = nil
        secondaryLogin = nil
        secondaryUser = nil
    }

    // MARK: - Sync Helpers

    /// Wait for SSE sync events to propagate between users
    /// - Parameter timeout: Maximum time to wait for sync
    func waitForSync(timeout: TimeInterval = 3) {
        // Allow time for SSE events to propagate
        Thread.sleep(forTimeInterval: timeout)
    }

    /// Refresh primary app view (pull to refresh or navigate away/back)
    func refreshPrimary() {
        // Attempt pull to refresh on main list
        let firstCell = primaryApp.cells.firstMatch
        if firstCell.exists {
            let start = firstCell.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: 0.5))
            let end = firstCell.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: 2.0))
            start.press(forDuration: 0.1, thenDragTo: end)
        }
    }

    /// Refresh secondary app view
    func refreshSecondary() {
        guard let secondary = secondaryApp else { return }

        let firstCell = secondary.cells.firstMatch
        if firstCell.exists {
            let start = firstCell.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: 0.5))
            let end = firstCell.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: 2.0))
            start.press(forDuration: 0.1, thenDragTo: end)
        }
    }

    // MARK: - State Verification

    /// Check if both users are logged in
    var bothUsersLoggedIn: Bool {
        primaryLogin.isLoggedIn && (secondaryLogin?.isLoggedIn ?? false)
    }

    /// Check if primary user is logged in
    var primaryLoggedIn: Bool {
        primaryLogin.isLoggedIn
    }

    /// Check if secondary user is logged in
    var secondaryLoggedIn: Bool {
        secondaryLogin?.isLoggedIn ?? false
    }

    // MARK: - Cleanup

    /// Terminate all app instances
    func cleanup() {
        primaryApp.terminate()
        secondaryApp?.terminate()
        primaryUser = nil
        secondaryUser = nil
        secondaryApp = nil
        secondaryLogin = nil
    }
}

// MARK: - Multi-User Test Actions

extension MultiAppHelper {

    /// Perform an action in primary app and verify result in secondary
    /// - Parameters:
    ///   - primaryAction: Action to perform in primary app
    ///   - verification: Verification to run in secondary app
    ///   - syncTimeout: Time to wait for sync
    /// - Returns: True if verification passes
    func performAndVerify(
        primaryAction: () -> Void,
        verification: (XCUIApplication) -> Bool,
        syncTimeout: TimeInterval = 3
    ) -> Bool {
        // Perform action in primary
        primaryAction()

        // Wait for sync
        waitForSync(timeout: syncTimeout)

        // Refresh secondary
        refreshSecondary()

        // Verify in secondary
        guard let secondary = secondaryApp else { return false }
        return verification(secondary)
    }

    /// Navigate to a tab in primary app by label name
    /// Use "Guilds", "Events", "Discover", or "Profile"
    func navigatePrimaryToTab(_ tabLabel: String) {
        let tabButton = primaryApp.tabBars.buttons[tabLabel]
        if tabButton.waitForExistence(timeout: TestConfig.shortTimeout) {
            tabButton.tap()
        }
    }

    /// Navigate to a tab in secondary app
    func navigateSecondaryToTab(_ tabLabel: String) {
        guard let secondary = secondaryApp else { return }
        let tabButton = secondary.tabBars.buttons[tabLabel]
        if tabButton.waitForExistence(timeout: TestConfig.shortTimeout) {
            tabButton.tap()
        }
    }

    // MARK: - Tab Label Constants
    enum Tab {
        static let guilds = "Guilds"
        static let events = "Events"
        static let discover = "Discover"
        static let profile = "Profile"
    }
}
