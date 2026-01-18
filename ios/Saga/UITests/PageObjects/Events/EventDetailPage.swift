import XCTest

/// Page object for the event detail screen
@MainActor
class EventDetailPage: BasePage {

    // MARK: - Elements

    /// Event detail container
    var detailView: XCUIElement {
        app.scrollViews[AccessibilityID.Event.detailView]
    }

    /// RSVP Going button
    var rsvpGoingButton: XCUIElement {
        app.buttons[AccessibilityID.Event.rsvpGoing]
    }

    /// RSVP Maybe button
    var rsvpMaybeButton: XCUIElement {
        app.buttons[AccessibilityID.Event.rsvpMaybe]
    }

    /// RSVP Not Going button
    var rsvpNotGoingButton: XCUIElement {
        app.buttons[AccessibilityID.Event.rsvpNotGoing]
    }

    /// Edit button (host only)
    var editButton: XCUIElement {
        app.buttons["Edit"]
    }

    /// Cancel event button (host only)
    var cancelEventButton: XCUIElement {
        app.buttons["Cancel Event"]
    }

    /// Back button
    var backButton: XCUIElement {
        app.navigationBars.buttons.element(boundBy: 0)
    }

    // MARK: - State Checks

    /// Check if event detail screen is displayed
    func isDisplayed() -> Bool {
        // Check if RSVP buttons exist (indicates loaded event detail)
        rsvpGoingButton.waitForExistence(timeout: TestConfig.defaultTimeout) ||
        app.scrollViews.firstMatch.waitForExistence(timeout: TestConfig.defaultTimeout)
    }

    /// Check if user has RSVP'd as Going
    func isGoingSelected() -> Bool {
        rsvpGoingButton.isSelected
    }

    /// Check if user has RSVP'd as Maybe
    func isMaybeSelected() -> Bool {
        rsvpMaybeButton.isSelected
    }

    /// Check if user has RSVP'd as Not Going
    func isNotGoingSelected() -> Bool {
        rsvpNotGoingButton.isSelected
    }

    /// Check if user is host (can edit/cancel)
    func isHost() -> Bool {
        editButton.exists
    }

    /// Check if event has title visible
    func hasTitle(_ title: String) -> Bool {
        app.staticTexts[title].exists
    }

    /// Check if event has location visible
    func hasLocation(_ location: String) -> Bool {
        app.staticTexts[location].exists
    }

    // MARK: - RSVP Actions

    /// RSVP as Going
    @discardableResult
    func rsvpGoing() -> EventDetailPage {
        rsvpGoingButton.tap()
        Wait.forNetwork()
        return self
    }

    /// RSVP as Maybe
    @discardableResult
    func rsvpMaybe() -> EventDetailPage {
        rsvpMaybeButton.tap()
        Wait.forNetwork()
        return self
    }

    /// RSVP as Not Going
    @discardableResult
    func rsvpNotGoing() -> EventDetailPage {
        rsvpNotGoingButton.tap()
        Wait.forNetwork()
        return self
    }

    // MARK: - Host Actions

    /// Tap edit button to open edit sheet
    @discardableResult
    func tapEdit() -> BasePage {
        editButton.tap()
        // Return EditEventPage when implemented
        return BasePage(app: app)
    }

    /// Tap cancel event button
    @discardableResult
    func tapCancelEvent() -> EventDetailPage {
        cancelEventButton.tap()
        return self
    }

    /// Confirm event cancellation in alert
    @discardableResult
    func confirmCancelEvent() -> EventListPage {
        let confirmButton = app.alerts.buttons["Cancel Event"]
        if confirmButton.waitForExistence(timeout: TestConfig.shortTimeout) {
            confirmButton.tap()
        }
        Wait.forNetwork()
        return EventListPage(app: app)
    }

    // MARK: - Navigation

    /// Go back to event list
    @discardableResult
    func goBack() -> EventListPage {
        backButton.tap()
        return EventListPage(app: app)
    }

    /// Pull to refresh event details
    func pullToRefresh() {
        let scrollView = app.scrollViews.firstMatch
        guard scrollView.waitForExistence(timeout: TestConfig.shortTimeout) else { return }
        let start = scrollView.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: 0.3))
        let end = scrollView.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: 0.9))
        start.press(forDuration: 0.1, thenDragTo: end)
        Wait.forNetwork()
    }
}
