import XCTest

/// Tests for discovery functionality

@MainActor
final class DiscoveryTests: SagaUITestCase {

    // MARK: - Tab Navigation Tests

    func testCanNavigateToDiscoverTab() throws {
        _ = launchAndLoginWithDemoUser()

        // Navigate to Discover tab
        _ = tabBar.tapDiscover()

        XCTAssertTrue(tabBar.isDiscoverSelected(), "Discover tab should be selected")
    }

    func testDiscoverTabShowsNavigationBar() throws {
        _ = launchAndLoginWithDemoUser()

        // Navigate to Discover tab
        _ = tabBar.tapDiscover()

        // Check navigation bar exists
        let navBar = app.navigationBars["Discover"]
        XCTAssertTrue(
            navBar.waitForExistence(timeout: TestConfig.defaultTimeout),
            "Discover navigation bar should exist"
        )
    }

    // MARK: - Tab Picker Tests

    func testDiscoverHasTabPicker() throws {
        _ = launchAndLoginWithDemoUser()

        // Navigate to Discover tab
        _ = tabBar.tapDiscover()
        Wait.forNetwork()

        // Check tab picker exists
        let tabPicker = app.segmentedControls.firstMatch
        XCTAssertTrue(
            tabPicker.waitForExistence(timeout: TestConfig.defaultTimeout),
            "Discover tab picker should exist"
        )
    }

    func testCanSwitchBetweenDiscoverTabs() throws {
        _ = launchAndLoginWithDemoUser()

        // Navigate to Discover tab
        _ = tabBar.tapDiscover()
        Wait.forNetwork()

        let tabPicker = app.segmentedControls.firstMatch
        guard tabPicker.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            throw XCTSkip("Tab picker not found")
        }

        // Default should be People
        let peopleButton = tabPicker.buttons["People"]
        let eventsButton = tabPicker.buttons["Events"]
        let interestsButton = tabPicker.buttons["Interests"]

        // Switch to Events
        if eventsButton.exists {
            eventsButton.tap()
            Wait.seconds(0.5)
            XCTAssertTrue(eventsButton.isSelected, "Events tab should be selected")
        }

        // Switch to Interests
        if interestsButton.exists {
            interestsButton.tap()
            Wait.seconds(0.5)
            XCTAssertTrue(interestsButton.isSelected, "Interests tab should be selected")
        }

        // Switch back to People
        if peopleButton.exists {
            peopleButton.tap()
            Wait.seconds(0.5)
            XCTAssertTrue(peopleButton.isSelected, "People tab should be selected")
        }
    }

    // MARK: - People Discovery Tests

    func testPeopleTabShowsSearchButton() throws {
        _ = launchAndLoginWithDemoUser()

        // Navigate to Discover tab
        _ = tabBar.tapDiscover()
        Wait.forNetwork()

        // Search button should exist in People tab
        let searchButton = app.buttons[AccessibilityID.Discovery.searchButton]
        XCTAssertTrue(
            searchButton.waitForExistence(timeout: TestConfig.defaultTimeout),
            "Search button should exist in People tab"
        )
    }

    func testPeopleTabShowsCompatibilitySlider() throws {
        _ = launchAndLoginWithDemoUser()

        // Navigate to Discover tab
        _ = tabBar.tapDiscover()
        Wait.forNetwork()

        // Should see compatibility text
        let compatibilityText = app.staticTexts.containing(NSPredicate(format: "label CONTAINS 'Compatibility'")).firstMatch
        XCTAssertTrue(
            compatibilityText.waitForExistence(timeout: TestConfig.defaultTimeout),
            "Compatibility slider label should exist"
        )
    }

    func testPeopleTabShowsRadiusSlider() throws {
        _ = launchAndLoginWithDemoUser()

        // Navigate to Discover tab
        _ = tabBar.tapDiscover()
        Wait.forNetwork()

        // Should see radius text
        let radiusText = app.staticTexts.containing(NSPredicate(format: "label CONTAINS 'Radius'")).firstMatch
        XCTAssertTrue(
            radiusText.waitForExistence(timeout: TestConfig.defaultTimeout),
            "Radius slider label should exist"
        )
    }

    // MARK: - Events Discovery Tests

    func testEventsTabShowsContentOrEmptyState() throws {
        _ = launchAndLoginWithDemoUser()

        // Navigate to Discover tab
        _ = tabBar.tapDiscover()
        Wait.forNetwork()

        // Switch to Events tab
        let tabPicker = app.segmentedControls.firstMatch
        guard tabPicker.waitForExistence(timeout: TestConfig.shortTimeout) else {
            throw XCTSkip("Tab picker not found")
        }

        let eventsButton = tabPicker.buttons["Events"]
        guard eventsButton.exists else {
            throw XCTSkip("Events button not found")
        }

        eventsButton.tap()
        Wait.forNetwork()

        // Should show either events or empty state
        let noEventsText = app.staticTexts["No Events Nearby"]
        let hasEvents = !app.cells.matching(NSPredicate(format: "identifier BEGINSWITH 'event_'")).firstMatch.exists

        // One of these conditions should be true
        XCTAssertTrue(
            noEventsText.exists || !hasEvents,
            "Should show events or empty state"
        )
    }

    // MARK: - Interests Discovery Tests

    func testInterestsTabShowsContentOrEmptyState() throws {
        _ = launchAndLoginWithDemoUser()

        // Navigate to Discover tab
        _ = tabBar.tapDiscover()
        Wait.forNetwork()

        // Switch to Interests tab
        let tabPicker = app.segmentedControls.firstMatch
        guard tabPicker.waitForExistence(timeout: TestConfig.shortTimeout) else {
            throw XCTSkip("Tab picker not found")
        }

        let interestsButton = tabPicker.buttons["Interests"]
        guard interestsButton.exists else {
            throw XCTSkip("Interests button not found")
        }

        interestsButton.tap()
        Wait.forNetwork()

        // Should show either matches or empty state with "Add Interests" button
        let noMatchesText = app.staticTexts["No Matches Yet"]
        let addInterestsLink = app.buttons["Add Interests"]

        // One of these conditions should be true
        XCTAssertTrue(
            noMatchesText.exists || addInterestsLink.exists || app.cells.count > 0,
            "Should show matches or empty state"
        )
    }

    // MARK: - Navigation Tests

    func testCanNavigateBackToGuildsFromDiscover() throws {
        _ = launchAndLoginWithDemoUser()

        // Navigate to Discover tab
        _ = tabBar.tapDiscover()
        Wait.seconds(0.5)
        XCTAssertTrue(tabBar.isDiscoverSelected(), "Discover tab should be selected")

        // Navigate back to Guilds
        _ = tabBar.tapGuilds()
        Wait.seconds(0.5)
        XCTAssertTrue(tabBar.isGuildsSelected(), "Guilds tab should be selected")
        XCTAssertTrue(guildListPage.isDisplayed(), "Guild list should be displayed")
    }
}
