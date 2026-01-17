import SwiftUI

/// View showing blocked users with ability to unblock
struct BlockedUsersView: View {
    @State private var blockedUsers: [BlockDisplay] = []
    @State private var isLoading = true
    @State private var error: Error?
    @State private var userToUnblock: BlockDisplay?

    private let apiClient = APIClient.shared

    var body: some View {
        Group {
            if isLoading && blockedUsers.isEmpty {
                ProgressView("Loading...")
            } else if blockedUsers.isEmpty {
                ContentUnavailableView {
                    Label("No Blocked Users", systemImage: "person.slash")
                } description: {
                    Text("You haven't blocked anyone")
                }
            } else {
                List {
                    ForEach(blockedUsers) { blockDisplay in
                        BlockedUserRow(blockDisplay: blockDisplay) {
                            userToUnblock = blockDisplay
                        }
                    }
                }
            }
        }
        .navigationTitle("Blocked Users")
        .refreshable {
            await loadBlockedUsers()
        }
        .task {
            await loadBlockedUsers()
        }
        .alert("Unblock User", isPresented: .constant(userToUnblock != nil)) {
            Button("Cancel", role: .cancel) {
                userToUnblock = nil
            }
            Button("Unblock", role: .destructive) {
                if let user = userToUnblock {
                    Task { await unblockUser(user) }
                }
            }
        } message: {
            if let user = userToUnblock {
                Text("Are you sure you want to unblock \(user.blockedUser.displayName)?")
            }
        }
    }

    private func loadBlockedUsers() async {
        isLoading = true
        do {
            blockedUsers = try await apiClient.getBlockedUsers()
        } catch {
            self.error = error
        }
        isLoading = false
    }

    private func unblockUser(_ blockDisplay: BlockDisplay) async {
        do {
            try await apiClient.unblockUser(blockedId: blockDisplay.block.blockedId)
            blockedUsers.removeAll { $0.id == blockDisplay.id }
        } catch {
            self.error = error
        }
        userToUnblock = nil
    }
}

// MARK: - Blocked User Row

struct BlockedUserRow: View {
    let blockDisplay: BlockDisplay
    let onUnblock: () -> Void

    var body: some View {
        HStack(spacing: 12) {
            // Avatar
            ZStack {
                Circle()
                    .fill(.gray.opacity(0.3))
                    .frame(width: 44, height: 44)
                Text(blockDisplay.blockedUser.initials)
                    .font(.headline)
                    .foregroundStyle(.secondary)
            }

            // Info
            VStack(alignment: .leading, spacing: 2) {
                Text(blockDisplay.blockedUser.displayName)
                    .font(.headline)

                Text("Blocked \(blockDisplay.block.createdOn.formatted(date: .abbreviated, time: .omitted))")
                    .font(.caption)
                    .foregroundStyle(.secondary)

                if let reason = blockDisplay.block.reason {
                    Text(reason)
                        .font(.caption)
                        .foregroundStyle(.secondary)
                        .lineLimit(1)
                }
            }

            Spacer()

            Button {
                onUnblock()
            } label: {
                Text("Unblock")
                    .font(.subheadline)
            }
            .buttonStyle(.bordered)
        }
        .padding(.vertical, 4)
    }
}

#Preview {
    NavigationStack {
        BlockedUsersView()
    }
}
