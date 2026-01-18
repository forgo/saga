import XCTest

/// Helper for authentication operations in UI tests
@MainActor
class LoginHelper {
    let app: XCUIApplication

    init(app: XCUIApplication) {
        self.app = app
    }

    // MARK: - Login Operations

    /// Login with specified credentials and wait for main screen
    @discardableResult
    func login(email: String, password: String) -> Bool {
        let emailField = app.textFields[AccessibilityID.Auth.emailField]
        let passwordField = app.secureTextFields[AccessibilityID.Auth.passwordField]
        let submitButton = app.buttons[AccessibilityID.Auth.submitButton]

        guard emailField.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            return false
        }

        // Enter credentials
        emailField.tap()
        emailField.typeText(email)

        passwordField.tap()
        passwordField.typeText(password)

        // Dismiss keyboard and submit
        app.dismissKeyboardIfPresent()
        submitButton.tap()

        // Wait for guild list to appear (indicates successful login)
        let guildList = app.collectionViews[AccessibilityID.Guild.list]
        return guildList.waitForExistence(timeout: TestConfig.longTimeout)
    }

    /// Login with a test user
    @discardableResult
    func loginAs(_ user: TestUsers.User) -> Bool {
        login(email: user.email, password: user.password)
    }

    /// Login with demo user
    @discardableResult
    func loginWithDemoUser() -> Bool {
        loginAs(.demo)
    }

    // MARK: - Logout Operations

    /// Logout current user
    func logout() {
        // Navigate to profile tab (use button label, not accessibility ID)
        let profileTab = app.tabBars.buttons["Profile"]
        guard profileTab.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            return
        }
        profileTab.tap()

        // Scroll to find logout button (it's at the bottom of the list)
        let list = app.collectionViews.firstMatch.exists ? app.collectionViews.firstMatch : app.tables.firstMatch
        let logoutButton = app.buttons[AccessibilityID.Profile.logoutButton]

        var attempts = 0
        while !logoutButton.exists && attempts < 5 {
            list.swipeUp()
            // Brief pause for scroll to complete
            Thread.sleep(forTimeInterval: 0.3)
            attempts += 1
        }

        guard logoutButton.exists else {
            return
        }

        // Make sure it's hittable
        attempts = 0
        while !logoutButton.isHittable && attempts < 3 {
            list.swipeUp()
            Thread.sleep(forTimeInterval: 0.3)
            attempts += 1
        }

        guard logoutButton.isHittable else {
            return
        }

        logoutButton.tap()

        // Confirm logout if alert appears (try multiple button names)
        let signOutConfirm = app.alerts.buttons["Sign Out"]
        let logoutConfirm = app.alerts.buttons["Logout"]
        if signOutConfirm.waitForExistence(timeout: TestConfig.shortTimeout) {
            signOutConfirm.tap()
        } else if logoutConfirm.exists {
            logoutConfirm.tap()
        }

        // Wait for auth screen
        let emailField = app.textFields[AccessibilityID.Auth.emailField]
        _ = emailField.waitForExistence(timeout: TestConfig.defaultTimeout)
    }

    // MARK: - User Switching

    /// Switch to a different user (logout and login)
    @discardableResult
    func switchTo(_ user: TestUsers.User) -> Bool {
        logout()
        return loginAs(user)
    }

    // MARK: - State Checks

    /// Check if user is logged in
    var isLoggedIn: Bool {
        // Check for tab bar (indicates logged in state)
        app.tabBars.firstMatch.exists
    }

    /// Check if on auth screen
    var isOnAuthScreen: Bool {
        app.textFields[AccessibilityID.Auth.emailField].exists
    }

    /// Check if login error is displayed
    var hasLoginError: Bool {
        app.staticTexts[AccessibilityID.Auth.errorMessage].exists
    }

    /// Get the error message text
    var errorMessage: String? {
        let errorLabel = app.staticTexts[AccessibilityID.Auth.errorMessage]
        return errorLabel.exists ? errorLabel.label : nil
    }
}
