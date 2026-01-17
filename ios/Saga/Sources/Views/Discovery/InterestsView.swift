import SwiftUI

/// View for managing user's interests
struct InterestsView: View {
    @Environment(DiscoveryService.self) private var discoveryService

    @State private var showAddInterest = false

    var body: some View {
        List {
            // My Interests Section
            if discoveryService.myInterests.isEmpty {
                Section {
                    ContentUnavailableView {
                        Label("No Interests", systemImage: "star")
                    } description: {
                        Text("Add interests to find people with similar passions")
                    } actions: {
                        Button("Add Interest") {
                            showAddInterest = true
                        }
                        .buttonStyle(.borderedProminent)
                    }
                }
                .listRowBackground(Color.clear)
            } else {
                // Group by category
                let grouped = Dictionary(grouping: discoveryService.myInterests) { $0.categoryName ?? "Other" }

                ForEach(grouped.keys.sorted(), id: \.self) { category in
                    Section(category) {
                        ForEach(grouped[category] ?? []) { interest in
                            UserInterestRow(interest: interest)
                                .swipeActions(edge: .trailing, allowsFullSwipe: true) {
                                    Button(role: .destructive) {
                                        Task {
                                            try? await discoveryService.removeInterest(userInterestId: interest.id)
                                        }
                                    } label: {
                                        Label("Remove", systemImage: "trash")
                                    }
                                }
                        }
                    }
                }
            }

            // Questionnaires Section
            Section {
                NavigationLink {
                    QuestionnaireListView()
                } label: {
                    HStack {
                        Image(systemName: "list.clipboard.fill")
                            .foregroundStyle(.blue)
                            .frame(width: 24)

                        VStack(alignment: .leading, spacing: 2) {
                            Text("Questionnaires")
                                .font(.subheadline)
                            Text("\(Int(discoveryService.overallProgress * 100))% complete")
                                .font(.caption)
                                .foregroundStyle(.secondary)
                        }

                        Spacer()

                        // Progress ring
                        ZStack {
                            Circle()
                                .stroke(.gray.opacity(0.3), lineWidth: 4)
                                .frame(width: 30, height: 30)
                            Circle()
                                .trim(from: 0, to: discoveryService.overallProgress)
                                .stroke(.blue, style: StrokeStyle(lineWidth: 4, lineCap: .round))
                                .frame(width: 30, height: 30)
                                .rotationEffect(.degrees(-90))
                        }
                    }
                }
            }
        }
        .navigationTitle("My Interests")
        .toolbar {
            ToolbarItem(placement: .primaryAction) {
                Button {
                    showAddInterest = true
                } label: {
                    Image(systemName: "plus")
                }
            }
        }
        .refreshable {
            await discoveryService.loadMyInterests()
            await discoveryService.loadQuestionnaires()
        }
        .sheet(isPresented: $showAddInterest) {
            NavigationStack {
                AddInterestSheet()
            }
        }
        .task {
            await discoveryService.loadMyInterests()
            await discoveryService.loadQuestionnaires()
        }
    }
}

// MARK: - User Interest Row

struct UserInterestRow: View {
    let interest: UserInterest

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Text(interest.interestName ?? "Unknown")
                    .font(.headline)

                Spacer()

                // Intent badge
                HStack(spacing: 4) {
                    Image(systemName: interest.intent.iconName)
                    Text(interest.intent.displayName)
                }
                .font(.caption.bold())
                .padding(.horizontal, 8)
                .padding(.vertical, 4)
                .background(intentColor.opacity(0.2))
                .foregroundStyle(intentColor)
                .clipShape(Capsule())
            }

            // Skill Level
            HStack(spacing: 4) {
                Image(systemName: interest.skillLevel.iconName)
                Text(interest.skillLevel.displayName)
            }
            .font(.caption)
            .foregroundStyle(.secondary)

            // Notes
            if let notes = interest.notes, !notes.isEmpty {
                Text(notes)
                    .font(.caption)
                    .foregroundStyle(.secondary)
                    .lineLimit(2)
            }
        }
        .padding(.vertical, 4)
    }

    private var intentColor: Color {
        switch interest.intent {
        case .teach: return .orange
        case .learn: return .blue
        case .both: return .purple
        }
    }
}

#Preview {
    NavigationStack {
        InterestsView()
            .environment(DiscoveryService.shared)
    }
}
