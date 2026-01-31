import SwiftUI

/// View showing the user's moderation status
struct ModerationStatusView: View {
    @State private var status: ModerationStatus?
    @State private var isLoading = true
    @State private var error: Error?

    private let apiClient = APIClient.shared

    var body: some View {
        Group {
            if isLoading && status == nil {
                ProgressView("Loading...")
            } else if let status = status {
                statusContent(status)
            } else {
                ContentUnavailableView {
                    Label("Unable to Load", systemImage: "exclamationmark.triangle")
                } description: {
                    Text("Could not load moderation status")
                }
            }
        }
        .navigationTitle("Account Status")
        .refreshable {
            await loadStatus()
        }
        .task {
            await loadStatus()
        }
    }

    @ViewBuilder
    private func statusContent(_ status: ModerationStatus) -> some View {
        List {
            // Status overview
            Section {
                if status.isSuspended {
                    HStack {
                        Image(systemName: "exclamationmark.octagon.fill")
                            .foregroundStyle(.red)
                        VStack(alignment: .leading) {
                            Text("Account Suspended")
                                .font(.headline)
                                .foregroundStyle(.red)
                            if let until = status.suspendedUntil {
                                Text("Until \(until.formatted(date: .long, time: .shortened))")
                                    .font(.caption)
                            }
                            if let reason = status.suspensionReason {
                                Text(reason)
                                    .font(.caption)
                                    .foregroundStyle(.secondary)
                            }
                        }
                    }
                } else if status.hasRestrictions {
                    HStack {
                        Image(systemName: "exclamationmark.triangle.fill")
                            .foregroundStyle(.orange)
                        VStack(alignment: .leading) {
                            Text("Account Restricted")
                                .font(.headline)
                                .foregroundStyle(.orange)
                            Text("Some features are limited")
                                .font(.caption)
                                .foregroundStyle(.secondary)
                        }
                    }
                } else {
                    HStack {
                        Image(systemName: "checkmark.shield.fill")
                            .foregroundStyle(.green)
                        VStack(alignment: .leading) {
                            Text("Account in Good Standing")
                                .font(.headline)
                            Text("No restrictions on your account")
                                .font(.caption)
                                .foregroundStyle(.secondary)
                        }
                    }
                }
            }

            // Warnings
            if status.warningCount > 0 {
                Section("Warnings") {
                    HStack {
                        Image(systemName: "exclamationmark.bubble.fill")
                            .foregroundStyle(.yellow)
                        Text("\(status.warningCount) warning\(status.warningCount == 1 ? "" : "s")")
                        Spacer()
                        if let lastWarning = status.lastWarningOn {
                            Text("Last: \(lastWarning.formatted(date: .abbreviated, time: .omitted))")
                                .font(.caption)
                                .foregroundStyle(.secondary)
                        }
                    }
                }
            }

            // Active restrictions
            if !status.restrictions.isEmpty {
                Section("Active Restrictions") {
                    ForEach(status.restrictions.filter { $0.isActive }, id: \.type) { restriction in
                        HStack {
                            Image(systemName: restriction.type.iconName)
                                .foregroundStyle(.orange)
                                .frame(width: 24)

                            VStack(alignment: .leading) {
                                Text(restriction.type.displayName)
                                    .font(.subheadline)

                                if let reason = restriction.reason {
                                    Text(reason)
                                        .font(.caption)
                                        .foregroundStyle(.secondary)
                                }

                                if let expires = restriction.expiresOn {
                                    Text("Expires: \(expires.formatted(date: .abbreviated, time: .shortened))")
                                        .font(.caption)
                                        .foregroundStyle(.secondary)
                                }
                            }
                        }
                    }
                }
            }

            // Help section
            Section {
                Link(destination: URL(string: "https://example.com/community-guidelines")!) {
                    Label("Community Guidelines", systemImage: "book")
                }

                Link(destination: URL(string: "https://example.com/appeal")!) {
                    Label("Appeal a Decision", systemImage: "envelope")
                }
            } header: {
                Text("Help")
            } footer: {
                Text("If you believe a decision was made in error, you can submit an appeal.")
            }
        }
    }

    private func loadStatus() async {
        isLoading = true
        do {
            status = try await apiClient.getModerationStatus()
        } catch {
            self.error = error
        }
        isLoading = false
    }
}

#Preview {
    NavigationStack {
        ModerationStatusView()
    }
}
