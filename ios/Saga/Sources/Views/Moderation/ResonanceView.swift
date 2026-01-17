import SwiftUI

/// View showing user's resonance score and breakdown
struct ResonanceView: View {
    @State private var resonance: Resonance?
    @State private var isLoading = true
    @State private var error: Error?

    private let apiClient = APIClient.shared

    var body: some View {
        ScrollView {
            if isLoading && resonance == nil {
                ProgressView("Loading...")
                    .padding(.top, 100)
            } else if let resonance = resonance {
                resonanceContent(resonance)
            } else {
                ContentUnavailableView {
                    Label("Unable to Load", systemImage: "exclamationmark.triangle")
                } description: {
                    Text("Could not load your resonance score")
                }
            }
        }
        .navigationTitle("Resonance")
        .refreshable {
            await loadResonance()
        }
        .task {
            await loadResonance()
        }
    }

    @ViewBuilder
    private func resonanceContent(_ resonance: Resonance) -> some View {
        VStack(spacing: 24) {
            // Level card
            levelCard(resonance)

            // Score breakdown
            breakdownSection(resonance.breakdown)

            // Recent activity
            if !resonance.recentActivity.isEmpty {
                activitySection(resonance.recentActivity)
            }
        }
        .padding()
    }

    @ViewBuilder
    private func levelCard(_ resonance: Resonance) -> some View {
        VStack(spacing: 16) {
            // Level icon
            ZStack {
                Circle()
                    .fill((Color(hex: resonance.level.color) ?? .purple).gradient)
                    .frame(width: 80, height: 80)
                Image(systemName: resonance.level.iconName)
                    .font(.largeTitle)
                    .foregroundStyle(.white)
            }

            // Level name
            Text(resonance.level.displayName)
                .font(.title.bold())

            Text(resonance.level.description)
                .font(.subheadline)
                .foregroundStyle(.secondary)

            // Score
            Text("\(resonance.totalScore) points")
                .font(.headline)
                .foregroundStyle(.secondary)

            // Progress to next level
            if let nextLevel = resonance.level.nextLevel {
                VStack(spacing: 8) {
                    ProgressView(value: resonance.progressToNextLevel)
                        .tint(Color(hex: resonance.level.color) ?? .purple)

                    if let pointsNeeded = resonance.pointsToNextLevel {
                        Text("\(pointsNeeded) points to \(nextLevel.displayName)")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                }
                .padding(.horizontal)
            } else {
                Text("Maximum level reached!")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }
        }
        .padding()
        .frame(maxWidth: .infinity)
        .background(.ultraThinMaterial)
        .clipShape(RoundedRectangle(cornerRadius: 16))
    }

    @ViewBuilder
    private func breakdownSection(_ breakdown: ResonanceBreakdown) -> some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("Score Breakdown")
                .font(.headline)

            LazyVGrid(columns: [
                GridItem(.flexible()),
                GridItem(.flexible())
            ], spacing: 12) {
                ForEach(breakdown.categories, id: \.name) { category in
                    HStack {
                        Image(systemName: category.icon)
                            .foregroundStyle(.blue)
                            .frame(width: 24)

                        VStack(alignment: .leading) {
                            Text(category.name)
                                .font(.caption)
                                .foregroundStyle(.secondary)
                            Text("\(category.score)")
                                .font(.subheadline.bold())
                        }

                        Spacer()
                    }
                    .padding()
                    .background(.ultraThinMaterial)
                    .clipShape(RoundedRectangle(cornerRadius: 8))
                }
            }
        }
    }

    @ViewBuilder
    private func activitySection(_ activities: [ResonanceActivity]) -> some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("Recent Activity")
                .font(.headline)

            ForEach(activities) { activity in
                HStack {
                    Image(systemName: activity.type.iconName)
                        .foregroundStyle(.green)
                        .frame(width: 24)

                    VStack(alignment: .leading) {
                        Text(activity.description)
                            .font(.subheadline)
                        Text(activity.createdOn.formatted(date: .abbreviated, time: .shortened))
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }

                    Spacer()

                    Text("+\(activity.points)")
                        .font(.subheadline.bold())
                        .foregroundStyle(.green)
                }
                .padding()
                .background(.ultraThinMaterial)
                .clipShape(RoundedRectangle(cornerRadius: 8))
            }
        }
    }

    private func loadResonance() async {
        isLoading = true
        do {
            resonance = try await apiClient.getMyResonance()
        } catch {
            self.error = error
        }
        isLoading = false
    }
}

#Preview {
    NavigationStack {
        ResonanceView()
    }
}
