import SwiftUI

/// Detailed view of a matching pool
struct PoolDetailView: View {
    let poolId: String

    @State private var details: PoolWithDetails?
    @State private var isLoading = true
    @State private var error: Error?
    @State private var isJoining = false
    @State private var showPreferencesSheet = false

    private let apiClient = APIClient.shared

    var pool: Pool? { details?.pool }

    var body: some View {
        Group {
            if isLoading && details == nil {
                ProgressView("Loading pool...")
            } else if let pool = pool {
                poolContent(pool)
            } else {
                ContentUnavailableView {
                    Label("Pool Not Found", systemImage: "exclamationmark.triangle")
                } description: {
                    Text("Unable to load this pool")
                }
            }
        }
        .navigationTitle(pool?.name ?? "Pool")
        .navigationBarTitleDisplayMode(.inline)
        .refreshable {
            await loadDetails()
        }
        .sheet(isPresented: $showPreferencesSheet) {
            if let membership = details?.myMembership {
                NavigationStack {
                    PoolPreferencesSheet(
                        poolId: poolId,
                        currentPreferences: membership.preferences
                    )
                }
            }
        }
        .task {
            await loadDetails()
        }
    }

    @ViewBuilder
    private func poolContent(_ pool: Pool) -> some View {
        List {
            // Header Section
            Section {
                VStack(alignment: .leading, spacing: 12) {
                    HStack {
                        PoolStatusBadge(isActive: pool.isActive)
                        Spacer()
                        Label("\(pool.memberCount) members", systemImage: "person.2")
                            .font(.subheadline)
                    }

                    if let description = pool.description {
                        Text(description)
                            .font(.body)
                    }

                    // Frequency
                    HStack {
                        Image(systemName: "clock")
                            .foregroundStyle(.blue)
                        Text("Matches \(pool.matchingFrequency.displayName.lowercased())")
                    }
                    .font(.subheadline)

                    // Next matching
                    if let nextMatch = pool.nextMatchDate {
                        HStack {
                            Image(systemName: "calendar")
                                .foregroundStyle(.orange)
                            Text("Next: \(nextMatch.formatted(date: .abbreviated, time: .omitted))")
                        }
                        .font(.subheadline)
                    }
                }
                .padding(.vertical, 4)
            }

            // Membership Section
            if let membership = details?.myMembership {
                Section("Your Membership") {
                    HStack {
                        Image(systemName: "checkmark.circle.fill")
                            .foregroundStyle(.green)
                        Text("Member since \(membership.joinedOn.formatted(date: .abbreviated, time: .omitted))")
                    }

                    Button {
                        showPreferencesSheet = true
                    } label: {
                        Label("Edit Preferences", systemImage: "slider.horizontal.3")
                    }

                    Button(role: .destructive) {
                        Task { await leavePool() }
                    } label: {
                        Label("Leave Pool", systemImage: "xmark.circle")
                    }
                }
            } else if pool.isActive {
                Section {
                    Button {
                        Task { await joinPool() }
                    } label: {
                        HStack {
                            Spacer()
                            if isJoining {
                                ProgressView()
                            } else {
                                Label("Join Pool", systemImage: "plus.circle")
                            }
                            Spacer()
                        }
                    }
                    .disabled(isJoining)
                }
            }

            // Recent Matches Section
            if let matches = details?.recentMatches, !matches.isEmpty {
                Section("Recent Matches") {
                    ForEach(matches, id: \.match.id) { display in
                        MatchRow(display: display)
                    }
                }
            }
        }
    }

    private func loadDetails() async {
        isLoading = true
        do {
            details = try await apiClient.getPool(poolId: poolId)
        } catch {
            self.error = error
        }
        isLoading = false
    }

    private func joinPool() async {
        isJoining = true
        do {
            _ = try await apiClient.joinPool(poolId: poolId)
            await loadDetails()
        } catch {
            self.error = error
        }
        isJoining = false
    }

    private func leavePool() async {
        do {
            try await apiClient.leavePool(poolId: poolId)
            await loadDetails()
        } catch {
            self.error = error
        }
    }
}

// MARK: - Match Row

struct MatchRow: View {
    let display: PoolMatchDisplay

    var participantNames: String {
        display.participants.map { $0.displayName }.joined(separator: ", ")
    }

    var body: some View {
        HStack {
            ZStack {
                Circle()
                    .fill(.blue.gradient)
                    .frame(width: 40, height: 40)
                Image(systemName: "person.2")
                    .font(.subheadline.bold())
                    .foregroundStyle(.white)
            }

            VStack(alignment: .leading) {
                Text(participantNames)
                    .font(.subheadline)
                    .lineLimit(1)

                if let meetingTime = display.match.meetingScheduled {
                    Text("Meet: \(meetingTime.formatted(date: .abbreviated, time: .shortened))")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                } else {
                    Text(display.match.matchDate.formatted(date: .abbreviated, time: .omitted))
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }
            }

            Spacer()

            MatchStatusBadge(status: display.match.status)
        }
    }
}

// MARK: - Match Status Badge

struct MatchStatusBadge: View {
    let status: MatchStatus

    var color: Color {
        switch status {
        case .pending: return .orange
        case .accepted: return .green
        case .scheduled: return .blue
        case .completed: return .purple
        case .skipped: return .gray
        }
    }

    var body: some View {
        Text(status.displayName)
            .font(.caption2.bold())
            .padding(.horizontal, 6)
            .padding(.vertical, 3)
            .background(color.opacity(0.2))
            .foregroundStyle(color)
            .clipShape(Capsule())
    }
}

#Preview {
    NavigationStack {
        PoolDetailView(poolId: "test")
    }
}
