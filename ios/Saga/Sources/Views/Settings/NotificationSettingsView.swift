import SwiftUI
@preconcurrency import UserNotifications

/// Notification preferences
struct NotificationSettingsView: View {
    @State private var notificationsEnabled = false
    @State private var isCheckingPermission = true

    // Notification preferences (would be persisted to UserDefaults/API)
    @AppStorage("notify_events") private var notifyEvents = true
    @AppStorage("notify_messages") private var notifyMessages = true
    @AppStorage("notify_trust") private var notifyTrust = true
    @AppStorage("notify_pool_matches") private var notifyPoolMatches = true
    @AppStorage("notify_reminders") private var notifyReminders = true
    @AppStorage("notify_marketing") private var notifyMarketing = false

    var body: some View {
        List {
            // Permission status
            Section {
                HStack {
                    Image(systemName: notificationsEnabled ? "bell.badge.fill" : "bell.slash.fill")
                        .foregroundStyle(notificationsEnabled ? .green : .red)
                        .font(.title2)

                    VStack(alignment: .leading) {
                        Text(notificationsEnabled ? "Notifications Enabled" : "Notifications Disabled")
                            .font(.headline)
                        Text(notificationsEnabled
                            ? "You'll receive notifications for important updates"
                            : "Enable notifications in Settings to stay updated")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                }
                .padding(.vertical, 4)

                if !notificationsEnabled {
                    Button("Open Settings") {
                        openSettings()
                    }
                }
            }

            // Event notifications
            Section {
                Toggle("Event Invitations", isOn: $notifyEvents)
                Toggle("Event Reminders", isOn: $notifyReminders)
            } header: {
                Text("Events")
            } footer: {
                Text("Get notified about event invitations and upcoming events")
            }

            // Social notifications
            Section {
                Toggle("Messages", isOn: $notifyMessages)
                Toggle("Trust Requests", isOn: $notifyTrust)
                Toggle("Pool Matches", isOn: $notifyPoolMatches)
            } header: {
                Text("Social")
            } footer: {
                Text("Stay connected with your community")
            }

            // Marketing
            Section {
                Toggle("Tips & Updates", isOn: $notifyMarketing)
            } header: {
                Text("Marketing")
            } footer: {
                Text("Occasional tips on getting the most out of Saga")
            }
        }
        .navigationTitle("Notifications")
        .navigationBarTitleDisplayMode(.inline)
        .task {
            await checkNotificationPermission()
        }
    }

    private func checkNotificationPermission() async {
        isCheckingPermission = true
        let settings = await UNUserNotificationCenter.current().notificationSettings()
        notificationsEnabled = settings.authorizationStatus == .authorized
        isCheckingPermission = false
    }

    private func openSettings() {
        if let url = URL(string: UIApplication.openSettingsURLString) {
            UIApplication.shared.open(url)
        }
    }
}

#Preview {
    NavigationStack {
        NotificationSettingsView()
    }
}
