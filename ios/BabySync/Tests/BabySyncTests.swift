import Testing
@testable import BabySync

@Suite("Model Tests")
struct ModelTests {
    @Test("Timer elapsed calculation")
    func timerElapsed() {
        let resetDate = Date().addingTimeInterval(-3600) // 1 hour ago
        let timer = BabyTimer(
            id: "timer:1",
            babyId: "baby:1",
            activityId: "act:1",
            resetDate: resetDate,
            enabled: true,
            push: false,
            createdOn: Date(),
            updatedOn: Date()
        )

        // Elapsed should be approximately 3600 seconds (1 hour)
        #expect(timer.elapsed >= 3599 && timer.elapsed <= 3601)
    }

    @Test("Timer disabled returns zero elapsed")
    func timerDisabledElapsed() {
        let resetDate = Date().addingTimeInterval(-3600)
        let timer = BabyTimer(
            id: "timer:1",
            babyId: "baby:1",
            activityId: "act:1",
            resetDate: resetDate,
            enabled: false,
            push: false,
            createdOn: Date(),
            updatedOn: Date()
        )

        #expect(timer.elapsed == 0)
    }

    @Test("Activity duration formatting")
    func activityDurationFormatting() {
        let activity = Activity(
            id: "act:1",
            familyId: "family:1",
            name: "Feeding",
            icon: "🍼",
            warn: 10800, // 3 hours
            critical: 14400, // 4 hours
            createdOn: Date(),
            updatedOn: Date()
        )

        #expect(activity.warnFormatted == "3h 0m")
        #expect(activity.criticalFormatted == "4h 0m")
    }

    @Test("User display name")
    func userDisplayName() {
        let userWithName = User(
            id: "user:1",
            email: "test@example.com",
            username: "testuser",
            firstname: "John",
            lastname: "Doe",
            emailVerified: true,
            createdOn: Date(),
            updatedOn: Date()
        )

        #expect(userWithName.displayName == "John Doe")

        let userWithoutName = User(
            id: "user:2",
            email: "test@example.com",
            username: "testuser",
            firstname: nil,
            lastname: nil,
            emailVerified: true,
            createdOn: Date(),
            updatedOn: Date()
        )

        #expect(userWithoutName.displayName == "testuser")
    }
}

@Suite("API Response Tests")
struct APIResponseTests {
    @Test("Decode TokenResponse")
    func decodeTokenResponse() throws {
        let json = """
        {
            "access_token": "abc123",
            "refresh_token": "xyz789",
            "token_type": "Bearer",
            "expires_in": 900
        }
        """

        let decoder = JSONDecoder()
        let response = try decoder.decode(TokenResponse.self, from: Data(json.utf8))

        #expect(response.accessToken == "abc123")
        #expect(response.refreshToken == "xyz789")
        #expect(response.tokenType == "Bearer")
        #expect(response.expiresIn == 900)
    }

    @Test("Decode ProblemDetails")
    func decodeProblemDetails() throws {
        let json = """
        {
            "type": "https://api.babysync.app/errors/validation",
            "title": "Validation Error",
            "status": 422,
            "detail": "Name is required",
            "errors": [
                {"field": "name", "message": "Name is required"}
            ]
        }
        """

        let decoder = JSONDecoder()
        let problem = try decoder.decode(ProblemDetails.self, from: Data(json.utf8))

        #expect(problem.status == 422)
        #expect(problem.title == "Validation Error")
        #expect(problem.errors?.count == 1)
        #expect(problem.errors?.first?.field == "name")
    }
}
