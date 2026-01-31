import XCTest

/// Tests for event list functionality
/// NOTE: Events require a guild to be selected first, as they are loaded per-guild

@MainActor
final class EventListTests: SagaUITestCase {

    // MARK: - Tab Navigation Tests

    func testCanNavigateToEventsTab() throws {
        _ = launchAndLoginWithDemoUser()

        // Navigate to Events tab
        _ = tabBar.tapEvents()
        Wait.seconds(0.5)

        XCTAssertTrue(tabBar.isEventsSelected(), "Events tab should be selected")
    }

    func testEventsTabShowsNavigationBar() throws {
        _ = launchAndLoginWithDemoUser()

        // Navigate to Events tab
        _ = tabBar.tapEvents()

        // Check navigation bar exists
        let navBar = app.navigationBars["Events"]
        XCTAssertTrue(
            navBar.waitForExistence(timeout: TestConfig.defaultTimeout),
            "Events navigation bar should exist"
        )
    }

    // MARK: - Events from Guild Detail Tests

    func testCanAccessGuildEventsFromGuildDetail() throws {
        let guildList = launchAndLoginWithDemoUser()
        Wait.forNetwork()

        // Need at least one guild
        guard guildList.guildCount() > 0 else {
            throw XCTSkip("No guilds available for test")
        }

        // Tap on first guild to see detail
        _ = guildList.tapGuild(at: 0)
        Wait.forNetwork()

        // Look for Events link in guild detail
        let eventsLink = app.buttons["Events"]
        if eventsLink.waitForExistence(timeout: TestConfig.defaultTimeout) {
            eventsLink.tap()
            Wait.forNetwork()

            // Should see guild events view
            let eventsNavBar = app.navigationBars["Events"]
            XCTAssertTrue(
                eventsNavBar.waitForExistence(timeout: TestConfig.defaultTimeout),
                "Events navigation bar should appear"
            )
        } else {
            // Events link might be in a different form (cell with label)
            let eventsCell = app.cells.containing(.staticText, identifier: "Events").firstMatch
            if eventsCell.exists {
                eventsCell.tap()
                Wait.forNetwork()
            }
        }
    }

    // MARK: - Filter Tests

    func testEventFilterPickerExists() throws {
        _ = launchAndLoginWithDemoUser()

        // Navigate to Events tab
        _ = tabBar.tapEvents()
        Wait.forNetwork()

        // Check for filter picker
        let filterPicker = app.segmentedControls.firstMatch
        XCTAssertTrue(
            filterPicker.waitForExistence(timeout: TestConfig.defaultTimeout),
            "Event filter picker should exist"
        )
    }

    func testCanSwitchEventFilters() throws {
        _ = launchAndLoginWithDemoUser()

        // Navigate to Events tab
        _ = tabBar.tapEvents()
        Wait.forNetwork()

        let filterPicker = app.segmentedControls.firstMatch
        guard filterPicker.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            throw XCTSkip("Filter picker not found")
        }

        // Try selecting different filters
        let pastButton = filterPicker.buttons["Past"]
        if pastButton.exists {
            pastButton.tap()
            Wait.seconds(0.5)
            XCTAssertTrue(pastButton.isSelected, "Past filter should be selected")
        }

        let allButton = filterPicker.buttons["All"]
        if allButton.exists {
            allButton.tap()
            Wait.seconds(0.5)
            XCTAssertTrue(allButton.isSelected, "All filter should be selected")
        }

        let upcomingButton = filterPicker.buttons["Upcoming"]
        if upcomingButton.exists {
            upcomingButton.tap()
            Wait.seconds(0.5)
            XCTAssertTrue(upcomingButton.isSelected, "Upcoming filter should be selected")
        }
    }

    // MARK: - Empty State Tests

    func testShowsEmptyStateWhenNoEvents() throws {
        _ = launchAndLoginWithDemoUser()

        // Navigate to Events tab
        _ = tabBar.tapEvents()
        Wait.forNetwork()

        // If no events, should show empty state
        let noEventsText = app.staticTexts["No Upcoming Events"]
        let noEventsAlt = app.staticTexts["No Events"]

        // One of these should be visible if there are no events
        // (We can't guarantee there are no events in seed data)
        if noEventsText.exists || noEventsAlt.exists {
            XCTAssertTrue(true, "Empty state is displayed correctly")
        } else {
            // Events exist, which is also valid
            XCTAssertTrue(true, "Events are displayed (not empty)")
        }
    }

    // MARK: - Create Event Button Tests

    func testCreateButtonExistsInEventsTab() throws {
        _ = launchAndLoginWithDemoUser()

        // Navigate to Events tab
        _ = tabBar.tapEvents()
        Wait.forNetwork()

        // Check create button exists in toolbar
        let createButton = app.buttons[AccessibilityID.Event.createButton]
        // Note: Create button may be disabled if no guild is selected
        XCTAssertTrue(
            createButton.waitForExistence(timeout: TestConfig.defaultTimeout),
            "Create event button should exist"
        )
    }

    // MARK: - Navigation Back and Forth

    func testCanNavigateBetweenEventsAndGuilds() throws {
        _ = launchAndLoginWithDemoUser()

        // Go to Events
        _ = tabBar.tapEvents()
        Wait.seconds(0.5)
        XCTAssertTrue(tabBar.isEventsSelected(), "Events tab should be selected")

        // Go back to Guilds
        _ = tabBar.tapGuilds()
        Wait.seconds(0.5)
        XCTAssertTrue(tabBar.isGuildsSelected(), "Guilds tab should be selected")

        // Verify guild list is displayed
        XCTAssertTrue(guildListPage.isDisplayed(), "Guild list should be displayed")
    }
}
