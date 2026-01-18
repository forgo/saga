import XCTest

/// Page object for the authentication screen
@MainActor
class AuthPage: BasePage {

    // MARK: - Elements

    /// Email input field
    var emailField: XCUIElement {
        app.textFields[AccessibilityID.Auth.emailField]
    }

    /// Password input field
    var passwordField: XCUIElement {
        app.secureTextFields[AccessibilityID.Auth.passwordField]
    }

    /// Submit button (Sign In / Create Account)
    var submitButton: XCUIElement {
        app.buttons[AccessibilityID.Auth.submitButton]
    }

    /// Error message label
    var errorMessage: XCUIElement {
        app.staticTexts[AccessibilityID.Auth.errorMessage]
    }

    /// Mode picker (Sign In / Create Account)
    var modePicker: XCUIElement {
        app.segmentedControls[AccessibilityID.Auth.modePicker]
    }

    /// Sign In tab
    var signInTab: XCUIElement {
        modePicker.buttons["Sign In"]
    }

    /// Create Account tab
    var createAccountTab: XCUIElement {
        modePicker.buttons["Create Account"]
    }

    /// First name field (registration only)
    var firstnameField: XCUIElement {
        app.textFields[AccessibilityID.Auth.firstnameField]
    }

    /// Last name field (registration only)
    var lastnameField: XCUIElement {
        app.textFields[AccessibilityID.Auth.lastnameField]
    }

    /// Sign in with Google button
    var googleButton: XCUIElement {
        app.buttons[AccessibilityID.Auth.googleButton]
    }

    /// Sign in with Apple button
    var appleButton: XCUIElement {
        app.buttons[AccessibilityID.Auth.appleButton]
    }

    /// Sign in with Passkey button
    var passkeyButton: XCUIElement {
        app.buttons[AccessibilityID.Auth.passkeyButton]
    }

    // MARK: - State Checks

    /// Check if auth screen is displayed
    func isDisplayed() -> Bool {
        emailField.waitForExistence(timeout: TestConfig.defaultTimeout)
    }

    /// Check if error message is shown
    func hasError() -> Bool {
        errorMessage.exists
    }

    /// Get error message text
    func getErrorText() -> String? {
        errorMessage.exists ? errorMessage.label : nil
    }

    /// Check if submit button is enabled
    func isSubmitEnabled() -> Bool {
        submitButton.isEnabled
    }

    /// Check if in login mode
    func isInLoginMode() -> Bool {
        signInTab.isSelected
    }

    /// Check if in register mode
    func isInRegisterMode() -> Bool {
        createAccountTab.isSelected
    }

    // MARK: - Actions

    /// Switch to login mode
    @discardableResult
    func switchToLogin() -> AuthPage {
        if !isInLoginMode() {
            signInTab.tap()
        }
        return self
    }

    /// Switch to register mode
    @discardableResult
    func switchToRegister() -> AuthPage {
        if !isInRegisterMode() {
            createAccountTab.tap()
        }
        return self
    }

    /// Enter email
    @discardableResult
    func enterEmail(_ email: String) -> AuthPage {
        emailField.tap()
        emailField.clearAndType(email)
        return self
    }

    /// Enter password
    @discardableResult
    func enterPassword(_ password: String) -> AuthPage {
        passwordField.tap()
        passwordField.typeText(password)
        return self
    }

    /// Enter first name (registration)
    @discardableResult
    func enterFirstname(_ name: String) -> AuthPage {
        firstnameField.tap()
        firstnameField.typeText(name)
        return self
    }

    /// Enter last name (registration)
    @discardableResult
    func enterLastname(_ name: String) -> AuthPage {
        lastnameField.tap()
        lastnameField.typeText(name)
        return self
    }

    /// Tap submit button
    @discardableResult
    func tapSubmit() -> AuthPage {
        app.dismissKeyboardIfPresent()
        submitButton.tap()
        return self
    }

    /// Perform full login flow
    @discardableResult
    func login(email: String, password: String) -> GuildListPage {
        switchToLogin()
        enterEmail(email)
        enterPassword(password)
        tapSubmit()

        // Wait for transition to guild list and data to load
        let guildListPage = GuildListPage(app: app)
        _ = guildListPage.waitForGuildsToLoad()
        return guildListPage
    }

    /// Perform login with test user
    @discardableResult
    func loginAs(_ user: TestUsers.User) -> GuildListPage {
        login(email: user.email, password: user.password)
    }

    /// Perform full registration flow
    @discardableResult
    func register(
        email: String,
        password: String,
        firstname: String? = nil,
        lastname: String? = nil
    ) -> GuildListPage {
        switchToRegister()

        if let fn = firstname {
            enterFirstname(fn)
        }
        if let ln = lastname {
            enterLastname(ln)
        }

        enterEmail(email)
        enterPassword(password)
        tapSubmit()

        return GuildListPage(app: app)
    }

    /// Tap Sign in with Google
    @discardableResult
    func tapGoogleSignIn() -> AuthPage {
        googleButton.tap()
        return self
    }

    /// Tap Sign in with Apple
    @discardableResult
    func tapAppleSignIn() -> AuthPage {
        appleButton.tap()
        return self
    }

    /// Tap Sign in with Passkey
    @discardableResult
    func tapPasskeySignIn() -> AuthPage {
        passkeyButton.tap()
        return self
    }
}
