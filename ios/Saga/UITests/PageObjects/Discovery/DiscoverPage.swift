import XCTest

/// Page object for the discovery screen
@MainActor
class DiscoverPage: BasePage {

    // MARK: - Elements

    /// Navigation bar
    var navigationBar: XCUIElement {
        app.navigationBars["Discover"]
    }

    /// Tab picker (People/Events/Interests)
    var tabPicker: XCUIElement {
        app.segmentedControls[AccessibilityID.Discovery.tabPicker]
    }

    /// People tab button
    var peopleTab: XCUIElement {
        tabPicker.buttons["People"]
    }

    /// Events tab button
    var eventsTab: XCUIElement {
        tabPicker.buttons["Events"]
    }

    /// Interests tab button
    var interestsTab: XCUIElement {
        tabPicker.buttons["Interests"]
    }

    /// Search button (in People tab)
    var searchButton: XCUIElement {
        app.buttons[AccessibilityID.Discovery.searchButton]
    }

    /// No People Found empty state
    var noPeopleFound: XCUIElement {
        app.staticTexts["No People Found"]
    }

    /// No Events Nearby empty state
    var noEventsNearby: XCUIElement {
        app.staticTexts["No Events Nearby"]
    }

    /// No Matches Yet empty state
    var noMatchesYet: XCUIElement {
        app.staticTexts["No Matches Yet"]
    }

    // MARK: - State Checks

    /// Check if discover screen is displayed
    func isDisplayed() -> Bool {
        navigationBar.waitForExistence(timeout: TestConfig.defaultTimeout)
    }

    /// Check if People tab is selected
    func isPeopleTabSelected() -> Bool {
        peopleTab.isSelected
    }

    /// Check if Events tab is selected
    func isEventsTabSelected() -> Bool {
        eventsTab.isSelected
    }

    /// Check if Interests tab is selected
    func isInterestsTabSelected() -> Bool {
        interestsTab.isSelected
    }

    // MARK: - Actions

    /// Select People tab
    @discardableResult
    func selectPeopleTab() -> DiscoverPage {
        peopleTab.tap()
        Wait.seconds(0.5)
        return self
    }

    /// Select Events tab
    @discardableResult
    func selectEventsTab() -> DiscoverPage {
        eventsTab.tap()
        Wait.seconds(0.5)
        return self
    }

    /// Select Interests tab
    @discardableResult
    func selectInterestsTab() -> DiscoverPage {
        interestsTab.tap()
        Wait.seconds(0.5)
        return self
    }

    /// Tap search button to search for people
    @discardableResult
    func tapSearch() -> DiscoverPage {
        searchButton.tap()
        Wait.forNetwork()
        return self
    }
}
