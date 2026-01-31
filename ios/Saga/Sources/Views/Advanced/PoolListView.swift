import SwiftUI

/// View listing matching pools for a guild
struct PoolListView: View {
    let guildId: String

    @State private var pools: [Pool] = []
    @State private var isLoading = true
    @State private var error: Error?
    @State private var showCreateSheet = false

    private let apiClient = APIClient.shared

    var body: some View {
        Group {
            if isLoading && pools.isEmpty {
                ProgressView("Loading pools...")
            } else if pools.isEmpty {
                ContentUnavailableView {
                    Label("No Pools", systemImage: "person.2.circle")
                } description: {
                    Text("Create a matching pool to connect members")
                } actions: {
                    Button("Create Pool") {
                        showCreateSheet = true
                    }
                    .buttonStyle(.borderedProminent)
                }
            } else {
                List {
                    ForEach(pools) { pool in
                        NavigationLink {
                            PoolDetailView(poolId: pool.id)
                        } label: {
                            PoolRow(pool: pool)
                        }
                    }
                }
            }
        }
        .accessibilityIdentifier("pool_list")
        .navigationTitle("Matching Pools")
        .toolbar {
            ToolbarItem(placement: .primaryAction) {
                Button {
                    showCreateSheet = true
                } label: {
                    Image(systemName: "plus")
                }
                .accessibilityIdentifier("pool_create_button")
            }
        }
        .refreshable {
            await loadPools()
        }
        .sheet(isPresented: $showCreateSheet) {
            NavigationStack {
                CreatePoolSheet(guildId: guildId)
            }
        }
        .task {
            await loadPools()
        }
    }

    private func loadPools() async {
        isLoading = true
        do {
            pools = try await apiClient.getPools(guildId: guildId)
        } catch {
            self.error = error
        }
        isLoading = false
    }
}

// MARK: - Pool Row

struct PoolRow: View {
    let pool: Pool

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Text(pool.name)
                    .font(.headline)

                Spacer()

                PoolStatusBadge(isActive: pool.isActive)
            }

            if let description = pool.description {
                Text(description)
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
                    .lineLimit(2)
            }

            HStack {
                // Frequency
                Label(pool.matchingFrequency.displayName, systemImage: "clock")
                    .font(.caption)

                Spacer()

                // Member count
                Label("\(pool.memberCount) members", systemImage: "person.2")
                    .font(.caption)
            }
            .foregroundStyle(.secondary)
        }
        .padding(.vertical, 4)
    }
}

// MARK: - Pool Status Badge

struct PoolStatusBadge: View {
    let isActive: Bool

    var body: some View {
        Text(isActive ? "Active" : "Paused")
            .font(.caption2.bold())
            .padding(.horizontal, 8)
            .padding(.vertical, 4)
            .background(isActive ? Color.green.opacity(0.2) : Color.gray.opacity(0.2))
            .foregroundStyle(isActive ? .green : .gray)
            .clipShape(Capsule())
    }
}

#Preview {
    NavigationStack {
        PoolListView(guildId: "test-guild")
    }
}
