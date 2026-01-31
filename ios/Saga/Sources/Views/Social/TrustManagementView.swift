import SwiftUI

/// View for managing trust relationships
struct TrustManagementView: View {
    @Environment(ProfileService.self) private var profileService

    @State private var selectedTab: TrustTab = .granted

    enum TrustTab: String, CaseIterable {
        case granted = "I Trust"
        case received = "Trust Me"
        case irl = "IRL Pending"
    }

    var body: some View {
        VStack(spacing: 0) {
            // Tab Picker
            Picker("Tab", selection: $selectedTab) {
                ForEach(TrustTab.allCases, id: \.self) { tab in
                    Text(tab.rawValue).tag(tab)
                }
            }
            .pickerStyle(.segmented)
            .padding()

            // Content
            Group {
                switch selectedTab {
                case .granted:
                    grantedList
                case .received:
                    receivedList
                case .irl:
                    irlList
                }
            }
        }
        .navigationTitle("Trust")
        .task {
            await loadData()
        }
        .refreshable {
            await loadData()
        }
    }

    private func loadData() async {
        await profileService.loadTrustGrants()
        await profileService.loadPendingIRLRequests()
    }

    // MARK: - Granted List

    @ViewBuilder
    private var grantedList: some View {
        if profileService.myTrustGrants.isEmpty {
            ContentUnavailableView {
                Label("No Trust Grants", systemImage: "shield")
            } description: {
                Text("You haven't granted trust to anyone yet")
            }
        } else {
            List {
                ForEach(profileService.myTrustGrants) { grant in
                    TrustGrantRow(grant: grant, isIncoming: false)
                        .swipeActions(edge: .trailing, allowsFullSwipe: true) {
                            Button(role: .destructive) {
                                Task {
                                    try? await profileService.revokeTrustGrant(grantId: grant.id)
                                }
                            } label: {
                                Label("Revoke", systemImage: "xmark")
                            }
                        }
                }
            }
        }
    }

    // MARK: - Received List

    @ViewBuilder
    private var receivedList: some View {
        if profileService.trustGrantsToMe.isEmpty {
            ContentUnavailableView {
                Label("No Trust Received", systemImage: "shield.fill")
            } description: {
                Text("No one has granted you trust yet")
            }
        } else {
            List {
                ForEach(profileService.trustGrantsToMe) { grant in
                    TrustGrantRow(grant: grant, isIncoming: true)
                }
            }
        }
    }

    // MARK: - IRL List

    @ViewBuilder
    private var irlList: some View {
        if profileService.pendingIRLRequests.isEmpty {
            ContentUnavailableView {
                Label("No Pending Requests", systemImage: "person.2.fill")
            } description: {
                Text("No IRL confirmation requests pending")
            }
        } else {
            List {
                ForEach(profileService.pendingIRLRequests) { request in
                    IRLRequestRow(request: request)
                }
            }
        }
    }
}

// MARK: - Trust Grant Row

struct TrustGrantRow: View {
    let grant: TrustGrant
    let isIncoming: Bool

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Image(systemName: grant.trustLevel.iconName)
                    .foregroundStyle(.blue)

                Text(isIncoming ? "From: \(grant.grantorId)" : "To: \(grant.granteeId)")
                    .font(.headline)
                    .lineLimit(1)

                Spacer()

                Text(grant.trustLevel.displayName)
                    .font(.caption.bold())
                    .padding(.horizontal, 8)
                    .padding(.vertical, 4)
                    .background(Color.blue.opacity(0.1))
                    .foregroundStyle(.blue)
                    .clipShape(Capsule())
            }

            // Permissions
            if let permissions = grant.permissions, !permissions.isEmpty {
                HStack(spacing: 8) {
                    ForEach(permissions, id: \.self) { permission in
                        HStack(spacing: 4) {
                            Image(systemName: permission.iconName)
                            Text(permission.displayName)
                        }
                        .font(.caption2)
                        .padding(.horizontal, 6)
                        .padding(.vertical, 2)
                        .background(Color(.systemGray5))
                        .clipShape(Capsule())
                    }
                }
            }

            // Notes
            if let notes = grant.notes, !notes.isEmpty {
                Text(notes)
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }

            // Date
            Text("Granted \(grant.createdOn.formatted(date: .abbreviated, time: .omitted))")
                .font(.caption2)
                .foregroundStyle(.secondary)
        }
        .padding(.vertical, 4)
    }
}

// MARK: - IRL Request Row

struct IRLRequestRow: View {
    let request: IRLConfirmation

    @Environment(ProfileService.self) private var profileService

    @State private var isResponding = false

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Image(systemName: "person.2.fill")
                    .foregroundStyle(.orange)

                Text("From: \(request.requesterId)")
                    .font(.headline)
                    .lineLimit(1)

                Spacer()

                Text(request.context.displayName)
                    .font(.caption.bold())
                    .padding(.horizontal, 8)
                    .padding(.vertical, 4)
                    .background(Color.orange.opacity(0.1))
                    .foregroundStyle(.orange)
                    .clipShape(Capsule())
            }

            if let location = request.location {
                HStack {
                    Image(systemName: "mappin")
                        .font(.caption)
                    Text(location)
                        .font(.caption)
                }
                .foregroundStyle(.secondary)
            }

            Text("Requested \(request.createdOn.formatted(date: .abbreviated, time: .omitted))")
                .font(.caption2)
                .foregroundStyle(.secondary)

            // Action Buttons
            HStack(spacing: 12) {
                Button {
                    Task {
                        await respond(confirm: true)
                    }
                } label: {
                    Label("Confirm", systemImage: "checkmark")
                        .frame(maxWidth: .infinity)
                }
                .buttonStyle(.borderedProminent)
                .tint(.green)

                Button {
                    Task {
                        await respond(confirm: false)
                    }
                } label: {
                    Label("Decline", systemImage: "xmark")
                        .frame(maxWidth: .infinity)
                }
                .buttonStyle(.bordered)
            }
            .disabled(isResponding)
            .padding(.top, 4)
        }
        .padding(.vertical, 4)
    }

    private func respond(confirm: Bool) async {
        isResponding = true
        try? await profileService.respondToIRLRequest(confirmationId: request.id, confirm: confirm)
        isResponding = false
    }
}

#Preview {
    NavigationStack {
        TrustManagementView()
            .environment(ProfileService.shared)
    }
}
