import SwiftUI

/// Detailed view of an adventure
struct AdventureDetailView: View {
    let adventureId: String

    @State private var details: AdventureWithDetails?
    @State private var isLoading = true
    @State private var error: Error?
    @State private var isJoining = false
    @State private var joinMessage = ""
    @State private var showJoinSheet = false

    private let apiClient = APIClient.shared

    var adventure: Adventure? { details?.adventure }

    var body: some View {
        Group {
            if isLoading && details == nil {
                ProgressView("Loading adventure...")
            } else if let adventure = adventure {
                adventureContent(adventure)
            } else {
                ContentUnavailableView {
                    Label("Adventure Not Found", systemImage: "exclamationmark.triangle")
                } description: {
                    Text("Unable to load this adventure")
                }
            }
        }
        .navigationTitle(adventure?.title ?? "Adventure")
        .navigationBarTitleDisplayMode(.inline)
        .refreshable {
            await loadDetails()
        }
        .sheet(isPresented: $showJoinSheet) {
            joinSheet
        }
        .task {
            await loadDetails()
        }
    }

    @ViewBuilder
    private func adventureContent(_ adventure: Adventure) -> some View {
        List {
            // Header Section
            Section {
                VStack(alignment: .leading, spacing: 12) {
                    HStack {
                        AdventureStatusBadge(status: adventure.status)
                        Spacer()
                        if let spots = adventure.spotsRemainingText {
                            Text(spots)
                                .font(.subheadline.bold())
                                .foregroundStyle(adventure.hasAvailableSpots ? .green : .orange)
                        }
                    }

                    if let description = adventure.description {
                        Text(description)
                            .font(.body)
                    }

                    // Time
                    HStack {
                        Image(systemName: "calendar")
                            .foregroundStyle(.blue)
                        Text(adventure.startTime.formatted(date: .complete, time: .shortened))
                    }
                    .font(.subheadline)

                    // Location
                    if let location = adventure.location {
                        HStack {
                            Image(systemName: "mappin")
                                .foregroundStyle(.red)
                            Text(location)
                        }
                        .font(.subheadline)
                    }

                    // Admission Type
                    HStack {
                        Image(systemName: adventure.admissionType.iconName)
                            .foregroundStyle(.purple)
                        VStack(alignment: .leading) {
                            Text(adventure.admissionType.displayName)
                                .font(.subheadline.bold())
                            Text(adventure.admissionType.description)
                                .font(.caption)
                                .foregroundStyle(.secondary)
                        }
                    }
                }
                .padding(.vertical, 4)
            }

            // Admission Criteria Section
            if let criteria = adventure.admissionCriteria, adventure.admissionType == .criteria {
                Section("Admission Criteria") {
                    if let trustLevel = criteria.minTrustLevel {
                        Label("Minimum Trust: \(trustLevel.displayName)", systemImage: "shield")
                    }
                    if let compatibility = criteria.minCompatibility {
                        Label("Min Compatibility: \(Int(compatibility * 100))%", systemImage: "heart")
                    }
                    if let irl = criteria.minIrlConfirmations, irl > 0 {
                        Label("Min IRL Confirmations: \(irl)", systemImage: "person.2")
                    }
                    if criteria.guildMemberOnly {
                        Label("Guild Members Only", systemImage: "person.3")
                    }
                }
            }

            // My Status Section
            if let myStatus = details?.myStatus {
                Section("Your Status") {
                    HStack {
                        Image(systemName: statusIcon(for: myStatus))
                            .foregroundStyle(statusColor(for: myStatus))
                        Text(myStatus.displayName)
                            .font(.headline)
                    }

                    if myStatus == .approved {
                        Button(role: .destructive) {
                            Task { await leaveAdventure() }
                        } label: {
                            Label("Leave Adventure", systemImage: "xmark.circle")
                        }
                    }
                }
            } else if adventure.status == .open && adventure.hasAvailableSpots {
                Section {
                    Button {
                        if adventure.admissionType == .open {
                            Task { await joinAdventure(message: nil) }
                        } else {
                            showJoinSheet = true
                        }
                    } label: {
                        HStack {
                            Spacer()
                            if isJoining {
                                ProgressView()
                            } else {
                                Label("Join Adventure", systemImage: "plus.circle")
                            }
                            Spacer()
                        }
                    }
                    .disabled(isJoining)
                }
            }

            // Participants Section
            if let participants = details?.participants, !participants.isEmpty {
                Section("Participants (\(participants.count))") {
                    ForEach(participants, id: \.participant.id) { display in
                        HStack {
                            ZStack {
                                Circle()
                                    .fill(.blue.gradient)
                                    .frame(width: 36, height: 36)
                                Text(display.user.initials)
                                    .font(.caption.bold())
                                    .foregroundStyle(.white)
                            }

                            VStack(alignment: .leading) {
                                Text(display.user.displayName)
                                    .font(.subheadline)
                                if let role = display.participant.role {
                                    Text(role)
                                        .font(.caption)
                                        .foregroundStyle(.secondary)
                                }
                            }

                            Spacer()

                            Text(display.participant.status.displayName)
                                .font(.caption)
                                .foregroundStyle(.secondary)
                        }
                    }
                }
            }
        }
    }

    @ViewBuilder
    private var joinSheet: some View {
        NavigationStack {
            Form {
                Section("Request to Join") {
                    TextField("Message (optional)", text: $joinMessage, axis: .vertical)
                        .lineLimit(3...6)
                }

                Section {
                    Text("Your request will be reviewed by the host.")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }
            }
            .navigationTitle("Join Adventure")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") {
                        showJoinSheet = false
                    }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Send Request") {
                        Task {
                            await joinAdventure(message: joinMessage.isEmpty ? nil : joinMessage)
                            showJoinSheet = false
                        }
                    }
                    .disabled(isJoining)
                }
            }
        }
        .presentationDetents([.medium])
    }

    private func loadDetails() async {
        isLoading = true
        do {
            details = try await apiClient.getAdventure(adventureId: adventureId)
        } catch {
            self.error = error
        }
        isLoading = false
    }

    private func joinAdventure(message: String?) async {
        isJoining = true
        do {
            _ = try await apiClient.joinAdventure(adventureId: adventureId, message: message)
            await loadDetails()
        } catch {
            self.error = error
        }
        isJoining = false
    }

    private func leaveAdventure() async {
        do {
            try await apiClient.leaveAdventure(adventureId: adventureId)
            await loadDetails()
        } catch {
            self.error = error
        }
    }

    private func statusIcon(for status: ParticipantStatus) -> String {
        switch status {
        case .pending: return "clock"
        case .approved: return "checkmark.circle.fill"
        case .rejected: return "xmark.circle.fill"
        case .withdrawn: return "arrow.uturn.left"
        }
    }

    private func statusColor(for status: ParticipantStatus) -> Color {
        switch status {
        case .pending: return .orange
        case .approved: return .green
        case .rejected: return .red
        case .withdrawn: return .gray
        }
    }
}

#Preview {
    NavigationStack {
        AdventureDetailView(adventureId: "test")
    }
}
