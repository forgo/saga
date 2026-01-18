import XCTest

/// Page object for the event list screen
@MainActor
class EventListPage: BasePage {

    // MARK: - Elements

    /// Navigation bar
    var navigationBar: XCUIElement {
        app.navigationBars["Events"]
    }

    /// Event list
    var eventList: XCUIElement {
        let collection = app.collectionViews[AccessibilityID.Event.list]
        if collection.exists { return collection }
        return app.tables[AccessibilityID.Event.list]
    }

    /// Filter picker (Upcoming/Past/All)
    var filterPicker: XCUIElement {
        app.segmentedControls[AccessibilityID.Event.filterPicker]
    }

    /// Create event button in toolbar
    var createButton: XCUIElement {
        app.buttons[AccessibilityID.Event.createButton]
    }

    /// Empty state view
    var emptyState: XCUIElement {
        app.staticTexts["No Upcoming Events"]
    }

    // MARK: - Filter Buttons

    var upcomingFilterButton: XCUIElement {
        filterPicker.buttons["Upcoming"]
    }

    var pastFilterButton: XCUIElement {
        filterPicker.buttons["Past"]
    }

    var allFilterButton: XCUIElement {
        filterPicker.buttons["All"]
    }

    // MARK: - State Checks

    /// Check if event list screen is displayed
    func isDisplayed() -> Bool {
        navigationBar.waitForExistence(timeout: TestConfig.defaultTimeout)
    }

    /// Check if event list is empty
    func isEmpty() -> Bool {
        emptyState.exists || eventCount() == 0
    }

    /// Get number of events in the list
    func eventCount() -> Int {
        eventList.cells.count
    }

    /// Check if an event with name exists
    func hasEvent(named name: String, timeout: TimeInterval = TestConfig.defaultTimeout) -> Bool {
        app.staticTexts[name].waitForExistence(timeout: timeout)
    }

    // MARK: - Element Accessors

    /// Get event row by ID
    func eventRow(id: String) -> XCUIElement {
        app.buttons[AccessibilityID.Event.row(id: id)]
    }

    /// Get event row by index
    func eventRow(at index: Int) -> XCUIElement {
        eventList.cells.element(boundBy: index)
    }

    // MARK: - Actions

    /// Select a filter option
    @discardableResult
    func selectFilter(_ filter: EventFilter) -> EventListPage {
        switch filter {
        case .upcoming:
            upcomingFilterButton.tap()
        case .past:
            pastFilterButton.tap()
        case .all:
            allFilterButton.tap()
        }
        Wait.seconds(0.5) // Wait for filter to apply
        return self
    }

    /// Tap on an event by index
    @discardableResult
    func tapEvent(at index: Int) -> EventDetailPage {
        eventRow(at: index).tap()
        return EventDetailPage(app: app)
    }

    /// Tap on an event by name
    @discardableResult
    func tapEvent(named name: String) -> EventDetailPage {
        eventList.cells.staticTexts[name].tap()
        return EventDetailPage(app: app)
    }

    /// Tap create event button to open creation sheet
    @discardableResult
    func tapCreateEvent() -> CreateEventPage {
        createButton.tap()
        return CreateEventPage(app: app)
    }

    /// Pull to refresh the event list
    func pullToRefresh() {
        guard eventList.waitForExistence(timeout: TestConfig.shortTimeout) else { return }
        let start = eventList.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: 0.3))
        let end = eventList.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: 0.9))
        start.press(forDuration: 0.1, thenDragTo: end)
        Wait.forNetwork()
    }

    // MARK: - Filter Enum

    enum EventFilter {
        case upcoming
        case past
        case all
    }
}
