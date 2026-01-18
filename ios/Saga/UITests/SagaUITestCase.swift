import XCTest

/// Base test case class for all Saga UI tests
/// Provides common setup, teardown, and helper methods
class SagaUITestCase: XCTestCase {

    /// The main application under test
    var app: XCUIApplication!

    /// Helper for login operations
    var loginHelper: LoginHelper!

    // MARK: - Setup & Teardown

    override func setUpWithError() throws {
        try super.setUpWithError()

        // Stop immediately when a failure occurs
        continueAfterFailure = false

        // Initialize app with testing configuration
        app = XCUIApplication()
        app.launchArguments = [TestConfig.uiTestingArg]
        app.launchEnvironment = [
            TestConfig.apiURLEnvKey: TestConfig.apiBaseURL
        ]

        // Initialize helpers
        loginHelper = LoginHelper(app: app)
    }

    override func tearDownWithError() throws {
        app = nil
        loginHelper = nil
        try super.tearDownWithError()
    }

    // MARK: - Launch Helpers

    /// Launch the app normally (shows auth screen)
    func launchApp() {
        app.launch()
    }

    /// Launch app and login with demo user
    @discardableResult
    func launchAndLoginWithDemoUser() -> GuildListPage {
        app.launch()

        let authPage = AuthPage(app: app)
        XCTAssertTrue(authPage.isDisplayed(), "Auth screen should be displayed")

        return authPage.loginAs(.demo)
    }

    /// Launch app and login with specified user
    @discardableResult
    func launchAndLogin(as user: TestUsers.User) -> GuildListPage {
        app.launch()

        let authPage = AuthPage(app: app)
        XCTAssertTrue(authPage.isDisplayed(), "Auth screen should be displayed")

        return authPage.loginAs(user)
    }

    /// Launch app in demo mode (bypasses login)
    @discardableResult
    func launchInDemoMode() -> GuildListPage {
        app.launchArguments.append(TestConfig.demoArg)
        app.launch()

        let guildList = GuildListPage(app: app)
        XCTAssertTrue(guildList.isDisplayed(), "Guild list should be displayed in demo mode")

        return guildList
    }

    // MARK: - Page Object Factories

    /// Get the auth page
    var authPage: AuthPage {
        AuthPage(app: app)
    }

    /// Get the tab bar page
    var tabBar: TabBarPage {
        TabBarPage(app: app)
    }

    /// Get the guild list page
    var guildListPage: GuildListPage {
        GuildListPage(app: app)
    }

    /// Get the event list page
    var eventListPage: EventListPage {
        EventListPage(app: app)
    }

    // MARK: - Common Assertions

    /// Assert user is logged in (tab bar visible)
    func assertLoggedIn() {
        XCTAssertTrue(tabBar.isDisplayed(), "User should be logged in (tab bar visible)")
    }

    /// Assert user is on auth screen
    func assertOnAuthScreen() {
        XCTAssertTrue(authPage.isDisplayed(), "User should be on auth screen")
    }

    /// Assert guild list is displayed
    func assertGuildListDisplayed() {
        XCTAssertTrue(guildListPage.isDisplayed(), "Guild list should be displayed")
    }

    // MARK: - Screenshot Helpers

    /// Take a screenshot and attach to test results
    func takeScreenshot(name: String) {
        let screenshot = XCUIScreen.main.screenshot()
        let attachment = XCTAttachment(screenshot: screenshot)
        attachment.name = name
        attachment.lifetime = .keepAlways
        add(attachment)
    }

    /// Take screenshot on failure
    func screenshotOnFailure() {
        if testRun?.hasSucceeded == false {
            takeScreenshot(name: "Failure - \(name)")
        }
    }
}
