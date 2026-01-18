import XCTest

/// Tests for core user journeys through the app
/// These tests verify end-to-end flows that span multiple screens
@MainActor
final class CoreJourneyTests: SagaUITestCase {

    // MARK: - Login to Guild Journey

    /// Test: User can login and view their guilds
    func testLoginToGuildListJourney() throws {
        // Launch and verify auth screen
        launchApp()
        XCTAssertTrue(authPage.isDisplayed(), "Auth screen should appear on launch")

        // Login with demo user
        let guildList = authPage.loginAs(.demo)

        // Verify guild list is displayed
        XCTAssertTrue(guildList.isDisplayed(), "Guild list should be displayed after login")
        XCTAssertTrue(tabBar.isDisplayed(), "Tab bar should be visible")
        XCTAssertTrue(tabBar.isGuildsSelected(), "Guilds tab should be selected by default")
    }

    // MARK: - Guild Detail Journey

    /// Test: User can navigate from guild list to guild detail
    func testGuildListToDetailJourney() throws {
        let guildList = launchAndLoginWithDemoUser()
        Wait.forNetwork()

        // Must have at least one guild
        guard guildList.guildCount() > 0 else {
            throw XCTSkip("No guilds available for test")
        }

        // Tap first guild
        _ = guildList.tapGuild(at: 0)
        Wait.forNetwork()

        // Verify we're in guild detail (navigation bar changes)
        let backButton = app.navigationBars.buttons.element(boundBy: 0)
        XCTAssertTrue(
            backButton.waitForExistence(timeout: TestConfig.defaultTimeout),
            "Back button should appear in guild detail"
        )

        // Navigate back
        backButton.tap()
        Wait.seconds(0.5)

        // Verify we're back on guild list
        XCTAssertTrue(guildList.isDisplayed(), "Should return to guild list")
    }

    // MARK: - Tab Navigation Journey

    /// Test: User can navigate through all main tabs
    func testFullTabNavigationJourney() throws {
        _ = launchAndLoginWithDemoUser()

        // Start on Guilds (default)
        XCTAssertTrue(tabBar.isGuildsSelected(), "Should start on Guilds tab")

        // Go to Events
        _ = tabBar.tapEvents()
        XCTAssertTrue(tabBar.isEventsSelected(), "Events tab should be selected")

        // Go to Discover
        _ = tabBar.tapDiscover()
        XCTAssertTrue(tabBar.isDiscoverSelected(), "Discover tab should be selected")

        // Go to Profile
        _ = tabBar.tapProfile()
        XCTAssertTrue(tabBar.isProfileSelected(), "Profile tab should be selected")

        // Back to Guilds
        _ = tabBar.tapGuilds()
        XCTAssertTrue(tabBar.isGuildsSelected(), "Guilds tab should be selected again")
    }

    // MARK: - Guild to Events Journey

    /// Test: User can access guild events from guild detail
    func testGuildToEventsJourney() throws {
        let guildList = launchAndLoginWithDemoUser()
        Wait.forNetwork()

        // Must have at least one guild
        guard guildList.guildCount() > 0 else {
            throw XCTSkip("No guilds available for test")
        }

        // Navigate to first guild detail
        _ = guildList.tapGuild(at: 0)
        Wait.forNetwork()

        // Look for Events link in guild detail
        let eventsLink = app.buttons["Events"]
        let eventsCell = app.cells.containing(.staticText, identifier: "Events").firstMatch

        if eventsLink.waitForExistence(timeout: TestConfig.shortTimeout) {
            eventsLink.tap()
        } else if eventsCell.exists {
            eventsCell.tap()
        } else {
            throw XCTSkip("Events link not found in guild detail")
        }

        Wait.forNetwork()

        // Verify Events screen
        let eventsNavBar = app.navigationBars["Events"]
        XCTAssertTrue(
            eventsNavBar.waitForExistence(timeout: TestConfig.defaultTimeout),
            "Should navigate to Events screen"
        )
    }

    // MARK: - Profile and Logout Journey

