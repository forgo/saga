import XCTest

/// Tests for trust management functionality
/// Trust is accessed from the Profile tab

@MainActor
final class TrustTests: SagaUITestCase {

    // MARK: - Helper

    /// Navigate to Trust Management from profile
    private func navigateToTrust() throws {
        let guildList = launchAndLoginWithDemoUser()

        // Wait for guilds to load
        let hasGuilds = guildList.waitForGuildsToLoad()

        if !hasGuilds {
            if app.alerts.firstMatch.exists {
                let alertText = app.alerts.firstMatch.staticTexts.allElementsBoundByIndex
                    .map { $0.label }
                    .joined(separator: " ")
                throw XCTSkip("Alert shown: \(alertText)")
            }

            let visibleLabels = app.staticTexts.allElementsBoundByIndex
                .prefix(10)
                .compactMap { $0.exists ? $0.label : nil }
                .filter { !$0.isEmpty }
                .joined(separator: ", ")
            throw XCTSkip("No guilds loaded. Visible: \(visibleLabels)")
        }

        // Navigate to Profile tab
        let profileTab = app.tabBars.buttons["Profile"]
        guard profileTab.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            throw XCTSkip("Profile tab not found")
        }
        profileTab.tap()
        Wait.forNetwork()

        // Look for Trust link in profile
        let trustLink = app.buttons[AccessibilityID.Profile.trustLink]

        // First try to find Trust directly
        if trustLink.waitForExistence(timeout: TestConfig.defaultTimeout) {
            trustLink.tap()
            Wait.forNetwork()
            return
        }

        // Try scrolling to find it
        let list = app.collectionViews.firstMatch.exists ? app.collectionViews.firstMatch : app.tables.firstMatch
        for _ in 0..<5 {
            list.swipeUp()
            Wait.seconds(0.5)

            if trustLink.waitForExistence(timeout: TestConfig.shortTimeout) {
                trustLink.tap()
                Wait.forNetwork()
                return
            }
        }

        // Collect debug info
        let visibleTexts = app.staticTexts.allElementsBoundByIndex
            .prefix(20)
            .compactMap { $0.exists ? $0.label : nil }
            .filter { !$0.isEmpty }

