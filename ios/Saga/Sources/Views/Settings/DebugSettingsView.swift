import SwiftUI

/// Debug settings for development
struct DebugSettingsView: View {
    @Environment(AuthService.self) private var authService
    @Environment(GuildService.self) private var guildService

    private var apiBaseURL: String { currentEnvironment.baseURL.absoluteString }
    @State private var showingClearConfirmation = false

    var body: some View {
        List {
            // Environment
            Section("Environment") {
                HStack {
                    Text("API URL")
                    Spacer()
                    Text(apiBaseURL)
                        .font(.caption.monospaced())
                        .foregroundStyle(.secondary)
                }

                HStack {
                    Text("Build")
                    Spacer()
                    #if DEBUG
                    Text("Debug")
                        .foregroundStyle(.orange)
                    #else
                    Text("Release")
                        .foregroundStyle(.green)
                    #endif
                }
            }

            // Auth state
            Section("Authentication") {
                HStack {
                    Text("Authenticated")
                    Spacer()
                    Image(systemName: authService.isAuthenticated ? "checkmark.circle.fill" : "xmark.circle.fill")
                        .foregroundStyle(authService.isAuthenticated ? .green : .red)
                }

                if let user = authService.currentUser {
                    HStack {
                        Text("User ID")
                        Spacer()
                        Text(user.id)
                            .font(.caption.monospaced())
                            .foregroundStyle(.secondary)
                            .lineLimit(1)
                    }
                }

                HStack {
                    Text("Identities")
                    Spacer()
                    Text("\(authService.identities.count)")
                        .foregroundStyle(.secondary)
                }

                HStack {
                    Text("Passkeys")
                    Spacer()
                    Text("\(authService.passkeys.count)")
                        .foregroundStyle(.secondary)
                }
            }

            // Guild state
            Section("Guilds") {
                HStack {
                    Text("Loaded Guilds")
                    Spacer()
                    Text("\(guildService.guilds.count)")
                        .foregroundStyle(.secondary)
                }

                HStack {
                    Text("Current Guild")
                    Spacer()
                    Text(guildService.currentGuild?.guild.name ?? "None")
                        .foregroundStyle(.secondary)
                }

                HStack {
                    Text("SSE Connected")
                    Spacer()
                    Image(systemName: guildService.isConnected ? "checkmark.circle.fill" : "xmark.circle.fill")
                        .foregroundStyle(guildService.isConnected ? .green : .red)
                }
            }

            // Actions
            Section("Actions") {
                Button("Clear All Caches") {
                    showingClearConfirmation = true
                }

                Button("Force Refresh Tokens") {
                    Task {
                        // TODO: Force token refresh
                    }
                }

                Button("Simulate Network Error") {
                    // TODO: Simulate network error
                }
            }

            // Logs
            Section("Logs") {
                NavigationLink("View Logs") {
                    LogsView()
                }
            }
        }
        .navigationTitle("Debug")
        .navigationBarTitleDisplayMode(.inline)
        .alert("Clear Caches", isPresented: $showingClearConfirmation) {
            Button("Cancel", role: .cancel) { }
            Button("Clear", role: .destructive) {
                clearCaches()
            }
        } message: {
            Text("This will clear all cached data. You may need to sign in again.")
        }
    }

    private func clearCaches() {
        // Clear UserDefaults (except essential keys)
        if let bundleID = Bundle.main.bundleIdentifier {
            UserDefaults.standard.removePersistentDomain(forName: bundleID)
        }

        // Clear URL cache
        URLCache.shared.removeAllCachedResponses()
    }
}

// MARK: - Logs View

struct LogsView: View {
    @State private var logs: [String] = []

    var body: some View {
        List {
            if logs.isEmpty {
                ContentUnavailableView(
                    "No Logs",
                    systemImage: "doc.text",
                    description: Text("Logs will appear here")
                )
            } else {
                ForEach(logs, id: \.self) { log in
                    Text(log)
                        .font(.caption.monospaced())
                }
            }
        }
        .navigationTitle("Logs")
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .primaryAction) {
                Button("Clear") {
                    logs.removeAll()
                }
            }
        }
    }
}

#Preview {
    NavigationStack {
        DebugSettingsView()
            .environment(AuthService.shared)
            .environment(GuildService.shared)
    }
}
