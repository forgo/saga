import XCTest

/// End-to-end UI tests for Saga app
/// Requires local API running with seed data: `make dev`
final class SagaUITests: XCTestCase {
    var app: XCUIApplication!

    override func setUpWithError() throws {
        continueAfterFailure = false
        app = XCUIApplication()

        // Enable UI testing mode
        app.launchArguments = ["--uitesting"]

        // Point to local API
        app.launchEnvironment = [
            "API_BASE_URL": "http://localhost:8080/v1"
        ]
    }

    override func tearDownWithError() throws {
        app = nil
    }

    // MARK: - Login Flow Tests

    func testLoginScreenAppears() throws {
        app.launch()

        // Verify login screen elements are present
        XCTAssertTrue(app.textFields["login_email_field"].waitForExistence(timeout: 5))
        XCTAssertTrue(app.secureTextFields["login_password_field"].exists)
        XCTAssertTrue(app.buttons["login_submit_button"].exists)
    }

    func testLoginWithDemoCredentials() throws {
        app.launch()

        let emailField = app.textFields["login_email_field"]
        let passwordField = app.secureTextFields["login_password_field"]
        let loginButton = app.buttons["login_submit_button"]

        // Wait for login screen
        XCTAssertTrue(emailField.waitForExistence(timeout: 5))

        // Enter demo credentials
        emailField.tap()
        emailField.typeText("demo@forgo.software")

        passwordField.tap()
        passwordField.typeText("password123")

        // Submit login
        loginButton.tap()

        // Verify we're logged in - guild list should appear
        XCTAssertTrue(app.navigationBars["Guilds"].waitForExistence(timeout: 10))
    }

    func testLoginWithInvalidCredentials() throws {
        app.launch()

        let emailField = app.textFields["login_email_field"]
        let passwordField = app.secureTextFields["login_password_field"]
        let loginButton = app.buttons["login_submit_button"]

        XCTAssertTrue(emailField.waitForExistence(timeout: 5))

        // Enter invalid credentials
        emailField.tap()
        emailField.typeText("invalid@example.com")

        passwordField.tap()
        passwordField.typeText("wrongpassword")

        loginButton.tap()

        // Verify error message appears
        XCTAssertTrue(app.staticTexts["login_error_message"].waitForExistence(timeout: 5))
    }

    // MARK: - Demo Mode Tests

    func testDemoModeAutoLogin() throws {
        // Enable demo mode for auto-login
        app.launchArguments.append("--demo")
        app.launch()

        // Should skip login and go directly to guild list
        XCTAssertTrue(app.navigationBars["Guilds"].waitForExistence(timeout: 10))
    }

    // MARK: - Guild List Tests

    func testGuildListDisplaysGuilds() throws {
        // Use demo mode for quick access
        app.launchArguments.append("--demo")
        app.launch()

        // Wait for guild list
        let guildList = app.collectionViews["guild_list"]
        XCTAssertTrue(guildList.waitForExistence(timeout: 10))

        // Verify at least one guild exists (from seed data)
        let firstGuild = guildList.cells.firstMatch
        XCTAssertTrue(firstGuild.exists)
    }

    func testCreateGuildButton() throws {
        app.launchArguments.append("--demo")
        app.launch()

        // Wait for guild list to load
        XCTAssertTrue(app.navigationBars["Guilds"].waitForExistence(timeout: 10))

        // Tap create guild button
        let createButton = app.buttons["create_guild_button"]
        XCTAssertTrue(createButton.exists)
    }

    // MARK: - Navigation Tests

    func testTabBarNavigation() throws {
        app.launchArguments.append("--demo")
        app.launch()

        // Wait for main interface
        XCTAssertTrue(app.tabBars.firstMatch.waitForExistence(timeout: 10))

        // Verify tab bar items exist
        let tabBar = app.tabBars.firstMatch
        XCTAssertTrue(tabBar.buttons["Guilds"].exists)
        XCTAssertTrue(tabBar.buttons["Profile"].exists)
    }
}
