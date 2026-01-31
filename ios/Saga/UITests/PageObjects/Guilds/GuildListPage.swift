import XCTest

/// Page object for the guild list screen
@MainActor
class GuildListPage: BasePage {

    // MARK: - Elements

    /// Navigation bar
    var navigationBar: XCUIElement {
        app.navigationBars["Guilds"]
    }

    /// Guild list (SwiftUI List can be tables or collectionViews depending on iOS version)
    var guildList: XCUIElement {
        // Try collectionViews first (iOS 16+), fall back to tables
        let collection = app.collectionViews[AccessibilityID.Guild.list]
        if collection.exists { return collection }
        return app.tables[AccessibilityID.Guild.list]
    }

    /// Create guild button in toolbar
    var createButton: XCUIElement {
        app.buttons[AccessibilityID.Guild.createButton]
    }

    /// Empty state view when no guilds exist
    var emptyState: XCUIElement {
        app.staticTexts[AccessibilityID.Guild.emptyState]
    }

    // MARK: - State Checks

    /// Check if guild list screen is displayed
    func isDisplayed() -> Bool {
        navigationBar.waitForExistence(timeout: TestConfig.defaultTimeout)
    }

    /// Check if guild list is empty
    func isEmpty() -> Bool {
        emptyState.exists || guildCount() == 0
    }

    /// Wait for guilds to load (either shows guilds or confirms empty state)
    /// Returns true if guilds are loaded, false if empty state or timeout
    @discardableResult
    func waitForGuildsToLoad(timeout: TimeInterval = TestConfig.longTimeout) -> Bool {
        // First wait for navigation bar
        guard navigationBar.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            return false
        }

        // Wait for network
        Wait.forNetwork()

        // Use polling to wait for either guilds or empty state to stabilize
        let deadline = Date().addingTimeInterval(timeout)
        while Date() < deadline {
            // Check if we have guild cells
            if guildList.exists && guildList.cells.count > 0 {
                return true
            }
            // Check for empty state
            if emptyState.exists {
                return false
            }
            // Also check for guild name text as backup (list element types vary)
            if app.staticTexts["Close Friends"].exists {
                return true
            }
            Wait.seconds(0.3)
        }

        return guildList.cells.count > 0
    }

    /// Get number of guilds in the list
    func guildCount() -> Int {
        guildList.cells.count
    }

    /// Check if a guild with specific ID exists
    func hasGuild(id: String) -> Bool {
        guildRow(id: id).exists
    }

    /// Check if a guild with name exists (searches by label)
    func hasGuild(named name: String, timeout: TimeInterval = TestConfig.defaultTimeout) -> Bool {
        // Search for the guild name text anywhere in the app (List element types vary by iOS version)
        app.staticTexts[name].waitForExistence(timeout: timeout)
    }

    // MARK: - Element Accessors

    /// Get guild row by ID
    func guildRow(id: String) -> XCUIElement {
        app.buttons[AccessibilityID.Guild.row(id: id)]
    }

    /// Get guild row by index
    func guildRow(at index: Int) -> XCUIElement {
        guildList.cells.element(boundBy: index)
    }

    /// Get all guild cells
    var allGuildCells: XCUIElementQuery {
        guildList.cells
    }

    // MARK: - Actions

    /// Tap on a guild by ID
    @discardableResult
    func tapGuild(id: String) -> BasePage {
        guildRow(id: id).tap()
        // Return GuildDetailPage when implemented
        return BasePage(app: app)
    }

    /// Tap on a guild by index
    @discardableResult
    func tapGuild(at index: Int) -> BasePage {
        guildRow(at: index).tap()
        // Return GuildDetailPage when implemented
        return BasePage(app: app)
    }

    /// Tap on a guild by name
    @discardableResult
    func tapGuild(named name: String) -> BasePage {
        guildList.cells.staticTexts[name].tap()
        // Return GuildDetailPage when implemented
        return BasePage(app: app)
    }

    /// Tap create guild button to open creation sheet
    @discardableResult
    func tapCreateGuild() -> CreateGuildPage {
        createButton.tap()
        return CreateGuildPage(app: app)
    }

    /// Pull to refresh the guild list
    func pullToRefresh() {
        // Wait for list to be ready
        guard guildList.waitForExistence(timeout: TestConfig.shortTimeout) else { return }

        // Use a slow drag gesture to trigger pull-to-refresh
        // SwiftUI's refreshable needs a slow pull, not a fast swipe
        let start = guildList.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: 0.3))
        let end = guildList.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: 0.9))
        start.press(forDuration: 0.1, thenDragTo: end)

        // Wait for the refresh to complete
        Wait.forNetwork()
        Wait.forNetwork() // Double wait for API response
    }

    /// Delete a guild by swiping and confirming
    func deleteGuild(at index: Int) {
        let cell = guildRow(at: index)
        cell.swipeLeft()

        let deleteButton = app.buttons["Delete"]
        if deleteButton.waitForExistence(timeout: TestConfig.shortTimeout) {
            deleteButton.tap()

            // Confirm deletion if alert appears
            let confirmButton = app.alerts.buttons["Delete"]
            if confirmButton.waitForExistence(timeout: TestConfig.shortTimeout) {
                confirmButton.tap()
            }
        }
    }

    /// Scroll to find a guild by name
    func scrollToGuild(named name: String) {
        let guildLabel = guildList.cells.staticTexts[name]
        scrollToElement(guildLabel, inContainer: guildList)
    }
}

