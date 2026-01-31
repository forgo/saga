import SwiftUI

/// Detailed view of a vote
struct VoteDetailView: View {
    let voteId: String

    @State private var details: VoteWithDetails?
    @State private var isLoading = true
    @State private var error: Error?
    @State private var showCastBallot = false

    private let apiClient = APIClient.shared

    var vote: Vote? { details?.vote }

    var body: some View {
        Group {
            if isLoading && details == nil {
                ProgressView("Loading vote...")
            } else if let vote = vote {
                voteContent(vote)
            } else {
                ContentUnavailableView {
                    Label("Vote Not Found", systemImage: "exclamationmark.triangle")
                } description: {
                    Text("Unable to load this vote")
                }
            }
        }
        .navigationTitle(vote?.title ?? "Vote")
        .navigationBarTitleDisplayMode(.inline)
        .refreshable {
            await loadDetails()
        }
        .sheet(isPresented: $showCastBallot) {
            if let vote = vote {
                NavigationStack {
                    CastBallotView(vote: vote)
                }
            }
        }
        .task {
            await loadDetails()
        }
    }

    @ViewBuilder
    private func voteContent(_ vote: Vote) -> some View {
        List {
            // Header Section
            Section {
                VStack(alignment: .leading, spacing: 12) {
                    HStack {
                        VoteStatusBadge(status: vote.status)
                        Spacer()
                        if let timeRemaining = vote.timeRemainingText {
                            Text(timeRemaining)
                                .font(.subheadline.bold())
                                .foregroundStyle(vote.isActive ? .green : .secondary)
                        }
                    }

                    if let description = vote.description {
                        Text(description)
                            .font(.body)
                    }

                    // Type
                    HStack {
                        Image(systemName: vote.voteType.iconName)
                            .foregroundStyle(.blue)
                        VStack(alignment: .leading) {
                            Text(vote.voteType.displayName)
                                .font(.subheadline.bold())
                            Text(vote.voteType.description)
                                .font(.caption)
                                .foregroundStyle(.secondary)
                        }
                    }

                    // Stats
                    HStack {
                        Label("\(vote.totalVoters) voted", systemImage: "person.fill")
                        Spacer()
                        Label("\(vote.options.count) options", systemImage: "list.bullet")
                    }
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
                }
                .padding(.vertical, 4)
            }

            // My Ballot Section
            if let ballot = details?.myBallot {
                Section("Your Vote") {
                    if ballot.abstained {
                        Label("You abstained", systemImage: "hand.raised")
                            .foregroundStyle(.secondary)
                    } else {
                        ForEach(ballot.selections, id: \.optionId) { selection in
                            if let option = vote.options.first(where: { $0.id == selection.optionId }) {
                                HStack {
                                    if let rank = selection.rank {
                                        Text("#\(rank)")
                                            .font(.caption.bold())
                                            .padding(6)
                                            .background(Color.blue.opacity(0.2))
                                            .clipShape(Circle())
                                    } else if let approved = selection.approved {
                                        Image(systemName: approved ? "hand.thumbsup.fill" : "hand.thumbsdown.fill")
                                            .foregroundStyle(approved ? .green : .red)
                                    } else {
                                        Image(systemName: "checkmark.circle.fill")
                                            .foregroundStyle(.green)
                                    }

                                    Text(option.text)
                                }
                            }
                        }
                    }
                }
            } else if vote.isActive {
                Section {
                    Button {
                        showCastBallot = true
                    } label: {
                        HStack {
                            Spacer()
                            Label("Cast Your Vote", systemImage: "hand.raised")
                            Spacer()
                        }
                    }
                }
            }

            // Results Section (if available)
            if let results = details?.results {
                Section("Results") {
                    ForEach(results.optionResults, id: \.optionId) { result in
                        VStack(alignment: .leading, spacing: 4) {
                            HStack {
                                Text(result.optionText)
                                    .font(.subheadline)
                                Spacer()
                                Text(result.formattedPercentage)
                                    .font(.subheadline.bold())
                            }

                            ProgressView(value: result.percentage)
                                .tint(result.optionId == results.winner?.id ? .green : .blue)

                            HStack {
                                Text("\(result.votes) votes")
                                    .font(.caption)
                                    .foregroundStyle(.secondary)

                                if let approvals = result.approvals {
                                    Text("• \(approvals) approvals")
                                        .font(.caption)
                                        .foregroundStyle(.secondary)
                                }

                                if let avgRank = result.averageRank {
                                    Text("• Avg rank: \(String(format: "%.1f", avgRank))")
                                        .font(.caption)
                                        .foregroundStyle(.secondary)
                                }
                            }
                        }
                        .padding(.vertical, 4)
                    }

                    // Winner
                    if let winner = results.winner {
                        HStack {
                            Image(systemName: "trophy.fill")
                                .foregroundStyle(.yellow)
                            Text("Winner: \(winner.text)")
                                .font(.headline)
                        }
                        .padding(.vertical, 4)
                    }

                    // Abstained
                    if results.totalAbstained > 0 {
                        Text("\(results.totalAbstained) abstained")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                }
            }

            // Options Section (if no results yet)
            if details?.results == nil {
                Section("Options") {
                    ForEach(vote.options) { option in
                        VStack(alignment: .leading, spacing: 4) {
                            Text(option.text)
                                .font(.subheadline)
                            if let description = option.description {
                                Text(description)
                                    .font(.caption)
                                    .foregroundStyle(.secondary)
                            }
                        }
                        .padding(.vertical, 2)
                    }
                }
            }

            // Settings Section
            Section("Settings") {
                if vote.settings.anonymousVoting {
                    Label("Anonymous voting", systemImage: "eye.slash")
                }
                if vote.settings.allowAbstain {
                    Label("Abstaining allowed", systemImage: "hand.raised.slash")
                }
                if vote.settings.showResultsBeforeEnd {
                    Label("Results visible before end", systemImage: "chart.bar")
                }
                if let max = vote.settings.maxSelections {
                    Label("Max selections: \(max)", systemImage: "checklist")
                }
            }
        }
    }

    private func loadDetails() async {
        isLoading = true
        do {
            details = try await apiClient.getVote(voteId: voteId)
        } catch {
            self.error = error
        }
        isLoading = false
    }
}

#Preview {
    NavigationStack {
        VoteDetailView(voteId: "test")
    }
}
