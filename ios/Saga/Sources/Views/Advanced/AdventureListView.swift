import SwiftUI

/// View listing adventures for a guild
struct AdventureListView: View {
    let guildId: String

    @State private var adventures: [Adventure] = []
    @State private var isLoading = true
    @State private var error: Error?
    @State private var showCreateSheet = false

    private let apiClient = APIClient.shared

    var body: some View {
        Group {
            if isLoading && adventures.isEmpty {
                ProgressView("Loading adventures...")
            } else if adventures.isEmpty {
                ContentUnavailableView {
                    Label("No Adventures", systemImage: "figure.hiking")
                } description: {
                    Text("Create an adventure to gather people for group activities")
                } actions: {
                    Button("Create Adventure") {
                        showCreateSheet = true
                    }
                    .buttonStyle(.borderedProminent)
                }
            } else {
                List {
                    ForEach(adventures) { adventure in
                        NavigationLink {
                            AdventureDetailView(adventureId: adventure.id)
                        } label: {
                            AdventureRow(adventure: adventure)
                        }
                    }
                }
            }
        }
        .navigationTitle("Adventures")
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
            await loadAdventures()
        }
        .sheet(isPresented: $showCreateSheet) {
            NavigationStack {
                CreateAdventureSheet(guildId: guildId)
            }
        }
        .task {
            await loadAdventures()
        }
    }

    private func loadAdventures() async {
        isLoading = true
        do {
            adventures = try await apiClient.getAdventures(guildId: guildId)
        } catch {
            self.error = error
        }
        isLoading = false
    }
}

// MARK: - Adventure Row

struct AdventureRow: View {
    let adventure: Adventure

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Text(adventure.title)
                    .font(.headline)

                Spacer()

                AdventureStatusBadge(status: adventure.status)
            }

            if let description = adventure.description {
                Text(description)
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
                    .lineLimit(2)
            }

            HStack {
                // Date
                Label(adventure.startTime.formatted(date: .abbreviated, time: .shortened), systemImage: "calendar")
                    .font(.caption)

                Spacer()

                // Admission type
                Label(adventure.admissionType.displayName, systemImage: adventure.admissionType.iconName)
                    .font(.caption)

                // Participants
                if let spots = adventure.spotsRemainingText {
                    Text(spots)
                        .font(.caption.bold())
                        .foregroundStyle(adventure.hasAvailableSpots ? .green : .orange)
                }
            }
            .foregroundStyle(.secondary)
        }
        .padding(.vertical, 4)
    }
}

// MARK: - Adventure Status Badge

struct AdventureStatusBadge: View {
    let status: AdventureStatus

    var color: Color {
        switch status {
        case .draft: return .gray
        case .open: return .green
        case .full: return .orange
        case .inProgress: return .blue
        case .completed: return .purple
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
        AdventureListView(guildId: "test-guild")
    }
}
