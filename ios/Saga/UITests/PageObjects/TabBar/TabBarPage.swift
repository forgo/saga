import XCTest

/// Page object for the main tab bar navigation
@MainActor
class TabBarPage: BasePage {

    // MARK: - Elements

    /// The tab bar container
    var tabBar: XCUIElement {
        app.tabBars.firstMatch
    }

    /// Guilds tab button (uses label text as SwiftUI tab buttons don't support custom accessibility IDs)
    var guildsTab: XCUIElement {
        tabBar.buttons["Guilds"]
    }

    /// Events tab button
    var eventsTab: XCUIElement {
        tabBar.buttons["Events"]
    }

    /// Discover tab button
    var discoverTab: XCUIElement {
        tabBar.buttons["Discover"]
    }

    /// Profile tab button
    var profileTab: XCUIElement {
        tabBar.buttons["Profile"]
    }

    // MARK: - State Checks

    /// Check if tab bar is displayed
    func isDisplayed() -> Bool {
        tabBar.waitForExistence(timeout: TestConfig.defaultTimeout)
    }

    /// Check if Guilds tab is selected
    func isGuildsSelected() -> Bool {
        guildsTab.isSelected
    }

    /// Check if Events tab is selected
    func isEventsSelected() -> Bool {
        eventsTab.isSelected
    }

    /// Check if Discover tab is selected
    func isDiscoverSelected() -> Bool {
        discoverTab.isSelected
    }

    /// Check if Profile tab is selected
    func isProfileSelected() -> Bool {
        profileTab.isSelected
    }

    // MARK: - Navigation Actions

    /// Navigate to Guilds tab
    @discardableResult
    func tapGuilds() -> GuildListPage {
        guildsTab.tap()
        return GuildListPage(app: app)
    }

    /// Navigate to Events tab
    @discardableResult
    func tapEvents() -> EventListPage {
        _ = eventsTab.waitForExistence(timeout: TestConfig.shortTimeout)
        eventsTab.tap()
        Wait.seconds(0.5)  // Wait for tab selection to update
        return EventListPage(app: app)
    }

    /// Navigate to Discover tab
    @discardableResult
    func tapDiscover() -> BasePage {
        _ = discoverTab.waitForExistence(timeout: TestConfig.shortTimeout)
        discoverTab.tap()
        Wait.seconds(0.5)  // Wait for tab selection to update
        // Return DiscoverPage when implemented
        return BasePage(app: app)
    }

    /// Navigate to Profile tab
    @discardableResult
    func tapProfile() -> BasePage {
        _ = profileTab.waitForExistence(timeout: TestConfig.shortTimeout)
        profileTab.tap()
        Wait.seconds(0.5)  // Wait for tab selection to update
        // Return ProfilePage when implemented
        return BasePage(app: app)
    }

    // MARK: - Convenience

    /// Get the currently selected tab name
    func selectedTabName() -> String? {
        if isGuildsSelected() { return "Guilds" }
        if isEventsSelected() { return "Events" }
        if isDiscoverSelected() { return "Discover" }
        if isProfileSelected() { return "Profile" }
        return nil
    }
}
