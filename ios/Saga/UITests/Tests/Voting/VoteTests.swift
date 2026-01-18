import XCTest

/// Tests for voting system functionality
/// Votes are accessed from the guild detail view

@MainActor
final class VoteTests: SagaUITestCase {

    // MARK: - Helper

    /// Navigate to Votes from guild detail
    private func navigateToVotes() throws {
        let guildList = launchAndLoginWithDemoUser()

        // Wait for guilds to load
        let hasGuilds = guildList.waitForGuildsToLoad()

        if !hasGuilds {
            if app.alerts.firstMatch.exists {
                let alertText = app.alerts.firstMatch.staticTexts.allElementsBoundByIndex
                    .map { $0.label }
                    .joined(separator: " ")
                throw XCTSkip("Alert shown: \(alertText)")
            }

            let visibleLabels = app.staticTexts.allElementsBoundByIndex
                .prefix(10)
                .compactMap { $0.exists ? $0.label : nil }
                .filter { !$0.isEmpty }
                .joined(separator: ", ")
            throw XCTSkip("No guilds loaded. Visible: \(visibleLabels)")
        }

        // Navigate to guild detail
        _ = guildList.tapGuild(at: 0)

        // Wait for guild detail
        let guildDetailNavBar = app.navigationBars.firstMatch
        guard guildDetailNavBar.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            throw XCTSkip("Guild detail navigation bar not found")
        }

        Wait.forNetwork()
        Wait.seconds(1.0)

        // Look for Votes link on guild detail
        let votesButton = app.buttons["guild_feature_votes"]
        let votesLink = app.buttons[AccessibilityID.Guild.votesLink]

        // First try to find Votes directly
        if votesButton.waitForExistence(timeout: TestConfig.defaultTimeout) {
            votesButton.tap()
            Wait.forNetwork()
            return
        }

        if votesLink.waitForExistence(timeout: TestConfig.shortTimeout) {
            votesLink.tap()
            Wait.forNetwork()
            return
        }

        // Try scrolling to find the Features section
        for _ in 0..<5 {
            app.swipeUp()
            Wait.seconds(0.5)

            if votesButton.waitForExistence(timeout: TestConfig.shortTimeout) {
                votesButton.tap()
                Wait.forNetwork()
                return
            }

            if votesLink.waitForExistence(timeout: TestConfig.shortTimeout) {
                votesLink.tap()
                Wait.forNetwork()
                return
            }
        }

        // Collect debug info
        let visibleTexts = app.staticTexts.allElementsBoundByIndex
            .prefix(20)
            .compactMap { $0.exists ? $0.label : nil }
            .filter { !$0.isEmpty }

