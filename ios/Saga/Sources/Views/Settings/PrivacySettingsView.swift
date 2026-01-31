import SwiftUI

/// Privacy settings and controls
struct PrivacySettingsView: View {
    // Privacy preferences (would be synced to API)
    @AppStorage("privacy_profile_visible") private var profileVisible = true
    @AppStorage("privacy_show_online") private var showOnlineStatus = true
    @AppStorage("privacy_show_location") private var showLocation = true
    @AppStorage("privacy_discoverable") private var discoverable = true
    @AppStorage("privacy_allow_messages") private var allowMessages = true
    @AppStorage("privacy_show_resonance") private var showResonance = true

    var body: some View {
        List {
            // Profile visibility
            Section {
                Toggle("Public Profile", isOn: $profileVisible)
                Toggle("Show Online Status", isOn: $showOnlineStatus)
                Toggle("Show Resonance Level", isOn: $showResonance)
            } header: {
                Text("Profile Visibility")
            } footer: {
                Text("Control what others can see about you")
            }

            // Discovery
            Section {
                Toggle("Appear in Discovery", isOn: $discoverable)
                Toggle("Share Location", isOn: $showLocation)
            } header: {
                Text("Discovery")
            } footer: {
                Text("Allow others to find you through discovery features")
            }

            // Communication
            Section {
                Toggle("Allow Messages", isOn: $allowMessages)
            } header: {
                Text("Communication")
            } footer: {
                Text("Control who can contact you")
            }

            // Blocked users
            Section {
                NavigationLink {
                    BlockedUsersView()
                } label: {
                    HStack {
                        Text("Blocked Users")
                        Spacer()
                        Image(systemName: "chevron.right")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                }
            }

            // Data
            Section {
                NavigationLink {
                    DataPrivacyView()
                } label: {
                    HStack {
                        Text("Data & Privacy")
                        Spacer()
                        Image(systemName: "chevron.right")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                }
            } footer: {
                Text("Manage your data and privacy preferences")
            }
        }
        .navigationTitle("Privacy")
        .navigationBarTitleDisplayMode(.inline)
    }
}

// MARK: - Data Privacy View

struct DataPrivacyView: View {
    @State private var showingExportConfirmation = false
    @State private var isExporting = false

    var body: some View {
        List {
            // Data export
            Section {
                Button {
                    showingExportConfirmation = true
                } label: {
                    HStack {
                        Image(systemName: "arrow.down.doc.fill")
                            .foregroundStyle(.blue)
                        Text("Export My Data")
                        Spacer()
                        if isExporting {
                            ProgressView()
                        }
                    }
                }
                .disabled(isExporting)
            } header: {
                Text("Your Data")
            } footer: {
                Text("Download a copy of all your data")
            }

            // Privacy policy
            Section("Legal") {
                Link(destination: URL(string: "https://example.com/privacy")!) {
                    HStack {
                        Text("Privacy Policy")
                        Spacer()
                        Image(systemName: "arrow.up.right")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                }

                Link(destination: URL(string: "https://example.com/terms")!) {
                    HStack {
                        Text("Terms of Service")
                        Spacer()
                        Image(systemName: "arrow.up.right")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                }

                Link(destination: URL(string: "https://example.com/cookies")!) {
                    HStack {
                        Text("Cookie Policy")
                        Spacer()
                        Image(systemName: "arrow.up.right")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                }
            }
        }
        .navigationTitle("Data & Privacy")
        .navigationBarTitleDisplayMode(.inline)
        .alert("Export Data", isPresented: $showingExportConfirmation) {
            Button("Cancel", role: .cancel) { }
            Button("Export") {
                Task { await exportData() }
            }
        } message: {
            Text("We'll prepare your data and send you a download link via email.")
        }
    }

    private func exportData() async {
        isExporting = true
        // TODO: Implement data export
        try? await Task.sleep(for: .seconds(2))
        isExporting = false
    }
}

#Preview {
    NavigationStack {
        PrivacySettingsView()
    }
}