// MARK: - Create Guild Page

/// Page object for the create guild sheet
@MainActor
class CreateGuildPage: BasePage {

    // MARK: - Elements

    /// Guild name text field
    var nameField: XCUIElement {
        app.textFields[AccessibilityID.Guild.nameField]
    }

    /// Guild description text field
    var descriptionField: XCUIElement {
        app.textFields[AccessibilityID.Guild.descriptionField]
    }

    /// Icon picker
    var iconPicker: XCUIElement {
        app.otherElements[AccessibilityID.Guild.iconPicker]
    }

    /// Color picker
    var colorPicker: XCUIElement {
        app.colorWells[AccessibilityID.Guild.colorPicker]
    }

    /// Create button
    var createButton: XCUIElement {
        app.buttons[AccessibilityID.Guild.createConfirmButton]
    }

    /// Cancel button
    var cancelButton: XCUIElement {
        app.buttons[AccessibilityID.Guild.cancelButton]
    }

    // MARK: - State Checks

    /// Check if create sheet is displayed
    func isDisplayed() -> Bool {
        nameField.waitForExistence(timeout: TestConfig.defaultTimeout)
    }

    /// Check if create button is enabled
    func isCreateEnabled() -> Bool {
        createButton.isEnabled
    }

    // MARK: - Actions

    /// Enter guild name
    @discardableResult
    func enterName(_ name: String) -> CreateGuildPage {
        // Wait for name field to exist before interacting
        _ = nameField.waitForExistence(timeout: TestConfig.defaultTimeout)
        nameField.tap()
        nameField.typeText(name)
        return self
    }

    /// Enter guild description
    @discardableResult
    func enterDescription(_ description: String) -> CreateGuildPage {
        descriptionField.tap()
        descriptionField.typeText(description)
        return self
    }

    /// Select an icon by index
    @discardableResult
    func selectIcon(at index: Int) -> CreateGuildPage {
        // Icons are buttons within the icon picker section
        let icons = app.buttons.matching(NSPredicate(format: "identifier BEGINSWITH 'icon_'"))
        if icons.count > index {
            icons.element(boundBy: index).tap()
        }
        return self
    }

    /// Tap create to submit
    @discardableResult
    func tapCreate() -> GuildListPage {
        app.dismissKeyboardIfPresent()
        createButton.tap()
        Wait.forNetwork()
        return GuildListPage(app: app)
    }

    /// Tap cancel to dismiss
    @discardableResult
    func tapCancel() -> GuildListPage {
        cancelButton.tap()
        return GuildListPage(app: app)
    }

    /// Create a guild with all details
    @discardableResult
    func createGuild(name: String, description: String? = nil) -> GuildListPage {
        enterName(name)
        if let desc = description {
            enterDescription(desc)
        }
        return tapCreate()
    }
}