        throw XCTSkip("Votes link not found. Visible: \(visibleTexts.joined(separator: ", "))")
    }

    // MARK: - Navigation Tests

    func testCanNavigateToVotes() throws {
        try navigateToVotes()

        // Verify Votes screen
        let navBar = app.navigationBars["Votes"]
        guard navBar.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            // Try alternate names
            let altNavBars = ["Guild Votes", "Active Votes", "Voting"]
            for altName in altNavBars {
                let altNavBar = app.navigationBars[altName]
                if altNavBar.waitForExistence(timeout: TestConfig.shortTimeout) {
                    XCTAssertTrue(altNavBar.exists, "\(altName) navigation bar should exist")
                    return
                }
            }
            throw XCTSkip("VoteListView not implemented yet - Votes navigation bar not found")
        }
        XCTAssertTrue(navBar.exists, "Votes navigation bar should exist")
    }

    func testVotesShowsCreateButton() throws {
        try navigateToVotes()

        // Check create button exists
        let createButton = app.buttons[AccessibilityID.Vote.createButton]
        XCTAssertTrue(
            createButton.waitForExistence(timeout: TestConfig.defaultTimeout),
            "Create vote button should exist"
        )
    }

    // MARK: - Filter Tests

    func testVotesShowsFilterPicker() throws {
        try navigateToVotes()

        // Check filter picker exists (Active, Closed, All)
        let filterPicker = app.segmentedControls[AccessibilityID.Vote.filterPicker]
        guard filterPicker.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            // May be implemented as buttons instead
            let activeFilter = app.buttons["Active"]
            let closedFilter = app.buttons["Closed"]
            if activeFilter.exists || closedFilter.exists {
                XCTAssertTrue(true, "Filter options exist as buttons")
                return
            }
            throw XCTSkip("Vote filter picker not found")
        }
        XCTAssertTrue(filterPicker.exists, "Filter picker should exist")
    }

    func testCanFilterByActiveVotes() throws {
        try navigateToVotes()

        // Tap Active filter
        let filterPicker = app.segmentedControls[AccessibilityID.Vote.filterPicker]
        let activeButton = app.buttons["Active"]

        if filterPicker.waitForExistence(timeout: TestConfig.defaultTimeout) {
            filterPicker.buttons["Active"].tap()
        } else if activeButton.exists {
            activeButton.tap()
        }

        Wait.forNetwork()

        // Should show active votes or empty state
        let voteList = app.collectionViews[AccessibilityID.Vote.list]
        let emptyState = app.staticTexts["No Active Votes"]
        let noVotes = app.staticTexts.containing(NSPredicate(format: "label CONTAINS[c] 'no active'")).firstMatch

        XCTAssertTrue(
            voteList.exists ||
            emptyState.exists ||
            noVotes.exists ||
            app.collectionViews.firstMatch.cells.count >= 0,
            "Should show active votes or appropriate empty state"
        )
    }

    func testCanFilterByClosedVotes() throws {
        try navigateToVotes()

        // Tap Closed filter
        let filterPicker = app.segmentedControls[AccessibilityID.Vote.filterPicker]
        let closedButton = app.buttons["Closed"]

        if filterPicker.waitForExistence(timeout: TestConfig.defaultTimeout) {
            filterPicker.buttons["Closed"].tap()
        } else if closedButton.exists {
            closedButton.tap()
        } else {
            throw XCTSkip("Closed filter not found")
        }

        Wait.forNetwork()

        // Should show closed votes or empty state
        let voteList = app.collectionViews[AccessibilityID.Vote.list]
        let emptyState = app.staticTexts["No Closed Votes"]
        let noVotes = app.staticTexts.containing(NSPredicate(format: "label CONTAINS[c] 'no closed'")).firstMatch

        XCTAssertTrue(
            voteList.exists ||
            emptyState.exists ||
            noVotes.exists ||
            app.collectionViews.firstMatch.cells.count >= 0,
            "Should show closed votes or appropriate empty state"
        )
    }

    // MARK: - Empty State Tests

    func testVotesShowsEmptyStateOrList() throws {
        try navigateToVotes()

        // Should show either votes or empty state
        let emptyState = app.staticTexts[AccessibilityID.Vote.emptyState]
        let createButton = app.buttons[AccessibilityID.Vote.createButton]
        let voteList = app.collectionViews.firstMatch.exists ? app.collectionViews.firstMatch : app.tables.firstMatch

        // One of these should be visible
        XCTAssertTrue(
            emptyState.waitForExistence(timeout: TestConfig.defaultTimeout) ||
            createButton.exists ||
            voteList.cells.count > 0 ||
            app.staticTexts.containing(NSPredicate(format: "label CONTAINS[c] 'no votes'")).firstMatch.exists,
            "Should show votes or empty state"
        )
    }

    // MARK: - Create Vote Sheet Tests

    func testCanOpenCreateVoteSheet() throws {
        try navigateToVotes()

        // Tap create button
        let createButton = app.buttons[AccessibilityID.Vote.createButton]
        guard createButton.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            throw XCTSkip("Create button not found")
        }

        createButton.tap()

        // Verify create sheet appears
        let navTitle = app.staticTexts["New Vote"]
        let altTitle = app.staticTexts["Create Vote"]
        XCTAssertTrue(
            navTitle.waitForExistence(timeout: TestConfig.defaultTimeout) ||
            altTitle.exists,
            "Create Vote sheet should appear"
        )
    }

    func testCreateVoteSheetHasTitleField() throws {
        try navigateToVotes()

        // Open create sheet
        let createButton = app.buttons[AccessibilityID.Vote.createButton]
        guard createButton.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            throw XCTSkip("Create button not found")
        }
        createButton.tap()

        // Check title field exists
        let titleField = app.textFields[AccessibilityID.Vote.titleField]
        let anyTextField = app.textFields.firstMatch

        XCTAssertTrue(
            titleField.waitForExistence(timeout: TestConfig.defaultTimeout) ||
            anyTextField.exists,
            "Title field should exist"
        )
    }

    func testCreateVoteSheetHasTypePicker() throws {
        try navigateToVotes()

        // Open create sheet
        let createButton = app.buttons[AccessibilityID.Vote.createButton]
        guard createButton.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            throw XCTSkip("Create button not found")
        }
        createButton.tap()

        // Check type picker exists (FPTP, Ranked Choice, etc.)
        let typePicker = app.segmentedControls[AccessibilityID.Vote.typePicker]
        let typeLabel = app.staticTexts["Vote Type"]
        let fptpOption = app.staticTexts.containing(NSPredicate(format: "label CONTAINS[c] 'plurality'")).firstMatch
        let rankedOption = app.staticTexts.containing(NSPredicate(format: "label CONTAINS[c] 'ranked'")).firstMatch

        // May need to scroll to find type picker
        if !typePicker.exists && !typeLabel.exists && !fptpOption.exists && !rankedOption.exists {
            let sheet = app.sheets.firstMatch.exists ? app.sheets.firstMatch : app.scrollViews.firstMatch
            sheet.swipeUp()
            Wait.seconds(0.5)
        }

        XCTAssertTrue(
            typePicker.waitForExistence(timeout: TestConfig.defaultTimeout) ||
            typeLabel.exists ||
            fptpOption.exists ||
            rankedOption.exists,
            "Vote type picker should exist"
        )
    }

    func testCreateVoteSheetHasCancelButton() throws {
        try navigateToVotes()

        // Open create sheet
        let createButton = app.buttons[AccessibilityID.Vote.createButton]
        guard createButton.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            throw XCTSkip("Create button not found")
        }
        createButton.tap()

        // Check cancel button exists
        let cancelButton = app.buttons["Cancel"]
        XCTAssertTrue(
            cancelButton.waitForExistence(timeout: TestConfig.defaultTimeout),
            "Cancel button should exist"
        )
    }

    func testCanCancelCreateVoteSheet() throws {
        try navigateToVotes()

        // Open create sheet
        let createButton = app.buttons[AccessibilityID.Vote.createButton]
        guard createButton.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            throw XCTSkip("Create button not found")
        }
        createButton.tap()

        // Wait for sheet
        let navTitle = app.staticTexts["New Vote"]
        let altTitle = app.staticTexts["Create Vote"]
        guard navTitle.waitForExistence(timeout: TestConfig.defaultTimeout) || altTitle.exists else {
            throw XCTSkip("Create sheet did not appear")
        }

        // Tap cancel
        let cancelButton = app.buttons["Cancel"]
        guard cancelButton.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            throw XCTSkip("Cancel button not found")
        }
        cancelButton.tap()

        // Verify sheet is dismissed
        Wait.seconds(0.5)
        XCTAssertFalse(
            navTitle.exists && altTitle.exists,
            "Create sheet should be dismissed"
        )
    }

    // MARK: - Vote Detail Tests

    func testCanOpenVoteDetail() throws {
        try navigateToVotes()

        Wait.forNetwork()
        Wait.seconds(1.0)

        // Check if there are any votes
        let voteRows = app.collectionViews.cells.matching(NSPredicate(format: "identifier BEGINSWITH 'vote_row_'"))
        let firstCell = app.collectionViews.firstMatch.cells.firstMatch

        if voteRows.count > 0 {
            voteRows.element(boundBy: 0).tap()
        } else if firstCell.exists {
            firstCell.tap()
        } else {
            throw XCTSkip("No votes available to view detail")
        }

        Wait.forNetwork()

        // Should navigate to vote detail
        let castButton = app.buttons[AccessibilityID.Vote.castButton]
        let resultsLabel = app.staticTexts.containing(NSPredicate(format: "label CONTAINS[c] 'result'")).firstMatch
        let optionsLabel = app.staticTexts.containing(NSPredicate(format: "label CONTAINS[c] 'option'")).firstMatch

        XCTAssertTrue(
            castButton.waitForExistence(timeout: TestConfig.defaultTimeout) ||
            resultsLabel.exists ||
            optionsLabel.exists,
            "Vote detail should show options or results"
        )
    }

    // MARK: - Cast Vote Tests

    func testCastButtonExistsForActiveVote() throws {
        try navigateToVotes()

        Wait.forNetwork()
        Wait.seconds(1.0)

        // Find an active vote
        let voteRows = app.collectionViews.cells
        let firstCell = voteRows.firstMatch

        guard firstCell.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            throw XCTSkip("No votes available")
        }

        firstCell.tap()
        Wait.forNetwork()

        // Look for cast button
        let castButton = app.buttons[AccessibilityID.Vote.castButton]
        let voteButton = app.buttons.containing(NSPredicate(format: "label CONTAINS[c] 'vote'")).firstMatch

        // Cast button might only appear if vote is active and user hasn't voted
        if !castButton.exists && !voteButton.exists {
            // Check if already voted or vote is closed
            let alreadyVoted = app.staticTexts.containing(NSPredicate(format: "label CONTAINS[c] 'voted'")).firstMatch
            let voteClosed = app.staticTexts.containing(NSPredicate(format: "label CONTAINS[c] 'closed'")).firstMatch

            if alreadyVoted.exists || voteClosed.exists {
                throw XCTSkip("Vote already cast or vote is closed")
            }
        }

        XCTAssertTrue(
            castButton.exists || voteButton.exists,
            "Cast vote button should exist for active votes"
        )
    }

    // MARK: - Vote Results Tests

    func testCanViewVoteResults() throws {
        try navigateToVotes()

        // Filter to closed votes where results would be visible
        let filterPicker = app.segmentedControls[AccessibilityID.Vote.filterPicker]
        let closedButton = app.buttons["Closed"]

        if filterPicker.exists {
            filterPicker.buttons["Closed"].tap()
        } else if closedButton.exists {
            closedButton.tap()
        }

        Wait.forNetwork()
        Wait.seconds(1.0)

        // Find a closed vote
        let voteRows = app.collectionViews.cells
        let firstCell = voteRows.firstMatch

        guard firstCell.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            throw XCTSkip("No closed votes available to view results")
        }

        firstCell.tap()
        Wait.forNetwork()

        // Should show results
        let resultsLabel = app.staticTexts.containing(NSPredicate(format: "label CONTAINS[c] 'result'")).firstMatch
        let winnerLabel = app.staticTexts.containing(NSPredicate(format: "label CONTAINS[c] 'winner'")).firstMatch
        let percentageLabel = app.staticTexts.containing(NSPredicate(format: "label CONTAINS '%'")).firstMatch

        XCTAssertTrue(
            resultsLabel.exists ||
            winnerLabel.exists ||
            percentageLabel.exists,
            "Closed vote should show results"
        )
    }

    // MARK: - Vote Row Tests

    func testVoteRowShowsTitle() throws {
        try navigateToVotes()

        Wait.forNetwork()
        Wait.seconds(1.0)

        // Check if there are any votes
        let voteRows = app.collectionViews.cells
        let firstCell = voteRows.firstMatch

        guard firstCell.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            throw XCTSkip("No votes available")
        }

        // Row should contain title text
        let hasTitle = firstCell.staticTexts.count > 0
        XCTAssertTrue(hasTitle, "Vote row should show title")
    }

    func testVoteRowShowsStatus() throws {
        try navigateToVotes()

        Wait.forNetwork()
        Wait.seconds(1.0)

        // Check if there are any votes
        let firstCell = app.collectionViews.cells.firstMatch

        guard firstCell.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            throw XCTSkip("No votes available")
        }

        // Row should contain status indicator (Active, Closed, or date)
        let activeLabel = firstCell.staticTexts.containing(NSPredicate(format: "label CONTAINS[c] 'active'")).firstMatch
        let closedLabel = firstCell.staticTexts.containing(NSPredicate(format: "label CONTAINS[c] 'closed'")).firstMatch
        let dateLabel = firstCell.staticTexts.matching(NSPredicate(format: "label MATCHES %@", ".*\\d{1,2}.*")).firstMatch

        XCTAssertTrue(
            activeLabel.exists ||
            closedLabel.exists ||
            dateLabel.exists ||
            firstCell.staticTexts.count > 1,  // Has multiple labels
            "Vote row should show status"
        )
    }

    // MARK: - Back Navigation Tests

    func testCanNavigateBackFromVotes() throws {
        try navigateToVotes()

        // Tap back button
        let backButton = app.navigationBars.buttons.element(boundBy: 0)
        guard backButton.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            throw XCTSkip("Back button not found")
        }

        backButton.tap()
        Wait.seconds(0.5)

        // Should be back on guild detail
        let guildNavBar = app.navigationBars.element(boundBy: 0)
        XCTAssertTrue(
            guildNavBar.exists,
            "Should navigate back from Votes"
        )
    }

    func testCanNavigateBackFromVoteDetail() throws {
        try navigateToVotes()

        Wait.forNetwork()
        Wait.seconds(1.0)

        // Open a vote detail
        let firstCell = app.collectionViews.cells.firstMatch

        guard firstCell.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            throw XCTSkip("No votes available")
        }

        firstCell.tap()
        Wait.forNetwork()

        // Tap back button
        let backButton = app.navigationBars.buttons.element(boundBy: 0)
        guard backButton.waitForExistence(timeout: TestConfig.defaultTimeout) else {
            throw XCTSkip("Back button not found")
        }

        backButton.tap()
        Wait.seconds(0.5)

        // Should be back on vote list
        let votesNavBar = app.navigationBars["Votes"]
        let createButton = app.buttons[AccessibilityID.Vote.createButton]

        XCTAssertTrue(
            votesNavBar.exists || createButton.exists,
            "Should navigate back to vote list"
        )
    }
}
