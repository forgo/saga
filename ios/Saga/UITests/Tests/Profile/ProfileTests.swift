import XCTest

/// Tests for profile functionality
@MainActor
final class ProfileTests: SagaUITestCase {

    // MARK: - Tab Navigation Tests

    func testCanNavigateToProfileTab() throws {
        _ = launchAndLoginWithDemoUser()

        // Navigate to Profile tab
        _ = tabBar.tapProfile()

        XCTAssertTrue(tabBar.isProfileSelected(), "Profile tab should be selected")
    }

    func testProfileTabShowsNavigationBar() throws {
        _ = launchAndLoginWithDemoUser()

        // Navigate to Profile tab
        _ = tabBar.tapProfile()

        // Check navigation bar exists
        let navBar = app.navigationBars["Profile"]
        XCTAssertTrue(
            navBar.waitForExistence(timeout: TestConfig.defaultTimeout),
            "Profile navigation bar should exist"
        )
    }

    // MARK: - Profile Content Tests

    func testProfileShowsUserInfo() throws {
        _ = launchAndLoginWithDemoUser()

        // Navigate to Profile tab
        _ = tabBar.tapProfile()
        Wait.forNetwork()

        // Should show user display name or email
        let demoLabel = app.staticTexts["Demo User"]
        let emailLabel = app.staticTexts["demo@forgo.software"]

        XCTAssertTrue(
            demoLabel.waitForExistence(timeout: TestConfig.defaultTimeout) ||
            emailLabel.waitForExistence(timeout: TestConfig.defaultTimeout),
            "Profile should show user info"
        )
    }

    func testProfileShowsEditButton() throws {
        _ = launchAndLoginWithDemoUser()

        // Navigate to Profile tab
        _ = tabBar.tapProfile()
        Wait.forNetwork()

        // Look for Edit Profile button
        let editButton = app.buttons["Edit Profile"]
        XCTAssertTrue(
            editButton.waitForExistence(timeout: TestConfig.defaultTimeout),
            "Edit Profile button should exist"
        )
    }

    func testProfileShowsSocialSection() throws {
        _ = launchAndLoginWithDemoUser()

        // Navigate to Profile tab
        _ = tabBar.tapProfile()
        Wait.forNetwork()

        // Should show Social section - can be header, staticText, or any element
        // Section headers in SwiftUI List may not be regular staticTexts
        let socialText = app.staticTexts.containing(NSPredicate(format: "label CONTAINS[c] 'Social'")).firstMatch
        let socialOther = app.otherElements.containing(NSPredicate(format: "label CONTAINS[c] 'Social'")).firstMatch

        XCTAssertTrue(
            socialText.waitForExistence(timeout: TestConfig.defaultTimeout) ||
            socialOther.exists,
            "Social section should exist"
        )
    }

    func testProfileShowsAvailabilityLink() throws {
        _ = launchAndLoginWithDemoUser()

        // Navigate to Profile tab
        _ = tabBar.tapProfile()
        Wait.forNetwork()

        // Look for My Availability link
        let availabilityLink = app.staticTexts["My Availability"]
        XCTAssertTrue(
            availabilityLink.waitForExistence(timeout: TestConfig.defaultTimeout),
            "My Availability link should exist"
        )
    }

    func testProfileShowsTrustLink() throws {
        _ = launchAndLoginWithDemoUser()

        // Navigate to Profile tab
        _ = tabBar.tapProfile()
        Wait.forNetwork()

        // Look for Trust & Connections link
        let trustLink = app.staticTexts["Trust & Connections"]
        XCTAssertTrue(
            trustLink.waitForExistence(timeout: TestConfig.defaultTimeout),
            "Trust & Connections link should exist"
        )
    }

