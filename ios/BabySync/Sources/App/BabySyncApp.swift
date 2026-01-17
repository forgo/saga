import SwiftUI

@main
struct BabySyncApp: App {
    @State private var authService = AuthService.shared
    @State private var familyService = FamilyService.shared

    var body: some Scene {
        WindowGroup {
            ContentView()
                .environment(authService)
                .environment(familyService)
        }
    }
}

struct ContentView: View {
    @Environment(AuthService.self) private var authService

    var body: some View {
        Group {
            if authService.isAuthenticated {
                MainTabView()
            } else {
                AuthView()
            }
        }
        .animation(.easeInOut, value: authService.isAuthenticated)
    }
}

struct MainTabView: View {
    var body: some View {
        TabView {
            FamilyListView()
                .tabItem {
                    Label("Families", systemImage: "house.fill")
                }

            SettingsView()
                .tabItem {
                    Label("Settings", systemImage: "gear")
                }
        }
    }
}