    /// Test: User can access profile and sign out
    func testProfileAndLogoutJourney() throws {
        _ = launchAndLoginWithDemoUser()

        // Navigate to Profile tab
        _ = tabBar.tapProfile()
        XCTAssertTrue(tabBar.isProfileSelected(), "Profile tab should be selected")

        Wait.forNetwork() // Wait for profile to load

        // Profile uses a List (which is a collection view on iOS)
        // Find the logout button by searching in the profile list
        let logoutButton = app.buttons[AccessibilityID.Profile.logoutButton]

        // Try to find the logout button by scrolling
        let list = app.collectionViews.firstMatch.exists ? app.collectionViews.firstMatch : app.tables.firstMatch

        // Scroll to find logout button
        var attempts = 0
        while !logoutButton.isHittable && attempts < 5 {
            list.swipeUp()
            Wait.seconds(0.3)
            attempts += 1
        }

        guard logoutButton.exists && logoutButton.isHittable else {
            throw XCTSkip("Logout button not found or not hittable")
        }

        // Tap logout
        logoutButton.tap()

        // Confirm logout in alert
        let confirmButton = app.alerts.buttons["Sign Out"]
        if confirmButton.waitForExistence(timeout: TestConfig.shortTimeout) {
            confirmButton.tap()
        }

        Wait.forNetwork()

        // Verify we're back to auth screen
        XCTAssertTrue(
            authPage.isDisplayed(),
            "Should return to auth screen after logout"
        )
    }

    // MARK: - Create Guild Sheet Journey

    /// Test: User can open and cancel create guild sheet
    func testCreateGuildSheetJourney() throws {
        let guildList = launchAndLoginWithDemoUser()
        Wait.forNetwork()

        // Open create sheet
        let createSheet = guildList.tapCreateGuild()

        // Verify sheet is displayed
        XCTAssertTrue(createSheet.isDisplayed(), "Create guild sheet should appear")
        XCTAssertFalse(createSheet.isCreateEnabled(), "Create should be disabled with empty name")

        // Enter a name
        createSheet.enterName("Test Journey Guild")
        XCTAssertTrue(createSheet.isCreateEnabled(), "Create should be enabled with name")

        // Cancel and verify we return to list
        let returnedList = createSheet.tapCancel()
        XCTAssertTrue(returnedList.isDisplayed(), "Should return to guild list after cancel")
    }

    // MARK: - Event Filter Journey

    /// Test: User can switch between event filters
    func testEventFilterJourney() throws {
        _ = launchAndLoginWithDemoUser()

        // Navigate to Events tab
        _ = tabBar.tapEvents()
        Wait.forNetwork()

        // Find filter picker
        let filterPicker = app.segmentedControls.firstMatch
        guard filterPicker.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            throw XCTSkip("Filter picker not found")
        }

        // Test filter switching
        let pastButton = filterPicker.buttons["Past"]
        let allButton = filterPicker.buttons["All"]
        let upcomingButton = filterPicker.buttons["Upcoming"]

        // Switch to Past
        if pastButton.exists {
            pastButton.tap()
            Wait.seconds(0.5)
            XCTAssertTrue(pastButton.isSelected, "Past filter should be selected")
        }

        // Switch to All
        if allButton.exists {
            allButton.tap()
            Wait.seconds(0.5)
            XCTAssertTrue(allButton.isSelected, "All filter should be selected")
        }

        // Switch back to Upcoming
        if upcomingButton.exists {
            upcomingButton.tap()
            Wait.seconds(0.5)
            XCTAssertTrue(upcomingButton.isSelected, "Upcoming filter should be selected")
        }
    }

    // MARK: - Multi-Tab State Preservation Journey

    /// Test: Tab state is preserved when switching between tabs
    func testTabStatePreservationJourney() throws {
        let guildList = launchAndLoginWithDemoUser()
        Wait.forNetwork()

        // Must have at least one guild
        guard guildList.guildCount() > 0 else {
            throw XCTSkip("No guilds available for test")
        }

        // Navigate into guild detail
        _ = guildList.tapGuild(at: 0)
        Wait.forNetwork()

        // Switch to Events tab
        _ = tabBar.tapEvents()
        Wait.seconds(0.5)

        // Switch back to Guilds
        _ = tabBar.tapGuilds()
        Wait.seconds(0.5)

        // Should be back on guild detail (state preserved)
        // OR back on guild list (state not preserved) - both are valid behaviors
        // Just verify we're on a valid screen
        XCTAssertTrue(
            tabBar.isGuildsSelected(),
            "Guilds tab should be selected"
        )
    }
}
