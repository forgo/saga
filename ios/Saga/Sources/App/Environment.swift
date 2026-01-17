import Foundation

/// API environment configuration
enum APIEnvironment {
    case development
    case staging
    case production

    var baseURL: URL {
        switch self {
        case .development:
            return URL(string: "http://localhost:8080/v1")!
        case .staging:
            return URL(string: "https://staging-api.saga.app/v1")!
        case .production:
            return URL(string: "https://api.saga.app/v1")!
        }
    }

    var relyingPartyIdentifier: String {
        switch self {
        case .development:
            return "localhost"
        case .staging:
            return "staging.saga.app"
        case .production:
            return "saga.app"
        }
    }

    var googleClientID: String {
        // TODO: Configure with real Google OAuth client ID
        return "YOUR_GOOGLE_CLIENT_ID.apps.googleusercontent.com"
    }
}

/// Current API environment - change this to switch between environments
let currentEnvironment: APIEnvironment = .development
