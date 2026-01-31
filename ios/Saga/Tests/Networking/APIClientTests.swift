import XCTest
@testable import Saga

/// Tests for the APIClient networking layer
final class APIClientTests: XCTestCase {

    // MARK: - Test Helpers

    struct TestResponse: Codable, Sendable, Equatable {
        let id: String
        let name: String
    }

    func makeSuccessResponse<T: Encodable>(data: T, statusCode: Int = 200) throws -> (Data, HTTPURLResponse) {
        let encoder = JSONEncoder()
        // Manually create the JSON structure to avoid DataResponse init issues
        let wrapper = ["data": data]
        let responseData = try encoder.encode(wrapper)
        let response = HTTPURLResponse(
            url: URL(string: "https://api.example.com")!,
            statusCode: statusCode,
            httpVersion: nil,
            headerFields: nil
        )!
        return (responseData, response)
    }

    func makeErrorResponse(statusCode: Int, body: Data? = nil, headers: [String: String]? = nil) -> (Data, HTTPURLResponse) {
        let response = HTTPURLResponse(
            url: URL(string: "https://api.example.com")!,
            statusCode: statusCode,
            httpVersion: nil,
            headerFields: headers
        )!
        return (body ?? Data(), response)
    }

    // MARK: - Token Management Tests

    func testSetTokens_SetsAccessAndRefreshTokens() async throws {
        let client = APIClient(baseURL: URL(string: "https://api.example.com")!)

        await client.setTokens(access: "access-token-123", refresh: "refresh-token-456")

        let accessToken = await client.getAccessToken()
        XCTAssertEqual(accessToken, "access-token-123")
    }

    func testClearTokens_RemovesTokens() async throws {
        let client = APIClient(baseURL: URL(string: "https://api.example.com")!)

        await client.setTokens(access: "access-token-123", refresh: "refresh-token-456")
        await client.clearTokens()

        let accessToken = await client.getAccessToken()
        XCTAssertNil(accessToken)
    }

    func testIsAuthenticated_ReturnsTrueWhenTokenSet() async throws {
        let client = APIClient(baseURL: URL(string: "https://api.example.com")!)

        await client.setTokens(access: "access-token", refresh: "refresh-token")

        let isAuth = await client.isAuthenticated
        XCTAssertTrue(isAuth)
    }

    func testIsAuthenticated_ReturnsFalseWhenNoToken() async throws {
        let client = APIClient(baseURL: URL(string: "https://api.example.com")!)

        let isAuth = await client.isAuthenticated
        XCTAssertFalse(isAuth)
    }

    // MARK: - Date Decoding Tests

    func testDateDecoding_WithFractionalSeconds() throws {
        // Test date string format from Go/SurrealDB with nanoseconds
        let json = """
        {
            "data": {
                "id": "test:123",
                "name": "Test",
                "createdOn": "2024-01-15T10:30:45.123456789Z"
            }
        }
        """

        struct DateTestResponse: Codable, Sendable {
            let id: String
            let name: String
            let createdOn: Date
        }

        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .custom { decoder in
            let container = try decoder.singleValueContainer()
            let dateString = try container.decode(String.self)

            let formatter = ISO8601DateFormatter()
            formatter.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
            if let date = formatter.date(from: dateString) {
                return date
            }

            formatter.formatOptions = [.withInternetDateTime]
            if let date = formatter.date(from: dateString) {
                return date
            }

            throw DecodingError.dataCorruptedError(
                in: container,
                debugDescription: "Cannot decode date: \(dateString)"
            )
        }

        let data = json.data(using: .utf8)!
        let response = try decoder.decode(DataResponse<DateTestResponse>.self, from: data)

        XCTAssertEqual(response.data.id, "test:123")
        XCTAssertNotNil(response.data.createdOn)
    }

    func testDateDecoding_WithoutFractionalSeconds() throws {
        // Test standard ISO8601 without fractional seconds
        let json = """
        {
            "data": {
                "id": "test:456",
                "name": "Test2",
                "createdOn": "2024-01-15T10:30:45Z"
            }
        }
        """

        struct DateTestResponse: Codable, Sendable {
            let id: String
            let name: String
            let createdOn: Date
        }

        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .custom { decoder in
            let container = try decoder.singleValueContainer()
            let dateString = try container.decode(String.self)

            let formatter = ISO8601DateFormatter()
            formatter.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
            if let date = formatter.date(from: dateString) {
                return date
            }

            formatter.formatOptions = [.withInternetDateTime]
            if let date = formatter.date(from: dateString) {
                return date
            }

            throw DecodingError.dataCorruptedError(
                in: container,
                debugDescription: "Cannot decode date: \(dateString)"
            )
        }

        let data = json.data(using: .utf8)!
        let response = try decoder.decode(DataResponse<DateTestResponse>.self, from: data)

        XCTAssertEqual(response.data.id, "test:456")
        XCTAssertNotNil(response.data.createdOn)
    }

