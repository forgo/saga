import Foundation

/// API environment configuration
enum APIEnvironment {
    case development
    case staging
    case production
    case testing

    var baseURL: URL {
        switch self {
        case .development, .testing:
            return URL(string: "http://localhost:8080/v1")!
        case .staging:
            return URL(string: "https://staging-api.saga.app/v1")!
        case .production:
            return URL(string: "https://api.saga.app/v1")!
        }
    }

    var relyingPartyIdentifier: String {
        switch self {
        case .development, .testing:
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

    var isTesting: Bool {
        self == .testing
    }
}

// MARK: - Environment Detection

/// Current API environment - automatically detected from launch arguments
let currentEnvironment: APIEnvironment = {
    #if DEBUG
    if ProcessInfo.processInfo.arguments.contains("--uitesting") {
        return .testing
    }
    #endif
    return .development
}()

#if DEBUG
/// Check if running in UI test mode
var isUITesting: Bool {
    ProcessInfo.processInfo.arguments.contains("--uitesting")
}

/// Check if demo mode is enabled (auto-login with demo user)
var isDemoMode: Bool {
    ProcessInfo.processInfo.arguments.contains("--demo")
}
#endif
