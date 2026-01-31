import SwiftUI

/// Main settings view
struct SettingsView: View {
    @Environment(AuthService.self) private var authService

    var body: some View {
        List {
            // Account section
            Section {
                NavigationLink {
                    AccountSettingsView()
                } label: {
                    SettingsRow(
                        icon: "person.circle.fill",
                        iconColor: .blue,
                        title: "Account",
                        subtitle: authService.currentUser?.email
                    )
                }
            }

            // Preferences section
            Section("Preferences") {
                NavigationLink {
                    NotificationSettingsView()
                } label: {
                    SettingsRow(
                        icon: "bell.fill",
                        iconColor: .red,
                        title: "Notifications"
                    )
                }

                NavigationLink {
                    PrivacySettingsView()
                } label: {
                    SettingsRow(
                        icon: "hand.raised.fill",
                        iconColor: .blue,
                        title: "Privacy"
                    )
                }

                NavigationLink {
                    AppearanceSettingsView()
                } label: {
                    SettingsRow(
                        icon: "paintbrush.fill",
                        iconColor: .purple,
                        title: "Appearance"
                    )
                }
            }

            // Support section
            Section("Support") {
                NavigationLink {
                    HelpView()
                } label: {
                    SettingsRow(
                        icon: "questionmark.circle.fill",
                        iconColor: .green,
                        title: "Help & Support"
                    )
                }

                NavigationLink {
                    AboutView()
                } label: {
                    SettingsRow(
                        icon: "info.circle.fill",
                        iconColor: .gray,
                        title: "About"
                    )
                }
            }

            // Debug section (only in debug builds)
            #if DEBUG
            Section("Developer") {
                NavigationLink {
                    DebugSettingsView()
                } label: {
                    SettingsRow(
                        icon: "hammer.fill",
                        iconColor: .orange,
                        title: "Debug"
                    )
                }
            }
            #endif
        }
        .navigationTitle("Settings")
    }
}

// MARK: - Settings Row

struct SettingsRow: View {
    let icon: String
    let iconColor: Color
    let title: String
    var subtitle: String? = nil

    var body: some View {
        HStack(spacing: 12) {
            Image(systemName: icon)
                .font(.title3)
                .foregroundStyle(iconColor)
                .frame(width: 28)

            VStack(alignment: .leading, spacing: 2) {
                Text(title)
                if let subtitle = subtitle {
                    Text(subtitle)
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }
            }
        }
        .padding(.vertical, 2)
    }
}

#Preview {
    NavigationStack {
        SettingsView()
            .environment(AuthService.shared)
    }
}