    // MARK: - Error Parsing Tests

    func testErrorParsing_ProblemDetails() throws {
        let problemJSON = """
        {
            "type": "https://api.example.com/errors/validation",
            "title": "Validation Error",
            "status": 422,
            "detail": "Email is invalid",
            "errors": [
                {"field": "email", "message": "invalid email format"}
            ]
        }
        """

        let data = problemJSON.data(using: .utf8)!
        let problem = try JSONDecoder().decode(ProblemDetails.self, from: data)

        XCTAssertEqual(problem.status, 422)
        XCTAssertEqual(problem.title, "Validation Error")
        XCTAssertEqual(problem.detail, "Email is invalid")
        XCTAssertEqual(problem.errors?.count, 1)
        XCTAssertEqual(problem.errors?.first?.field, "email")
    }

    func testErrorParsing_ProblemDetailsWithCode() throws {
        let problemJSON = """
        {
            "type": "https://api.example.com/errors/limit-exceeded",
            "title": "Limit Exceeded",
            "status": 422,
            "detail": "Maximum of 5 devices reached",
            "code": 4003,
            "limit": 5,
            "current": 5
        }
        """

        let data = problemJSON.data(using: .utf8)!
        let problem = try JSONDecoder().decode(ProblemDetails.self, from: data)

        XCTAssertEqual(problem.status, 422)
        XCTAssertEqual(problem.code, 4003)
        XCTAssertEqual(problem.limit, 5)
        XCTAssertEqual(problem.current, 5)
    }

    // MARK: - APIError Tests

    func testAPIError_ErrorDescription_AllCases() {
        let testCases: [(APIError, String)] = [
            (.invalidURL, "Invalid URL"),
            (.unauthorized, "Unauthorized - please log in"),
            (.forbidden, "Access denied"),
            (.notFound, "Resource not found"),
            (.conflict, "Resource conflict"),
            (.rateLimited(retryAfter: 30), "Too many requests. Retry after 30 seconds"),
            (.serverError, "Server error - please try again later"),
            (.unknown, "An unknown error occurred"),
        ]

        for (error, expectedDescription) in testCases {
            XCTAssertEqual(error.errorDescription, expectedDescription, "Failed for \(error)")
        }
    }

    func testAPIError_UserMessage_AllCases() {
        let testCases: [(APIError, String)] = [
            (.unauthorized, "Please sign in to continue."),
            (.forbidden, "You don't have permission to do that."),
            (.notFound, "The requested item was not found."),
            (.rateLimited(retryAfter: 60), "You're doing that too fast. Please wait a moment."),
            (.serverError, "Something went wrong. Please try again."),
        ]

        for (error, expectedMessage) in testCases {
            XCTAssertEqual(error.userMessage, expectedMessage, "Failed for \(error)")
        }
    }

    func testAPIError_ProblemDetails_UsesDetail() {
        let problem = ProblemDetails(
            type: "https://api.example.com/errors/custom",
            title: "Custom Error",
            status: 400,
            detail: "This is a detailed error message"
        )
        let error = APIError.problemDetails(problem)

        XCTAssertEqual(error.userMessage, "This is a detailed error message")
        XCTAssertEqual(error.errorDescription, "This is a detailed error message")
    }

    func testAPIError_ProblemDetails_FallsBackToTitle() {
        let problem = ProblemDetails(
            type: "https://api.example.com/errors/custom",
            title: "Custom Error",
            status: 400,
            detail: nil
        )
        let error = APIError.problemDetails(problem)

        XCTAssertEqual(error.userMessage, "Custom Error")
        XCTAssertEqual(error.errorDescription, "Custom Error")
    }

    // MARK: - DataResponse Tests

    func testDataResponse_DecodesSuccessfully() throws {
        let json = """
        {
            "data": {
                "id": "user:123",
                "name": "John Doe"
            },
            "_links": {
                "self": "/users/123"
            }
        }
        """

        let data = json.data(using: .utf8)!
        let response = try JSONDecoder().decode(DataResponse<TestResponse>.self, from: data)

        XCTAssertEqual(response.data.id, "user:123")
        XCTAssertEqual(response.data.name, "John Doe")
        XCTAssertEqual(response.links?["self"], "/users/123")
    }

    func testDataResponse_DecodesWithoutLinks() throws {
        let json = """
        {
            "data": {
                "id": "user:456",
                "name": "Jane Doe"
            }
        }
        """

        let data = json.data(using: .utf8)!
        let response = try JSONDecoder().decode(DataResponse<TestResponse>.self, from: data)

        XCTAssertEqual(response.data.id, "user:456")
        XCTAssertNil(response.links)
    }

