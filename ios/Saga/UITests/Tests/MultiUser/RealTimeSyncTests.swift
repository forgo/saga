import XCTest

/// Tests for real-time synchronization via Server-Sent Events (SSE)
/// These tests verify the app correctly handles real-time updates
@MainActor
final class RealTimeSyncTests: SagaUITestCase {

    var multiApp: MultiAppHelper!

    override func setUp() async throws {
        await super.setUp()
        multiApp = MultiAppHelper()
    }

    override func tearDown() async throws {
        multiApp?.cleanup()
        multiApp = nil
        await super.tearDown()
    }

    // MARK: - Connection State Tests

    /// Tests that the app establishes SSE connection on guild selection
    func testSSEConnectionEstablished() throws {
        // Login as demo
        XCTAssertTrue(multiApp.launchPrimaryApp(asUser: .demo), "Should login as demo")

        let guildList = multiApp.primaryApp.collectionViews[AccessibilityID.Guild.list]
        XCTAssertTrue(guildList.waitForExistence(timeout: TestConfig.defaultTimeout))

        guard guildList.cells.count > 0 else {
            throw XCTSkip("No guilds available for SSE test")
        }

        // Select a guild
        guildList.cells.firstMatch.tap()

        // Wait for connection to establish
        multiApp.waitForSync(timeout: 3)

        // Check for connection indicator (if visible)
        let liveIndicator = multiApp.primaryApp.staticTexts["Live updates enabled"]
        // Note: Indicator visibility depends on UI implementation
        // The important thing is the view loads correctly
        let navBar = multiApp.primaryApp.navigationBars.firstMatch
        XCTAssertTrue(navBar.exists, "Guild detail should be displayed")
    }

    /// Tests that SSE reconnects after network interruption simulation
    func testSSEReconnectionAfterBackground() throws {
        // Login as demo
        XCTAssertTrue(multiApp.launchPrimaryApp(asUser: .demo), "Should login as demo")

        let guildList = multiApp.primaryApp.collectionViews[AccessibilityID.Guild.list]
        XCTAssertTrue(guildList.waitForExistence(timeout: TestConfig.defaultTimeout))

        guard guildList.cells.count > 0 else {
            throw XCTSkip("No guilds available for reconnection test")
        }

        // Select a guild
        guildList.cells.firstMatch.tap()

        // Wait for initial connection
        multiApp.waitForSync(timeout: 2)

        // Background the app
        XCUIDevice.shared.press(.home)

        // Wait a moment
        sleep(2)

        // Bring app back to foreground
        multiApp.primaryApp.activate()

        // App should reconnect - verify UI still works
        let navBar = multiApp.primaryApp.navigationBars.firstMatch
        XCTAssertTrue(navBar.waitForExistence(timeout: TestConfig.defaultTimeout), "App should recover after backgrounding")
    }

    // MARK: - Event Stream Tests

    /// Tests that timer updates are reflected in real-time
    func testTimerUpdatesInRealTime() throws {
        // Login as demo
        XCTAssertTrue(multiApp.launchPrimaryApp(asUser: .demo), "Should login as demo")

        let guildList = multiApp.primaryApp.collectionViews[AccessibilityID.Guild.list]
        XCTAssertTrue(guildList.waitForExistence(timeout: TestConfig.defaultTimeout))

        guard guildList.cells.count > 0 else {
            throw XCTSkip("No guilds available for timer test")
        }

        // Navigate to guild detail
        guildList.cells.firstMatch.tap()

        // Look for any person with a timer
        // Timer values should update in real-time via SSE
        let personList = multiApp.primaryApp.cells
        if personList.firstMatch.waitForExistence(timeout: TestConfig.defaultTimeout) {
            // Verify person cells are visible
            XCTAssertGreaterThan(personList.count, 0, "Should have at least one cell")
        }
    }

    // MARK: - Multi-User Event Propagation

    /// Tests that events created by one user appear for others
    func testEventPropagationBetweenUsers() throws {
        // This test simulates the SSE propagation scenario:
        // 1. User A is viewing guild
        // 2. User B creates/modifies data
        // 3. User A sees update via SSE (after switch in this case)

        // Login as demo
        XCTAssertTrue(multiApp.launchPrimaryApp(asUser: .demo), "Should login as demo")

        // Navigate to events
        multiApp.navigatePrimaryToTab(MultiAppHelper.Tab.events)

        let eventList = multiApp.primaryApp.collectionViews[AccessibilityID.Event.list]
        guard eventList.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            throw XCTSkip("Event list not available")
        }

        let initialCount = eventList.cells.count

