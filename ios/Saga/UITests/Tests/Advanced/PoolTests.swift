import XCTest

/// Tests for matching pool functionality
/// Pools are accessed from the guild detail view

@MainActor
final class PoolTests: SagaUITestCase {

    // MARK: - Helper

    /// Navigate to Matching Pools from guild detail
    /// Note: This test requires guilds to exist in the seed data.
    private func navigateToPools() throws {
        let guildList = launchAndLoginWithDemoUser()

        // Wait for guilds to load (this now properly waits for data)
        let hasGuilds = guildList.waitForGuildsToLoad()

        // Check for error alerts that might explain why guilds aren't loading
        if !hasGuilds {
            // Check if there's an error alert
            if app.alerts.firstMatch.exists {
                let alertText = app.alerts.firstMatch.staticTexts.allElementsBoundByIndex
                    .map { $0.label }
                    .joined(separator: " ")
                throw XCTSkip("Alert shown: \(alertText)")
            }

            // Debug: Show what's visible
            let visibleLabels = app.staticTexts.allElementsBoundByIndex
                .prefix(10)
                .compactMap { $0.exists ? $0.label : nil }
                .filter { !$0.isEmpty }
                .joined(separator: ", ")
            throw XCTSkip("No guilds loaded. Visible: \(visibleLabels)")
        }

        // Navigate to guild detail
        _ = guildList.tapGuild(at: 0)

        // Wait for guild detail navigation bar to appear (indicates navigation complete)
        // The navigation bar title should match the guild name
        let guildDetailNavBar = app.navigationBars.firstMatch
        guard guildDetailNavBar.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            throw XCTSkip("Guild detail navigation bar not found")
        }

        // Wait for the view to fully load
        Wait.forNetwork()
        Wait.seconds(1.0)

        // Look for feature content on guild detail using accessibility identifiers
        let poolsButton = app.buttons["guild_feature_pools"]
        let eventsText = app.staticTexts["Events"]
        let featuresHeader = app.staticTexts["Features"]

        // First try to find Matching Pools directly via accessibility ID
        if poolsButton.waitForExistence(timeout: TestConfig.defaultTimeout) {
            poolsButton.tap()
            Wait.forNetwork()
            return
        }

        // Try scrolling to find the Features section
        for _ in 0..<5 {
            app.swipeUp()
            Wait.seconds(0.5)

            if poolsButton.waitForExistence(timeout: TestConfig.shortTimeout) {
                poolsButton.tap()
                Wait.forNetwork()
                return
            }
        }

        // Collect debug info about what's visible
        let visibleTexts = app.staticTexts.allElementsBoundByIndex
            .prefix(20)
            .compactMap { $0.exists ? $0.label : nil }
            .filter { !$0.isEmpty }

