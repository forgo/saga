import Foundation

/// Global application state coordinator
@Observable
@MainActor
final class AppState {
    let authService: AuthService
    let passkeyService: PasskeyService
    let guildService: GuildService
    let eventService: EventService
    let profileService: ProfileService
    let discoveryService: DiscoveryService

    init() {
        self.authService = AuthService.shared
        self.passkeyService = PasskeyService.shared
        self.guildService = GuildService.shared
        self.eventService = EventService.shared
        self.profileService = ProfileService.shared
        self.discoveryService = DiscoveryService.shared

        #if DEBUG
        // In demo mode, auto-login with demo user for testing
        if isDemoMode {
            Task {
                try? await authService.loginWithDemoUser()
            }
        }
        #endif
    }

    /// Whether the user is currently authenticated
    var isAuthenticated: Bool {
        authService.isAuthenticated
    }
}
