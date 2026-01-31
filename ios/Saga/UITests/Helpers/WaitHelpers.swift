import XCTest

// MARK: - XCUIElement Wait Extensions

@MainActor
extension XCUIElement {

    /// Wait for element to become hittable (visible and interactable)
    @discardableResult
    func waitForHittable(timeout: TimeInterval = TestConfig.defaultTimeout) -> Bool {
        let predicate = NSPredicate(format: "isHittable == true")
        let expectation = XCTNSPredicateExpectation(predicate: predicate, object: self)
        return XCTWaiter.wait(for: [expectation], timeout: timeout) == .completed
    }

    /// Wait for element to be enabled
    @discardableResult
    func waitForEnabled(timeout: TimeInterval = TestConfig.defaultTimeout) -> Bool {
        let predicate = NSPredicate(format: "isEnabled == true")
        let expectation = XCTNSPredicateExpectation(predicate: predicate, object: self)
        return XCTWaiter.wait(for: [expectation], timeout: timeout) == .completed
    }

    /// Wait for element to disappear
    @discardableResult
    func waitForDisappear(timeout: TimeInterval = TestConfig.defaultTimeout) -> Bool {
        let predicate = NSPredicate(format: "exists == false")
        let expectation = XCTNSPredicateExpectation(predicate: predicate, object: self)
        return XCTWaiter.wait(for: [expectation], timeout: timeout) == .completed
    }

    /// Wait for element label to match expected value
    @discardableResult
    func waitForLabel(_ expectedLabel: String, timeout: TimeInterval = TestConfig.defaultTimeout) -> Bool {
        let predicate = NSPredicate(format: "label == %@", expectedLabel)
        let expectation = XCTNSPredicateExpectation(predicate: predicate, object: self)
        return XCTWaiter.wait(for: [expectation], timeout: timeout) == .completed
    }

    /// Wait for element label to contain expected text
    @discardableResult
    func waitForLabelContaining(_ text: String, timeout: TimeInterval = TestConfig.defaultTimeout) -> Bool {
        let predicate = NSPredicate(format: "label CONTAINS %@", text)
        let expectation = XCTNSPredicateExpectation(predicate: predicate, object: self)
        return XCTWaiter.wait(for: [expectation], timeout: timeout) == .completed
    }

    /// Wait for element value to match
    @discardableResult
    func waitForValue(_ expectedValue: String, timeout: TimeInterval = TestConfig.defaultTimeout) -> Bool {
        let predicate = NSPredicate(format: "value == %@", expectedValue)
        let expectation = XCTNSPredicateExpectation(predicate: predicate, object: self)
        return XCTWaiter.wait(for: [expectation], timeout: timeout) == .completed
    }

    /// Tap element after waiting for it to be hittable
    func tapWhenReady(timeout: TimeInterval = TestConfig.defaultTimeout) {
        _ = waitForHittable(timeout: timeout)
        tap()
    }

    /// Clear text and type new text
    func clearAndType(_ text: String) {
        guard let currentValue = value as? String, !currentValue.isEmpty else {
            tap()
            typeText(text)
            return
        }

        tap()
        // Select all and delete
        let deleteString = String(repeating: XCUIKeyboardKey.delete.rawValue, count: currentValue.count)
        typeText(deleteString)
        typeText(text)
    }
}

// MARK: - XCUIApplication Wait Extensions

@MainActor
extension XCUIApplication {

    /// Wait for any alert to appear
    @discardableResult
    func waitForAlert(timeout: TimeInterval = TestConfig.defaultTimeout) -> Bool {
        alerts.firstMatch.waitForExistence(timeout: timeout)
    }

    /// Wait for any sheet to appear
    @discardableResult
    func waitForSheet(timeout: TimeInterval = TestConfig.defaultTimeout) -> Bool {
        sheets.firstMatch.waitForExistence(timeout: timeout)
    }

    /// Dismiss keyboard if visible
    func dismissKeyboardIfPresent() {
        if keyboards.count > 0 {
            // Tap outside any text field to dismiss
            coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: 0.1)).tap()
        }
    }
}

// MARK: - Wait Utilities

enum Wait {

    /// Wait for a condition to be true
    static func until(
        _ condition: @escaping () -> Bool,
        timeout: TimeInterval = TestConfig.defaultTimeout,
        pollInterval: TimeInterval = 0.1
    ) -> Bool {
        let deadline = Date().addingTimeInterval(timeout)

        while Date() < deadline {
            if condition() {
                return true
            }
            Thread.sleep(forTimeInterval: pollInterval)
        }

        return condition()
    }

    /// Wait for a fixed duration
    static func seconds(_ seconds: TimeInterval) {
        Thread.sleep(forTimeInterval: seconds)
    }

    /// Wait for network operation
    static func forNetwork() {
        seconds(TestConfig.networkDelay)
    }

    /// Wait for animation
    static func forAnimation() {
        seconds(TestConfig.animationDelay)
    }

    /// Wait for SSE sync
    static func forSSESync() {
        seconds(TestConfig.sseDelay)
    }
}
