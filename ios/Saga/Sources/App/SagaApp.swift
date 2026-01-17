import SwiftUI

@main
struct SagaApp: App {
    @State private var appState = AppState()

    var body: some Scene {
        WindowGroup {
            ContentView()
                .environment(appState)
                .environment(appState.authService)
                .environment(appState.passkeyService)
                .environment(appState.guildService)
                .environment(appState.eventService)
                .environment(appState.profileService)
                .environment(appState.discoveryService)
        }
    }
}