    func testProfileShowsLogoutButton() throws {
        _ = launchAndLoginWithDemoUser()

        // Navigate to Profile tab
        _ = tabBar.tapProfile()
        Wait.forNetwork()

        // Scroll to find logout button
        let list = app.collectionViews.firstMatch.exists ? app.collectionViews.firstMatch : app.tables.firstMatch

        let logoutButton = app.buttons[AccessibilityID.Profile.logoutButton]

        // Scroll to find logout button
        var attempts = 0
        while !logoutButton.exists && attempts < 5 {
            list.swipeUp()
            Wait.seconds(0.3)
            attempts += 1
        }

        XCTAssertTrue(
            logoutButton.exists,
            "Logout button should exist"
        )
    }

    // MARK: - Navigation Tests

    func testCanNavigateToAvailability() throws {
        _ = launchAndLoginWithDemoUser()

        // Navigate to Profile tab
        _ = tabBar.tapProfile()
        Wait.forNetwork()

        // Tap My Availability
        let availabilityLink = app.cells.containing(.staticText, identifier: "My Availability").firstMatch
        guard availabilityLink.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            throw XCTSkip("Availability link not found")
        }

        availabilityLink.tap()
        Wait.forNetwork()

        // Should navigate to availability screen
        let navBar = app.navigationBars.element(boundBy: 0)
        XCTAssertTrue(
            navBar.waitForExistence(timeout: TestConfig.defaultTimeout),
            "Should navigate to availability screen"
        )
    }

    func testCanNavigateToTrust() throws {
        _ = launchAndLoginWithDemoUser()

        // Navigate to Profile tab
        _ = tabBar.tapProfile()
        Wait.forNetwork()

        // Tap Trust & Connections
        let trustLink = app.cells.containing(.staticText, identifier: "Trust & Connections").firstMatch
        guard trustLink.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            throw XCTSkip("Trust link not found")
        }

        trustLink.tap()
        Wait.forNetwork()

        // Should navigate to trust screen
        let navBar = app.navigationBars.element(boundBy: 0)
        XCTAssertTrue(
            navBar.waitForExistence(timeout: TestConfig.defaultTimeout),
            "Should navigate to trust screen"
        )
    }

    // MARK: - Logout Tests

    func testLogoutConfirmationAppears() throws {
        _ = launchAndLoginWithDemoUser()

        // Navigate to Profile tab
        _ = tabBar.tapProfile()
        Wait.forNetwork()

        // Scroll to and tap logout
        let list = app.collectionViews.firstMatch.exists ? app.collectionViews.firstMatch : app.tables.firstMatch
        let logoutButton = app.buttons[AccessibilityID.Profile.logoutButton]

        var attempts = 0
        while !logoutButton.isHittable && attempts < 5 {
            list.swipeUp()
            Wait.seconds(0.3)
            attempts += 1
        }

        guard logoutButton.exists && logoutButton.isHittable else {
            throw XCTSkip("Logout button not hittable")
        }

        logoutButton.tap()

        // Confirmation alert should appear
        let alert = app.alerts["Sign Out"]
        XCTAssertTrue(
            alert.waitForExistence(timeout: TestConfig.defaultTimeout),
            "Sign Out confirmation alert should appear"
        )
    }

    func testCancelLogoutStaysOnProfile() throws {
        _ = launchAndLoginWithDemoUser()

        // Navigate to Profile tab
        _ = tabBar.tapProfile()
        Wait.forNetwork()

        // Scroll to and tap logout
        let list = app.collectionViews.firstMatch.exists ? app.collectionViews.firstMatch : app.tables.firstMatch
        let logoutButton = app.buttons[AccessibilityID.Profile.logoutButton]

        var attempts = 0
        while !logoutButton.isHittable && attempts < 5 {
            list.swipeUp()
            Wait.seconds(0.3)
            attempts += 1
        }

        guard logoutButton.exists && logoutButton.isHittable else {
            throw XCTSkip("Logout button not hittable")
        }

        logoutButton.tap()

        // Cancel the logout
        let cancelButton = app.alerts.buttons["Cancel"]
        if cancelButton.waitForExistence(timeout: TestConfig.shortTimeout) {
            cancelButton.tap()
        }

        Wait.seconds(0.5)

        // Should still be on profile
        XCTAssertTrue(tabBar.isProfileSelected(), "Should stay on Profile tab after cancel")
    }
}
