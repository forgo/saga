import SwiftUI

/// View showing the resonance leaderboard
struct ResonanceLeaderboardView: View {
    let guildId: String?

    @State private var rankings: [ResonanceRanking] = []
    @State private var isLoading = true
    @State private var error: Error?

    private let apiClient = APIClient.shared

    init(guildId: String? = nil) {
        self.guildId = guildId
    }

    var body: some View {
        Group {
            if isLoading && rankings.isEmpty {
                ProgressView("Loading...")
            } else if rankings.isEmpty {
                ContentUnavailableView {
                    Label("No Rankings", systemImage: "trophy")
                } description: {
                    Text("No one on the leaderboard yet")
                }
            } else {
                List {
                    ForEach(rankings) { ranking in
                        LeaderboardRow(ranking: ranking)
                    }
                }
            }
        }
        .navigationTitle("Leaderboard")
        .refreshable {
            await loadLeaderboard()
        }
        .task {
            await loadLeaderboard()
        }
    }

    private func loadLeaderboard() async {
        isLoading = true
        do {
            rankings = try await apiClient.getResonanceLeaderboard(guildId: guildId)
        } catch {
            self.error = error
        }
        isLoading = false
    }
}

// MARK: - Leaderboard Row

struct LeaderboardRow: View {
    let ranking: ResonanceRanking

    var rankColor: Color {
        switch ranking.rank {
        case 1: return .yellow
        case 2: return .gray
        case 3: return .orange
        default: return .secondary
        }
    }

    var rankIcon: String {
        switch ranking.rank {
        case 1: return "trophy.fill"
        case 2: return "medal.fill"
        case 3: return "medal.fill"
        default: return ""
        }
    }

    var body: some View {
        HStack(spacing: 12) {
            // Rank
            ZStack {
                if ranking.rank <= 3 {
                    Image(systemName: rankIcon)
                        .font(.title2)
                        .foregroundStyle(rankColor)
                } else {
                    Text("#\(ranking.rank)")
                        .font(.headline.monospacedDigit())
                        .foregroundStyle(.secondary)
                }
            }
            .frame(width: 40)

            // Avatar
            ZStack {
                Circle()
                    .fill((Color(hex: ranking.level.color) ?? .purple).gradient)
                    .frame(width: 44, height: 44)
                Text(ranking.user.initials)
                    .font(.headline)
                    .foregroundStyle(.white)
            }

            // Info
            VStack(alignment: .leading, spacing: 2) {
                Text(ranking.user.displayName)
                    .font(.headline)

                HStack(spacing: 4) {
                    Image(systemName: ranking.level.iconName)
                        .font(.caption)
                    Text(ranking.level.displayName)
                        .font(.caption)
                }
                .foregroundStyle(Color(hex: ranking.level.color) ?? .purple)
            }

            Spacer()

            // Score
            VStack(alignment: .trailing) {
                Text("\(ranking.totalScore)")
                    .font(.headline.monospacedDigit())
                Text("points")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }
        }
        .padding(.vertical, 4)
    }
}

#Preview {
    NavigationStack {
        ResonanceLeaderboardView()
    }
}
