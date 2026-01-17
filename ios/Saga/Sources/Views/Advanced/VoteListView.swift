import SwiftUI

/// View listing votes for a guild
struct VoteListView: View {
    let guildId: String

    @State private var votes: [Vote] = []
    @State private var isLoading = true
    @State private var error: Error?
    @State private var showCreateSheet = false
    @State private var filterStatus: VoteStatus?

    private let apiClient = APIClient.shared

    var filteredVotes: [Vote] {
        if let status = filterStatus {
            return votes.filter { $0.status == status }
        }
        return votes
    }

    var body: some View {
        Group {
            if isLoading && votes.isEmpty {
                ProgressView("Loading votes...")
            } else if votes.isEmpty {
                ContentUnavailableView {
                    Label("No Votes", systemImage: "chart.bar.xaxis")
                } description: {
                    Text("Create a vote to gather opinions from members")
                } actions: {
                    Button("Create Vote") {
                        showCreateSheet = true
                    }
                    .buttonStyle(.borderedProminent)
                }
            } else {
                List {
                    // Filter Picker
                    Section {
                        Picker("Filter", selection: $filterStatus) {
                            Text("All").tag(nil as VoteStatus?)
                            ForEach(VoteStatus.allCases, id: \.self) { status in
                                Text(status.displayName).tag(status as VoteStatus?)
                            }
                        }
                        .pickerStyle(.segmented)
                    }
                    .listRowBackground(Color.clear)

                    // Votes
                    ForEach(filteredVotes) { vote in
                        NavigationLink {
                            VoteDetailView(voteId: vote.id)
                        } label: {
                            VoteRow(vote: vote)
                        }
                    }
                }
            }
        }
        .navigationTitle("Votes")
        .toolbar {
            ToolbarItem(placement: .primaryAction) {
                Button {
                    showCreateSheet = true
                } label: {
                    Image(systemName: "plus")
                }
            }
        }
        .refreshable {
            await loadVotes()
        }
        .sheet(isPresented: $showCreateSheet) {
            NavigationStack {
                CreateVoteSheet(guildId: guildId)
            }
        }
        .task {
            await loadVotes()
        }
    }

    private func loadVotes() async {
        isLoading = true
        do {
            votes = try await apiClient.getVotes(guildId: guildId)
        } catch {
            self.error = error
        }
        isLoading = false
    }
}

// MARK: - Vote Row

struct VoteRow: View {
    let vote: Vote

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Text(vote.title)
                    .font(.headline)

                Spacer()

                VoteStatusBadge(status: vote.status)
            }

            if let description = vote.description {
                Text(description)
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
                    .lineLimit(2)
            }

            HStack {
                // Type
                Label(vote.voteType.displayName, systemImage: vote.voteType.iconName)
                    .font(.caption)

                Spacer()

                // Options count
                Label("\(vote.options.count) options", systemImage: "list.bullet")
                    .font(.caption)

                // Voters
                Label("\(vote.totalVoters)", systemImage: "person.fill")
                    .font(.caption)
            }
            .foregroundStyle(.secondary)

            // Time remaining
            if let timeRemaining = vote.timeRemainingText {
                Text(timeRemaining)
                    .font(.caption.bold())
                    .foregroundStyle(vote.isActive ? .green : .secondary)
            }
        }
        .padding(.vertical, 4)
    }
}

// MARK: - Vote Status Badge

struct VoteStatusBadge: View {
    let status: VoteStatus

    var color: Color {
        switch status {
        case .draft: return .gray
        case .active: return .green
        case .ended: return .purple
        case .cancelled: return .red
        }
    }

    var body: some View {
        Text(status.displayName)
            .font(.caption2.bold())
            .padding(.horizontal, 8)
            .padding(.vertical, 4)
            .background(color.opacity(0.2))
            .foregroundStyle(color)
            .clipShape(Capsule())
    }
}

#Preview {
    NavigationStack {
        VoteListView(guildId: "test-guild")
    }
}
