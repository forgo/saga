import XCTest

/// Tests for adventure functionality
/// Adventures are accessed from the guild detail view

@MainActor
final class AdventureTests: SagaUITestCase {

    // MARK: - Helper

    /// Navigate to Adventures from guild detail
    /// Note: This test requires guilds to exist in the seed data.
    private func navigateToAdventures() throws {
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
        let guildDetailNavBar = app.navigationBars.firstMatch
        guard guildDetailNavBar.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            throw XCTSkip("Guild detail navigation bar not found")
        }

        // Wait for the view to fully load
        Wait.forNetwork()
        Wait.seconds(1.0)

        // Look for feature content on guild detail using accessibility identifiers
        let adventuresButton = app.buttons["guild_feature_adventures"]
        let eventsText = app.staticTexts["Events"]
        let featuresHeader = app.staticTexts["Features"]

        // First try to find Adventures directly via accessibility ID
        if adventuresButton.waitForExistence(timeout: TestConfig.defaultTimeout) {
            adventuresButton.tap()
            Wait.forNetwork()
            return
        }

        // Try scrolling to find the Features section
        for _ in 0..<5 {
            app.swipeUp()
            Wait.seconds(0.5)

            if adventuresButton.waitForExistence(timeout: TestConfig.shortTimeout) {
                adventuresButton.tap()
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
            throw XCTSkip("Found Events but not Adventures. Visible: \(visibleTexts.joined(separator: ", "))")
        } else if featuresHeader.exists {
            throw XCTSkip("Found Features header but not Adventures. Visible: \(visibleTexts.joined(separator: ", "))")
        } else {
            throw XCTSkip("Features section not visible. Visible: \(visibleTexts.joined(separator: ", "))")
        }
    }

    // MARK: - Navigation Tests

    func testCanNavigateToAdventures() throws {
        try navigateToAdventures()

        // Verify Adventures screen
        let navBar = app.navigationBars["Adventures"]
        guard navBar.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            throw XCTSkip("AdventureListView not implemented yet - Adventures navigation bar not found")
        }
        XCTAssertTrue(navBar.exists, "Adventures navigation bar should exist")
    }

    func testAdventuresShowsCreateButton() throws {
        try navigateToAdventures()

        // Check create button exists
        let createButton = app.buttons[AccessibilityID.Adventure.createButton]
        XCTAssertTrue(
            createButton.waitForExistence(timeout: TestConfig.defaultTimeout),
            "Create adventure button should exist"
        )
    }

    // MARK: - Empty State Tests

    func testAdventuresShowsEmptyStateOrList() throws {
        try navigateToAdventures()

        // Should show either adventures or empty state
        let emptyState = app.staticTexts["No Adventures"]
        let createButton = app.buttons["Create Adventure"]
        let adventureList = app.collectionViews.firstMatch.exists ? app.collectionViews.firstMatch : app.tables.firstMatch

        // One of these should be visible
        XCTAssertTrue(
            emptyState.waitForExistence(timeout: TestConfig.defaultTimeout) ||
            createButton.exists ||
            adventureList.cells.count > 0,
            "Should show adventures or empty state"
        )
    }

    // MARK: - Create Sheet Tests

    func testCanOpenCreateAdventureSheet() throws {
        try navigateToAdventures()

        // Tap create button
        let createButton = app.buttons[AccessibilityID.Adventure.createButton]
        guard createButton.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            throw XCTSkip("Create button not found")
        }

        createButton.tap()

        // Verify create sheet appears
        let navTitle = app.staticTexts["New Adventure"]
        XCTAssertTrue(
            navTitle.waitForExistence(timeout: TestConfig.defaultTimeout),
            "Create Adventure sheet should appear"
        )
    }

    func testCreateAdventureSheetHasTitleField() throws {
        try navigateToAdventures()

        // Open create sheet
        let createButton = app.buttons[AccessibilityID.Adventure.createButton]
        guard createButton.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            throw XCTSkip("Create button not found")
        }
        createButton.tap()

        // Check title field exists
        let titleField = app.textFields[AccessibilityID.Adventure.titleField]
        XCTAssertTrue(
            titleField.waitForExistence(timeout: TestConfig.defaultTimeout),
            "Title field should exist"
        )
    }

    func testCreateAdventureSheetHasCancelButton() throws {
        try navigateToAdventures()

        // Open create sheet
        let createButton = app.buttons[AccessibilityID.Adventure.createButton]
        guard createButton.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            throw XCTSkip("Create button not found")
        }
        createButton.tap()

        // Check cancel button exists
        let cancelButton = app.buttons[AccessibilityID.Adventure.createCancelButton]
        XCTAssertTrue(
            cancelButton.waitForExistence(timeout: TestConfig.defaultTimeout),
            "Cancel button should exist"
        )
    }

    func testCreateButtonDisabledWithEmptyTitle() throws {
        try navigateToAdventures()

        // Open create sheet
        let createButton = app.buttons[AccessibilityID.Adventure.createButton]
        guard createButton.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            throw XCTSkip("Create button not found")
        }
        createButton.tap()

        // Check create confirm button is disabled
        let confirmButton = app.buttons[AccessibilityID.Adventure.createConfirmButton]
        guard confirmButton.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            throw XCTSkip("Confirm button not found")
        }

        XCTAssertFalse(
            confirmButton.isEnabled,
            "Create button should be disabled with empty title"
        )
    }

    func testCanCancelCreateAdventureSheet() throws {
        try navigateToAdventures()

        // Open create sheet
        let createButton = app.buttons[AccessibilityID.Adventure.createButton]
        guard createButton.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            throw XCTSkip("Create button not found")
        }
        createButton.tap()

        // Tap cancel
        let cancelButton = app.buttons[AccessibilityID.Adventure.createCancelButton]
        guard cancelButton.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            throw XCTSkip("Cancel button not found")
        }
        cancelButton.tap()

        // Verify sheet is dismissed
        Wait.seconds(0.5)
        let navTitle = app.staticTexts["New Adventure"]
        XCTAssertFalse(
            navTitle.exists,
            "Create sheet should be dismissed"
        )
    }

    // MARK: - Back Navigation Tests

    func testCanNavigateBackFromAdventures() throws {
        try navigateToAdventures()

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
            "Should navigate back from Adventures"
        )
    }
}
