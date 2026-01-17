import Foundation

/// Global application state coordinator
@Observable
final class AppState: @unchecked Sendable {
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
    }

    /// Whether the user is currently authenticated
    var isAuthenticated: Bool {
        authService.isAuthenticated
    }
}
