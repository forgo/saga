import SwiftUI

struct SettingsView: View {
    @Environment(AuthService.self) private var authService

    @State private var showLogoutConfirm = false
    @State private var showError = false
    @State private var errorMessage = ""

    var body: some View {
        NavigationStack {
            List {
                // Account section
                Section {
                    if let user = authService.currentUser {
                        HStack {
                            Image(systemName: "person.circle.fill")
                                .font(.largeTitle)
                                .foregroundStyle(.blue)

                            VStack(alignment: .leading) {
                                Text(user.displayName)
                                    .font(.headline)
                                Text(user.email)
                                    .font(.caption)
                                    .foregroundStyle(.secondary)
                            }
                        }
                        .padding(.vertical, 4)
                    }
                } header: {
                    Text("Account")
                }

                // Passkeys section
                Section {
                    NavigationLink {
                        PasskeysListView()
                    } label: {
                        Label("Manage Passkeys", systemImage: "person.badge.key")
                    }
                } header: {
                    Text("Security")
                } footer: {
                    Text("Sign in faster with Face ID or Touch ID")
                }

                // Notifications section
                Section {
                    NavigationLink {
                        NotificationSettingsView()
                    } label: {
                        Label("Notifications", systemImage: "bell.badge")
                    }
                } header: {
                    Text("Preferences")
                }

                // About section
                Section {
                    LabeledContent("Version", value: "1.0.0")
                    LabeledContent("Build", value: "1")

                    Link(destination: URL(string: "https://babysync.app/privacy")!) {
                        Label("Privacy Policy", systemImage: "hand.raised")
                    }

                    Link(destination: URL(string: "https://babysync.app/terms")!) {
                        Label("Terms of Service", systemImage: "doc.text")
                    }
                } header: {
                    Text("About")
                }

                // Sign out section
                Section {
                    Button(role: .destructive) {
                        showLogoutConfirm = true
                    } label: {
                        HStack {
                            Spacer()
                            Text("Sign Out")
                            Spacer()
                        }
                    }
                }
            }
            .navigationTitle("Settings")
            .confirmationDialog("Sign Out", isPresented: $showLogoutConfirm) {
                Button("Sign Out", role: .destructive) {
                    Task { await logout() }
                }
            } message: {
                Text("Are you sure you want to sign out?")
            }
            .alert("Error", isPresented: $showError) {
                Button("OK") { }
            } message: {
                Text(errorMessage)
            }
        }
    }

    private func logout() async {
        do {
            try await authService.logout()
        } catch {
            errorMessage = error.localizedDescription
            showError = true
        }
    }
}

struct PasskeysListView: View {
    @State private var passkeys: [Passkey] = []

    var body: some View {
        List {
            if passkeys.isEmpty {
                ContentUnavailableView {
                    Label("No Passkeys", systemImage: "person.badge.key")
                } description: {
                    Text("Add a passkey to sign in faster with Face ID or Touch ID")
                }
            } else {
                ForEach(passkeys) { passkey in
                    HStack {
                        Image(systemName: "person.badge.key.fill")
                            .foregroundStyle(.blue)

                        VStack(alignment: .leading) {
                            Text(passkey.name)
                                .font(.headline)
                            if let lastUsed = passkey.lastUsedOn {
                                Text("Last used \(lastUsed.formatted(date: .abbreviated, time: .shortened))")
                                    .font(.caption)
                                    .foregroundStyle(.secondary)
                            }
                        }
                    }
                }
                .onDelete { _ in
                    // Delete passkey
                }
            }

            Section {
                Button {
                    // Add passkey
                } label: {
                    Label("Add Passkey", systemImage: "plus.circle.fill")
                }
            }
        }
        .navigationTitle("Passkeys")
    }
}

struct NotificationSettingsView: View {
    @AppStorage("notificationsEnabled") private var notificationsEnabled = true
    @AppStorage("warningNotifications") private var warningNotifications = true
    @AppStorage("criticalNotifications") private var criticalNotifications = true

    var body: some View {
        List {
            Section {
                Toggle("Enable Notifications", isOn: $notificationsEnabled)
            }

            Section {
                Toggle("Warning Alerts", isOn: $warningNotifications)
                    .disabled(!notificationsEnabled)
                Toggle("Critical Alerts", isOn: $criticalNotifications)
                    .disabled(!notificationsEnabled)
            } header: {
                Text("Timer Alerts")
            } footer: {
                Text("Get notified when timers reach their thresholds")
            }
        }
        .navigationTitle("Notifications")
    }
}

#Preview {
    SettingsView()
        .environment(AuthService.shared)
}