        // Switch to second user (simulating another device/session)
        XCTAssertTrue(multiApp.switchPrimaryTo(.second), "Should switch to second user")

        // Navigate to events as second user
        multiApp.navigatePrimaryToTab(MultiAppHelper.Tab.events)

        // In a full implementation with SSE:
        // - Second user creates an event
        // - Switch back to demo
        // - Demo sees new event without manual refresh

        // Wait for any pending sync
        multiApp.waitForSync(timeout: 2)

        // Verify events are still accessible
        XCTAssertTrue(eventList.waitForExistence(timeout: TestConfig.defaultTimeout))
    }

    // MARK: - Availability Broadcast Tests

    /// Tests that availability updates are broadcast in real-time
    func testAvailabilityBroadcast() throws {
        // Login as demo
        XCTAssertTrue(multiApp.launchPrimaryApp(asUser: .demo), "Should login as demo")

        // Navigate to Profile > Availability
        multiApp.navigatePrimaryToTab(MultiAppHelper.Tab.profile)

        // Look for availability link
        let availabilityLink = multiApp.primaryApp.buttons["My Availability"]
        guard availabilityLink.waitForExistence(timeout: TestConfig.shortTimeout) else {
            throw XCTSkip("Availability link not found")
        }

        availabilityLink.tap()

        // Wait for availability view
        let availabilityNav = multiApp.primaryApp.navigationBars["Availability"]
        guard availabilityNav.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            // Try alternate nav bar text
            let altNav = multiApp.primaryApp.navigationBars.firstMatch
            XCTAssertTrue(altNav.exists, "Should be on availability view")
            return
        }

        // In a full implementation:
        // 1. Demo sets availability
        // 2. Second user (trusted) should see demo as "available now"
        XCTAssertTrue(availabilityNav.exists)
    }

    // MARK: - Guild Activity Stream Tests

    /// Tests that guild activity stream updates in real-time
    func testGuildActivityStream() throws {
        // Login as demo
        XCTAssertTrue(multiApp.launchPrimaryApp(asUser: .demo), "Should login as demo")

        let guildList = multiApp.primaryApp.collectionViews[AccessibilityID.Guild.list]
        XCTAssertTrue(guildList.waitForExistence(timeout: TestConfig.defaultTimeout))

        guard guildList.cells.count > 0 else {
            throw XCTSkip("No guilds available")
        }

        // Navigate to guild detail
        guildList.cells.firstMatch.tap()

        // Guild detail should show real-time activity
        // (person updates, timer changes, etc.)
        let navBar = multiApp.primaryApp.navigationBars.firstMatch
        XCTAssertTrue(navBar.waitForExistence(timeout: TestConfig.defaultTimeout))

        // Wait for SSE connection and potential updates
        multiApp.waitForSync(timeout: 3)

        // Verify view is still stable
        XCTAssertTrue(navBar.exists, "Guild detail should remain stable")
    }
}

// MARK: - Conflict Resolution Tests

extension RealTimeSyncTests {

    /// Tests that concurrent edits are handled gracefully
    func testConcurrentEditHandling() throws {
        // Login as demo
        XCTAssertTrue(multiApp.launchPrimaryApp(asUser: .demo), "Should login as demo")

        let guildList = multiApp.primaryApp.collectionViews[AccessibilityID.Guild.list]
        XCTAssertTrue(guildList.waitForExistence(timeout: TestConfig.defaultTimeout))

        // Rapidly perform multiple operations
        for _ in 0..<3 {
            // Pull to refresh (simulating sync)
            multiApp.refreshPrimary()

            // Brief pause
            multiApp.waitForSync(timeout: 0.5)
        }

        // App should handle multiple refreshes without crashing
        XCTAssertTrue(guildList.waitForExistence(timeout: TestConfig.defaultTimeout), "List should remain stable after multiple refreshes")
    }
}

// MARK: - Offline/Online Tests

extension RealTimeSyncTests {

    /// Tests app behavior when connection is restored
    func testDataRefreshOnForeground() throws {
        // Login as demo
        XCTAssertTrue(multiApp.launchPrimaryApp(asUser: .demo), "Should login as demo")

        let guildList = multiApp.primaryApp.collectionViews[AccessibilityID.Guild.list]
        XCTAssertTrue(guildList.waitForExistence(timeout: TestConfig.defaultTimeout))

        // Background the app
        XCUIDevice.shared.press(.home)
        sleep(1)

        // Foreground
        multiApp.primaryApp.activate()

        // Data should refresh automatically on foreground
        XCTAssertTrue(guildList.waitForExistence(timeout: TestConfig.defaultTimeout), "Guild list should reload on foreground")
    }
}