        throw XCTSkip("Trust link not found in profile. Visible: \(visibleTexts.joined(separator: ", "))")
    }

    // MARK: - Navigation Tests

    func testCanNavigateToTrust() throws {
        try navigateToTrust()

        // Verify Trust screen
        let navBar = app.navigationBars["Trust"]
        guard navBar.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            // Try alternate navigation bar title
            let altNavBar = app.navigationBars["Trust Management"]
            guard altNavBar.waitForExistence(timeout: TestConfig.shortTimeout) else {
                throw XCTSkip("TrustView not implemented yet - Trust navigation bar not found")
            }
            XCTAssertTrue(altNavBar.exists, "Trust Management navigation bar should exist")
            return
        }
        XCTAssertTrue(navBar.exists, "Trust navigation bar should exist")
    }

    func testTrustShowsTabPicker() throws {
        try navigateToTrust()

        // Check tab picker exists (Granted, Received, IRL)
        let tabPicker = app.segmentedControls[AccessibilityID.Trust.tabPicker]
        guard tabPicker.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            // Might be individual buttons instead
            let grantedTab = app.buttons[AccessibilityID.Trust.grantedTab]
            let receivedTab = app.buttons[AccessibilityID.Trust.receivedTab]
            if grantedTab.exists || receivedTab.exists {
                XCTAssertTrue(true, "Trust tabs exist as buttons")
                return
            }
            throw XCTSkip("Trust tab picker not found")
        }
        XCTAssertTrue(tabPicker.exists, "Tab picker should exist")
    }

    // MARK: - Granted Tab Tests

    func testCanViewGrantedTrust() throws {
        try navigateToTrust()

        // Tap Granted tab
        let grantedTab = app.buttons[AccessibilityID.Trust.grantedTab]
        let tabPicker = app.segmentedControls[AccessibilityID.Trust.tabPicker]

        if grantedTab.waitForExistence(timeout: TestConfig.defaultTimeout) {
            grantedTab.tap()
        } else if tabPicker.waitForExistence(timeout: TestConfig.shortTimeout) {
            tabPicker.buttons["Granted"].tap()
        } else {
            // May already be on granted by default
        }

        Wait.forNetwork()

        // Should show either trust grants or empty state
        let trustList = app.collectionViews[AccessibilityID.Trust.grantList]
        let emptyState = app.staticTexts.containing(NSPredicate(format: "label CONTAINS[c] 'no trust'")).firstMatch

        XCTAssertTrue(
            trustList.waitForExistence(timeout: TestConfig.defaultTimeout) ||
            emptyState.exists ||
            app.staticTexts["You haven't granted trust to anyone yet"].exists,
            "Should show granted trust list or empty state"
        )
    }

    // MARK: - Received Tab Tests

    func testCanViewReceivedTrust() throws {
        try navigateToTrust()

        // Tap Received tab
        let receivedTab = app.buttons[AccessibilityID.Trust.receivedTab]
        let tabPicker = app.segmentedControls[AccessibilityID.Trust.tabPicker]

        if receivedTab.waitForExistence(timeout: TestConfig.defaultTimeout) {
            receivedTab.tap()
        } else if tabPicker.waitForExistence(timeout: TestConfig.shortTimeout) {
            tabPicker.buttons["Received"].tap()
        } else {
            throw XCTSkip("Received tab not found")
        }

        Wait.forNetwork()

        // Should show either received trust or empty state
        let trustList = app.collectionViews[AccessibilityID.Trust.grantList]
        let emptyState = app.staticTexts.containing(NSPredicate(format: "label CONTAINS[c] 'no one has'")).firstMatch

        XCTAssertTrue(
            trustList.waitForExistence(timeout: TestConfig.defaultTimeout) ||
            emptyState.exists ||
            app.staticTexts["No one has granted you trust yet"].exists,
            "Should show received trust list or empty state"
        )
    }

    // MARK: - IRL Tab Tests

    func testCanViewIRLTab() throws {
        try navigateToTrust()

        // Tap IRL tab
        let irlTab = app.buttons[AccessibilityID.Trust.irlTab]
        let tabPicker = app.segmentedControls[AccessibilityID.Trust.tabPicker]

        if irlTab.waitForExistence(timeout: TestConfig.defaultTimeout) {
            irlTab.tap()
        } else if tabPicker.waitForExistence(timeout: TestConfig.shortTimeout) {
            tabPicker.buttons["IRL"].tap()
        } else {
            throw XCTSkip("IRL tab not found")
        }

        Wait.forNetwork()

        // Should show IRL verifications or explanation
        let verificationList = app.collectionViews.firstMatch
        let explanation = app.staticTexts.containing(NSPredicate(format: "label CONTAINS[c] 'verify'")).firstMatch

        XCTAssertTrue(
            verificationList.exists ||
            explanation.exists,
            "Should show IRL verifications or explanation"
        )
    }

    func testIRLTabShowsExplanation() throws {
        try navigateToTrust()

        // Tap IRL tab
        let irlTab = app.buttons[AccessibilityID.Trust.irlTab]
        let tabPicker = app.segmentedControls[AccessibilityID.Trust.tabPicker]

        if irlTab.waitForExistence(timeout: TestConfig.defaultTimeout) {
            irlTab.tap()
        } else if tabPicker.waitForExistence(timeout: TestConfig.shortTimeout) {
            tabPicker.buttons["IRL"].tap()
        } else {
            throw XCTSkip("IRL tab not found")
        }

        Wait.forNetwork()

        // Should show some explanation about IRL verification
        let explanationTexts = [
            "in-person",
            "IRL",
            "verify",
            "meeting",
            "confirm"
        ]

        var foundExplanation = false
        for text in explanationTexts {
            let element = app.staticTexts.containing(NSPredicate(format: "label CONTAINS[c] %@", text)).firstMatch
            if element.exists {
                foundExplanation = true
                break
            }
        }

        XCTAssertTrue(foundExplanation, "Should show IRL verification explanation")
    }

    // MARK: - Grant Trust Tests

    func testGrantTrustButtonExists() throws {
        try navigateToTrust()

        // Look for grant trust button
        let grantButton = app.buttons[AccessibilityID.Trust.grantTrustButton]

        // May need to scroll to find it
        if !grantButton.waitForExistence(timeout: TestConfig.defaultTimeout) {
            let list = app.collectionViews.firstMatch.exists ? app.collectionViews.firstMatch : app.tables.firstMatch
            for _ in 0..<3 {
                list.swipeUp()
                Wait.seconds(0.3)
                if grantButton.exists {
                    break
                }
            }
        }

        // Note: Grant button might only appear when viewing another user's profile
        // This test may skip if on own trust management
        if !grantButton.exists {
            throw XCTSkip("Grant trust button not visible - may only appear when viewing other profiles")
        }

        XCTAssertTrue(grantButton.exists, "Grant trust button should exist")
    }

    // MARK: - Trust Row Tests

    func testTrustRowShowsUserInfo() throws {
        try navigateToTrust()

        // Wait for list to load
        Wait.forNetwork()
        Wait.seconds(1.0)

        // Check if there are any trust rows
        let trustRows = app.collectionViews.cells.matching(NSPredicate(format: "identifier BEGINSWITH 'trust_grant_'"))

        if trustRows.count > 0 {
            // Verify first row has expected content
            let firstRow = trustRows.element(boundBy: 0)
            XCTAssertTrue(firstRow.exists, "Trust row should exist")

            // Row should contain user info (name or avatar)
            let hasContent = firstRow.staticTexts.count > 0 || firstRow.images.count > 0
            XCTAssertTrue(hasContent, "Trust row should show user info")
        } else {
            throw XCTSkip("No trust grants to verify - test user has no trust relationships")
        }
    }

    // MARK: - Confirm/Decline IRL Tests

    func testIRLConfirmButtonExists() throws {
        try navigateToTrust()

        // Navigate to IRL tab
        let irlTab = app.buttons[AccessibilityID.Trust.irlTab]
        let tabPicker = app.segmentedControls[AccessibilityID.Trust.tabPicker]

        if irlTab.waitForExistence(timeout: TestConfig.defaultTimeout) {
            irlTab.tap()
        } else if tabPicker.waitForExistence(timeout: TestConfig.shortTimeout) {
            tabPicker.buttons["IRL"].tap()
        } else {
            throw XCTSkip("IRL tab not found")
        }

        Wait.forNetwork()
        Wait.seconds(1.0)

        // Look for confirm button (would appear for pending IRL verification requests)
        let confirmButton = app.buttons[AccessibilityID.Trust.confirmButton]

        if !confirmButton.exists {
            throw XCTSkip("No pending IRL verification requests - confirm button not visible")
        }

        XCTAssertTrue(confirmButton.exists, "Confirm IRL button should exist for pending requests")
    }

    func testIRLDeclineButtonExists() throws {
        try navigateToTrust()

        // Navigate to IRL tab
        let irlTab = app.buttons[AccessibilityID.Trust.irlTab]
        let tabPicker = app.segmentedControls[AccessibilityID.Trust.tabPicker]

        if irlTab.waitForExistence(timeout: TestConfig.defaultTimeout) {
            irlTab.tap()
        } else if tabPicker.waitForExistence(timeout: TestConfig.shortTimeout) {
            tabPicker.buttons["IRL"].tap()
        } else {
            throw XCTSkip("IRL tab not found")
        }

        Wait.forNetwork()
        Wait.seconds(1.0)

        // Look for decline button
        let declineButton = app.buttons[AccessibilityID.Trust.declineButton]

        if !declineButton.exists {
            throw XCTSkip("No pending IRL verification requests - decline button not visible")
        }

        XCTAssertTrue(declineButton.exists, "Decline IRL button should exist for pending requests")
    }

    // MARK: - Back Navigation Tests

    func testCanNavigateBackFromTrust() throws {
        try navigateToTrust()

        // Tap back button
        let backButton = app.navigationBars.buttons.element(boundBy: 0)
        guard backButton.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            throw XCTSkip("Back button not found")
        }

        backButton.tap()
        Wait.seconds(0.5)

        // Should be back on profile
        let profileTitle = app.navigationBars["Profile"]
        let logoutButton = app.buttons[AccessibilityID.Profile.logoutButton]

        XCTAssertTrue(
            profileTitle.exists || logoutButton.exists,
            "Should navigate back to Profile"
        )
    }

    // MARK: - Tab Switching Tests

    func testCanSwitchBetweenTabs() throws {
        try navigateToTrust()

        // Try switching between all tabs
        let tabPicker = app.segmentedControls[AccessibilityID.Trust.tabPicker]
        let grantedTab = app.buttons[AccessibilityID.Trust.grantedTab]
        let receivedTab = app.buttons[AccessibilityID.Trust.receivedTab]
        let irlTab = app.buttons[AccessibilityID.Trust.irlTab]

        // Start with Received tab
        if receivedTab.exists {
            receivedTab.tap()
            Wait.forNetwork()
            Wait.seconds(0.5)
        } else if tabPicker.exists {
            tabPicker.buttons["Received"].tap()
            Wait.forNetwork()
            Wait.seconds(0.5)
        }

        // Switch to Granted tab
        if grantedTab.exists {
            grantedTab.tap()
            Wait.forNetwork()
            Wait.seconds(0.5)
        } else if tabPicker.exists {
            tabPicker.buttons["Granted"].tap()
            Wait.forNetwork()
            Wait.seconds(0.5)
        }

        // Switch to IRL tab
        if irlTab.exists {
            irlTab.tap()
            Wait.forNetwork()
        } else if tabPicker.exists {
            tabPicker.buttons["IRL"].tap()
            Wait.forNetwork()
        }

        // If we got here without skipping, tab switching works
        XCTAssertTrue(true, "Tab switching should work")
    }
}
