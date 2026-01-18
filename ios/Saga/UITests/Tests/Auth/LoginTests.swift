import XCTest

/// Tests for user login functionality

final class LoginTests: SagaUITestCase {

    // MARK: - Auth Screen Tests

    func testAuthScreenAppearsOnLaunch() throws {
        launchApp()

        XCTAssertTrue(authPage.isDisplayed(), "Auth screen should be displayed on launch")
        XCTAssertTrue(authPage.emailField.exists, "Email field should exist")
        XCTAssertTrue(authPage.passwordField.exists, "Password field should exist")
        XCTAssertTrue(authPage.submitButton.exists, "Submit button should exist")
    }

    func testAuthScreenHasModePicker() throws {
        launchApp()

        XCTAssertTrue(authPage.modePicker.exists, "Mode picker should exist")
        XCTAssertTrue(authPage.isInLoginMode(), "Should default to login mode")
    }

    func testCanSwitchToRegisterMode() throws {
        launchApp()

        authPage.switchToRegister()

        XCTAssertTrue(authPage.isInRegisterMode(), "Should be in register mode")
        XCTAssertTrue(authPage.firstnameField.exists, "First name field should appear")
        XCTAssertTrue(authPage.lastnameField.exists, "Last name field should appear")
    }

    func testCanSwitchBackToLoginMode() throws {
        launchApp()

        authPage.switchToRegister()
        authPage.switchToLogin()

        XCTAssertTrue(authPage.isInLoginMode(), "Should be back in login mode")
        XCTAssertFalse(authPage.firstnameField.exists, "First name field should not exist")
    }

    // MARK: - Login Flow Tests

    func testLoginWithValidCredentials() throws {
        launchApp()

        let guildList = authPage.loginAs(.demo)

        XCTAssertTrue(guildList.isDisplayed(), "Guild list should be displayed after login")
        assertLoggedIn()
    }

    func testLoginWithInvalidCredentials() throws {
        launchApp()

        _ = authPage.login(email: "invalid@example.com", password: "wrongpassword")

        // Wait for error to appear
        XCTAssertTrue(
            authPage.errorMessage.waitForExistence(timeout: TestConfig.longTimeout),
            "Error message should appear for invalid credentials"
        )
        XCTAssertTrue(authPage.hasError(), "Should show error message")
    }

    func testLoginWithEmptyEmail() throws {
        launchApp()

        authPage.enterPassword("password123")

        // Submit button should be disabled with empty email
        XCTAssertFalse(authPage.isSubmitEnabled(), "Submit should be disabled with empty email")
    }

    func testLoginWithEmptyPassword() throws {
        launchApp()

        authPage.enterEmail("test@example.com")

        // Submit button should be disabled with empty password
        XCTAssertFalse(authPage.isSubmitEnabled(), "Submit should be disabled with empty password")
    }

    func testLoginWithShortPassword() throws {
        launchApp()

        authPage.enterEmail("test@example.com")
        authPage.enterPassword("short")

        // Submit button should be disabled with password < 8 chars
        XCTAssertFalse(authPage.isSubmitEnabled(), "Submit should be disabled with short password")
    }

    func testLoginWithInvalidEmailFormat() throws {
        launchApp()

        authPage.enterEmail("notanemail")
        authPage.enterPassword("password123")

        // Submit button should be disabled with invalid email
        XCTAssertFalse(authPage.isSubmitEnabled(), "Submit should be disabled with invalid email")
    }

    // MARK: - Demo Mode Tests

    func testDemoModeBypassesLogin() throws {
        let guildList = launchInDemoMode()

        XCTAssertTrue(guildList.isDisplayed(), "Guild list should be displayed in demo mode")
        assertLoggedIn()
    }

    // MARK: - Social Auth Button Tests

    func testGoogleSignInButtonExists() throws {
        launchApp()

        // Wait for auth screen to fully load
        XCTAssertTrue(authPage.isDisplayed(), "Auth screen should be displayed")
        XCTAssertTrue(
            authPage.googleButton.waitForExistence(timeout: TestConfig.defaultTimeout),
            "Google sign-in button should exist"
        )
    }

    func testPasskeyButtonExistsInLoginMode() throws {
        launchApp()

        XCTAssertTrue(authPage.isInLoginMode(), "Should be in login mode")
        XCTAssertTrue(authPage.passkeyButton.exists, "Passkey button should exist in login mode")
    }

    func testPasskeyButtonNotInRegisterMode() throws {
        launchApp()

        authPage.switchToRegister()

        XCTAssertFalse(authPage.passkeyButton.exists, "Passkey button should not exist in register mode")
    }
}