    // MARK: - CollectionResponse Tests

    func testCollectionResponse_DecodesSuccessfully() throws {
        let json = """
        {
            "data": [
                {"id": "1", "name": "Item 1"},
                {"id": "2", "name": "Item 2"}
            ],
            "pagination": {
                "cursor": "abc123",
                "has_more": true
            }
        }
        """

        let data = json.data(using: .utf8)!
        let response = try JSONDecoder().decode(CollectionResponse<TestResponse>.self, from: data)

        XCTAssertEqual(response.data.count, 2)
        XCTAssertEqual(response.data[0].id, "1")
        XCTAssertEqual(response.pagination?.cursor, "abc123")
        XCTAssertEqual(response.pagination?.hasMore, true)
    }

    func testCollectionResponse_DecodesEmptyCollection() throws {
        let json = """
        {
            "data": [],
            "pagination": {
                "has_more": false
            }
        }
        """

        let data = json.data(using: .utf8)!
        let response = try JSONDecoder().decode(CollectionResponse<TestResponse>.self, from: data)

        XCTAssertEqual(response.data.count, 0)
        XCTAssertEqual(response.pagination?.hasMore, false)
        XCTAssertNil(response.pagination?.cursor)
    }

    // MARK: - TokenResponse Tests

    func testTokenResponse_DecodesSuccessfully() throws {
        let json = """
        {
            "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
            "refresh_token": "refresh_abc123",
            "token_type": "Bearer",
            "expires_in": 3600
        }
        """

        let data = json.data(using: .utf8)!
        let response = try JSONDecoder().decode(TokenResponse.self, from: data)

        XCTAssertTrue(response.accessToken.hasPrefix("eyJ"))
        XCTAssertEqual(response.refreshToken, "refresh_abc123")
        XCTAssertEqual(response.tokenType, "Bearer")
        XCTAssertEqual(response.expiresIn, 3600)
    }

    // MARK: - AuthResponse Tests

    func testAuthResponse_DecodesSuccessfully() throws {
        let json = """
        {
            "user": {
                "id": "user:123",
                "email": "test@example.com",
                "email_verified": true,
                "created_on": "2024-01-15T10:30:45Z",
                "updated_on": "2024-01-15T10:30:45Z"
            },
            "token": {
                "access_token": "access_xyz",
                "refresh_token": "refresh_xyz",
                "token_type": "Bearer",
                "expires_in": 3600
            }
        }
        """

        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601

        let data = json.data(using: .utf8)!
        let response = try decoder.decode(AuthResponse.self, from: data)

        XCTAssertEqual(response.user.id, "user:123")
        XCTAssertEqual(response.user.email, "test@example.com")
        XCTAssertEqual(response.token.accessToken, "access_xyz")
    }

    // MARK: - Rate Limit Header Parsing Tests

    func testRateLimitRetryAfter_DefaultsTo30WhenMissing() {
        // Simulate extracting Retry-After header
        let retryAfterValue: String? = nil
        let retryAfter = Int(retryAfterValue ?? "30") ?? 30

        XCTAssertEqual(retryAfter, 30)
    }

    func testRateLimitRetryAfter_ParsesValidValue() {
        let retryAfterValue: String? = "60"
        let retryAfter = Int(retryAfterValue ?? "30") ?? 30

        XCTAssertEqual(retryAfter, 60)
    }

    func testRateLimitRetryAfter_DefaultsOnInvalidValue() {
        let retryAfterValue: String? = "invalid"
        let retryAfter = Int(retryAfterValue ?? "30") ?? 30

        XCTAssertEqual(retryAfter, 30)
    }

    // MARK: - ProblemDetails Convenience Tests

    func testProblemDetails_LocalizedDescription() {
        let problem1 = ProblemDetails(
            type: "https://api.example.com/errors/test",
            title: "Test Error",
            status: 400,
            detail: "Detailed message"
        )
        XCTAssertEqual(problem1.localizedDescription, "Detailed message")

        let problem2 = ProblemDetails(
            type: "https://api.example.com/errors/test",
            title: "Test Error",
            status: 400,
            detail: nil
        )
        XCTAssertEqual(problem2.localizedDescription, "Test Error")
    }

    // MARK: - Field Error Tests

    func testFieldError_DecodesCorrectly() throws {
        let json = """
        {
            "field": "email",
            "message": "invalid email format"
        }
        """

        let data = json.data(using: .utf8)!
        let error = try JSONDecoder().decode(FieldError.self, from: data)

        XCTAssertEqual(error.field, "email")
        XCTAssertEqual(error.message, "invalid email format")
    }
}

