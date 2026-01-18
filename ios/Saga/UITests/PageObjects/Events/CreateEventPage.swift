import XCTest

/// Page object for the create event sheet
@MainActor
class CreateEventPage: BasePage {

    // MARK: - Elements

    /// Event title text field
    var titleField: XCUIElement {
        app.textFields[AccessibilityID.Event.titleField]
    }

    /// Event description text field
    var descriptionField: XCUIElement {
        app.textViews[AccessibilityID.Event.descriptionField]
    }

    /// Create button
    var createButton: XCUIElement {
        app.buttons[AccessibilityID.Event.createConfirmButton]
    }

    /// Cancel button
    var cancelButton: XCUIElement {
        app.buttons[AccessibilityID.Event.createCancelButton]
    }

    /// Navigation title
    var navigationTitle: XCUIElement {
        app.staticTexts["New Event"]
    }

    // MARK: - State Checks

    /// Check if create sheet is displayed
    func isDisplayed() -> Bool {
        titleField.waitForExistence(timeout: TestConfig.defaultTimeout)
    }

    /// Check if create button is enabled
    func isCreateEnabled() -> Bool {
        createButton.isEnabled
    }

    // MARK: - Actions

    /// Enter event title
    @discardableResult
    func enterTitle(_ title: String) -> CreateEventPage {
        _ = titleField.waitForExistence(timeout: TestConfig.defaultTimeout)
        titleField.tap()
        titleField.typeText(title)
        return self
    }

    /// Enter event description
    @discardableResult
    func enterDescription(_ description: String) -> CreateEventPage {
        descriptionField.tap()
        descriptionField.typeText(description)
        return self
    }

    /// Tap create to submit
    @discardableResult
    func tapCreate() -> EventListPage {
        app.dismissKeyboardIfPresent()
        createButton.tap()
        Wait.forNetwork()
        return EventListPage(app: app)
    }

    /// Tap cancel to dismiss
    @discardableResult
    func tapCancel() -> EventListPage {
        cancelButton.tap()
        return EventListPage(app: app)
    }

    /// Create an event with title and optional description
    @discardableResult
    func createEvent(title: String, description: String? = nil) -> EventListPage {
        enterTitle(title)
        if let desc = description {
            enterDescription(desc)
        }
        return tapCreate()
    }
}
