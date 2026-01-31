import XCTest

/// Debug test to capture guild loading errors

@MainActor
final class GuildLoadingDebugTest: SagaUITestCase {

    /// Test that captures the error message when guilds fail to load
    func testCaptureGuildLoadingError() throws {
        app.launch()

        // Login
        let authPage = AuthPage(app: app)
        XCTAssertTrue(authPage.isDisplayed(), "Auth screen should be displayed")

        _ = authPage.loginAs(.demo)

        // Wait for the Guilds navigation bar
        let guildsNav = app.navigationBars["Guilds"]
        XCTAssertTrue(guildsNav.waitForExistence(timeout: TestConfig.longTimeout), "Should reach Guilds screen")

        // Wait a bit for any error to appear
        Wait.seconds(3.0)

        // Check for any error alerts
        let alertExists = app.alerts.firstMatch.waitForExistence(timeout: 5.0)
        if alertExists {
            // Capture all text in the alert
            let alertTexts = app.alerts.firstMatch.staticTexts.allElementsBoundByIndex
                .compactMap { $0.exists ? $0.label : nil }
            let alertMessage = alertTexts.joined(separator: " | ")

            // Log the error and fail the test with details
            XCTFail("Error alert appeared: \(alertMessage)")
        }

        // Check what's visible on screen
        let visibleTexts = app.staticTexts.allElementsBoundByIndex
            .prefix(15)
            .compactMap { $0.exists ? $0.label : nil }
            .filter { !$0.isEmpty }

        // Check for guilds
        let guildList = app.collectionViews["guild_list"]
        let hasGuilds = guildList.exists && guildList.cells.count > 0
        let hasEmptyState = app.staticTexts["No Guilds Yet"].exists

        if !hasGuilds && hasEmptyState {
            XCTFail("Guilds did not load. Empty state shown. Visible: \(visibleTexts.joined(separator: ", "))")
        } else if hasGuilds {
            // Success - guilds loaded
            XCTAssertTrue(true, "Guilds loaded successfully")
        } else {
            XCTFail("Unexpected state. Visible: \(visibleTexts.joined(separator: ", "))")
        }
    }
}