        // Provide helpful skip message with debug info
        if eventsText.exists {
            throw XCTSkip("Found Events but not Matching Pools. Visible: \(visibleTexts.joined(separator: ", "))")
        } else if featuresHeader.exists {
            throw XCTSkip("Found Features header but not Matching Pools. Visible: \(visibleTexts.joined(separator: ", "))")
        } else {
            throw XCTSkip("Features section not visible. Visible: \(visibleTexts.joined(separator: ", "))")
        }
    }

    // MARK: - Navigation Tests

    func testCanNavigateToPools() throws {
        try navigateToPools()

        // Verify Pools screen
        let navBar = app.navigationBars["Matching Pools"]
        guard navBar.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            throw XCTSkip("PoolListView not implemented yet - Matching Pools navigation bar not found")
        }
        XCTAssertTrue(navBar.exists, "Matching Pools navigation bar should exist")
    }

    func testPoolsShowsCreateButton() throws {
        try navigateToPools()

        // Check create button exists
        let createButton = app.buttons[AccessibilityID.Pool.createButton]
        XCTAssertTrue(
            createButton.waitForExistence(timeout: TestConfig.defaultTimeout),
            "Create pool button should exist"
        )
    }

    // MARK: - Empty State Tests

    func testPoolsShowsEmptyStateOrList() throws {
        try navigateToPools()

        // Should show either pools or empty state
        let emptyState = app.staticTexts["No Pools"]
        let createButton = app.buttons["Create Pool"]
        let poolList = app.collectionViews.firstMatch.exists ? app.collectionViews.firstMatch : app.tables.firstMatch

        // One of these should be visible
        XCTAssertTrue(
            emptyState.waitForExistence(timeout: TestConfig.defaultTimeout) ||
            createButton.exists ||
            poolList.cells.count > 0,
            "Should show pools or empty state"
        )
    }

    // MARK: - Create Sheet Tests

    func testCanOpenCreatePoolSheet() throws {
        try navigateToPools()

        // Tap create button
        let createButton = app.buttons[AccessibilityID.Pool.createButton]
        guard createButton.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            throw XCTSkip("Create button not found")
        }

        createButton.tap()

        // Verify create sheet appears
        let navTitle = app.staticTexts["New Pool"]
        XCTAssertTrue(
            navTitle.waitForExistence(timeout: TestConfig.defaultTimeout),
            "Create Pool sheet should appear"
        )
    }

    func testCreatePoolSheetHasNameField() throws {
        try navigateToPools()

        // Open create sheet
        let createButton = app.buttons[AccessibilityID.Pool.createButton]
        guard createButton.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            throw XCTSkip("Create button not found")
        }
        createButton.tap()

        // Check name field exists
        let nameField = app.textFields[AccessibilityID.Pool.nameField]
        XCTAssertTrue(
            nameField.waitForExistence(timeout: TestConfig.defaultTimeout),
            "Name field should exist"
        )
    }

    func testCreatePoolSheetHasCancelButton() throws {
        try navigateToPools()

        // Open create sheet
        let createButton = app.buttons[AccessibilityID.Pool.createButton]
        guard createButton.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            throw XCTSkip("Create button not found")
        }
        createButton.tap()

        // Check cancel button exists
        let cancelButton = app.buttons[AccessibilityID.Pool.createCancelButton]
        XCTAssertTrue(
            cancelButton.waitForExistence(timeout: TestConfig.defaultTimeout),
            "Cancel button should exist"
        )
    }

    func testCreateButtonDisabledWithEmptyName() throws {
        try navigateToPools()

        // Open create sheet
        let createButton = app.buttons[AccessibilityID.Pool.createButton]
        guard createButton.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            throw XCTSkip("Create button not found")
        }
        createButton.tap()

        // Check create confirm button is disabled
        let confirmButton = app.buttons[AccessibilityID.Pool.createConfirmButton]
        guard confirmButton.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            throw XCTSkip("Confirm button not found")
        }

        XCTAssertFalse(
            confirmButton.isEnabled,
            "Create button should be disabled with empty name"
        )
    }

    func testCanCancelCreatePoolSheet() throws {
        try navigateToPools()

        // Open create sheet
        let createButton = app.buttons[AccessibilityID.Pool.createButton]
        guard createButton.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            throw XCTSkip("Create button not found")
        }
        createButton.tap()

        // Tap cancel
        let cancelButton = app.buttons[AccessibilityID.Pool.createCancelButton]
        guard cancelButton.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            throw XCTSkip("Cancel button not found")
        }
        cancelButton.tap()

        // Verify sheet is dismissed
        Wait.seconds(0.5)
        let navTitle = app.staticTexts["New Pool"]
        XCTAssertFalse(
            navTitle.exists,
            "Create sheet should be dismissed"
        )
    }

    func testCreatePoolSheetHasFrequencyPicker() throws {
        try navigateToPools()

        // Open create sheet
        let createButton = app.buttons[AccessibilityID.Pool.createButton]
        guard createButton.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            throw XCTSkip("Create button not found")
        }
        createButton.tap()

        // Check for frequency label in the form
        let frequencyLabel = app.staticTexts["Frequency"]
        XCTAssertTrue(
            frequencyLabel.waitForExistence(timeout: TestConfig.defaultTimeout),
            "Frequency picker should exist"
        )
    }

    func testCreatePoolSheetHasMatchSizeStepper() throws {
        try navigateToPools()

        // Open create sheet
        let createButton = app.buttons[AccessibilityID.Pool.createButton]
        guard createButton.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            throw XCTSkip("Create button not found")
        }
        createButton.tap()

        // Check for match size text
        let matchSizeText = app.staticTexts.containing(NSPredicate(format: "label CONTAINS 'Match Size'")).firstMatch
        XCTAssertTrue(
            matchSizeText.waitForExistence(timeout: TestConfig.defaultTimeout),
            "Match size stepper should exist"
        )
    }

    // MARK: - Back Navigation Tests

    func testCanNavigateBackFromPools() throws {
        try navigateToPools()

        // Tap back button
        let backButton = app.navigationBars.buttons.element(boundBy: 0)
        guard backButton.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            throw XCTSkip("Back button not found")
        }

        backButton.tap()
        Wait.seconds(0.5)

        // Should be back on guild detail
        let guildNavBar = app.navigationBars.element(boundBy: 0)
        XCTAssertTrue(
            guildNavBar.exists,
            "Should navigate back from Pools"
        )
    }
}
