import XCTest

/// Base class for all page objects
@MainActor
class BasePage {
    let app: XCUIApplication

    init(app: XCUIApplication) {
        self.app = app
    }

    // MARK: - Wait Helpers

    /// Wait for an element to exist
    @discardableResult
    func waitForElement(_ element: XCUIElement, timeout: TimeInterval = TestConfig.defaultTimeout) -> Bool {
        element.waitForExistence(timeout: timeout)
    }

    /// Wait for element to disappear
    @discardableResult
    func waitForElementToDisappear(_ element: XCUIElement, timeout: TimeInterval = TestConfig.defaultTimeout) -> Bool {
        element.waitForDisappear(timeout: timeout)
    }

    // MARK: - Assertions

    /// Assert element exists
    func assertExists(_ element: XCUIElement, message: String = "") {
        XCTAssertTrue(element.exists, message.isEmpty ? "Element should exist" : message)
    }

    /// Assert element does not exist
    func assertNotExists(_ element: XCUIElement, message: String = "") {
        XCTAssertFalse(element.exists, message.isEmpty ? "Element should not exist" : message)
    }

    /// Assert element is enabled
    func assertEnabled(_ element: XCUIElement, message: String = "") {
        XCTAssertTrue(element.isEnabled, message.isEmpty ? "Element should be enabled" : message)
    }

    /// Assert element is disabled
    func assertDisabled(_ element: XCUIElement, message: String = "") {
        XCTAssertFalse(element.isEnabled, message.isEmpty ? "Element should be disabled" : message)
    }

    // MARK: - Scroll Helpers

    /// Scroll to find an element in a container
    func scrollToElement(_ element: XCUIElement, inContainer container: XCUIElement, maxAttempts: Int = 10) {
        var attempts = 0
        while !element.isHittable && attempts < maxAttempts {
            container.swipeUp()
            attempts += 1
        }
    }

    /// Scroll down in a container
    func scrollDown(in container: XCUIElement) {
        container.swipeUp()
    }

    /// Scroll up in a container
    func scrollUp(in container: XCUIElement) {
        container.swipeDown()
    }

    // MARK: - Navigation

    /// Go back using navigation back button
    func goBack() {
        let backButton = app.navigationBars.buttons.element(boundBy: 0)
        if backButton.exists {
            backButton.tap()
        }
    }

    /// Dismiss any presented sheet or modal
    func dismissSheet() {
        // Try tapping outside the sheet
        app.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: 0.1)).tap()
    }

    // MARK: - Common Elements

    /// Loading indicator
    var loadingIndicator: XCUIElement {
        app.activityIndicators[AccessibilityID.Common.loadingIndicator]
    }

    /// Error view
    var errorView: XCUIElement {
        app.otherElements[AccessibilityID.Common.errorView]
    }

    /// Retry button
    var retryButton: XCUIElement {
        app.buttons[AccessibilityID.Common.retryButton]
    }

    /// Wait for loading to complete
    func waitForLoadingToComplete(timeout: TimeInterval = TestConfig.longTimeout) {
        if loadingIndicator.exists {
            _ = loadingIndicator.waitForDisappear(timeout: timeout)
        }
    }
}
