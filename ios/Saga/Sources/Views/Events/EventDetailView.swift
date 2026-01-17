import SwiftUI

/// Detail view for an event showing RSVPs, roles, and actions
struct EventDetailView: View {
    let event: Event

    @Environment(EventService.self) private var eventService
    @Environment(AuthService.self) private var authService

    @State private var isLoading = true
    @State private var showingEditSheet = false
    @State private var showingCancelAlert = false
    @State private var showingAddRoleSheet = false
    @State private var errorMessage: String?
    @State private var showingError = false

    private var details: EventWithDetails? {
        eventService.currentEventDetails
    }

    var body: some View {
        ScrollView {
            VStack(spacing: 20) {
                if isLoading {
                    ProgressView()
                        .frame(maxWidth: .infinity, minHeight: 200)
                } else if let details = details {
                    // Event header
                    EventHeaderView(event: details.event)

                    // RSVP section
                    RSVPSection(details: details)

                    // Roles section
                    if !details.roles.isEmpty {
                        RolesSection(
                            roles: details.roles,
                            canManage: details.canManage,
                            onAddRole: { showingAddRoleSheet = true }
                        )
                    }

                    // Actions
                    if details.canManage {
                        HostActionsSection(
                            event: details.event,
                            onEdit: { showingEditSheet = true },
                            onCancel: { showingCancelAlert = true }
                        )
                    }
                } else {
                    ContentUnavailableView(
                        "Event Not Found",
                        systemImage: "calendar.badge.exclamationmark",
                        description: Text("This event may have been cancelled or deleted")
                    )
                }
            }
            .padding()
        }
        .navigationTitle(event.title)
        .navigationBarTitleDisplayMode(.inline)
        .task {
            await loadEvent()
        }
        .refreshable {
            await loadEvent()
        }
        .alert("Error", isPresented: $showingError) {
            Button("OK", role: .cancel) { }
        } message: {
            Text(errorMessage ?? "An error occurred")
        }
        .alert("Cancel Event", isPresented: $showingCancelAlert) {
            Button("Cancel Event", role: .destructive) {
                Task { await cancelEvent() }
            }
            Button("Keep Event", role: .cancel) { }
        } message: {
            Text("Are you sure you want to cancel this event? This cannot be undone.")
        }
        .sheet(isPresented: $showingEditSheet) {
            if let details = details {
                EditEventSheet(event: details.event)
            }
        }
        .sheet(isPresented: $showingAddRoleSheet) {
            AddRoleSheet(eventId: event.id)
        }
    }

    private func loadEvent() async {
        isLoading = true
        do {
            try await eventService.loadEventDetails(eventId: event.id)
        } catch {
            errorMessage = error.localizedDescription
            showingError = true
        }
        isLoading = false
    }

    private func cancelEvent() async {
        do {
            try await eventService.cancelEvent(eventId: event.id)
        } catch {
            errorMessage = error.localizedDescription
            showingError = true
        }
    }
}

// MARK: - Event Header View

struct EventHeaderView: View {
    let event: Event

    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            // Status badge
            if event.status != .published {
                StatusBadge(status: event.status)
            }

            // Date and time
            HStack(spacing: 12) {
                Image(systemName: "calendar")
                    .foregroundStyle(.blue)
                    .frame(width: 24)

                VStack(alignment: .leading, spacing: 2) {
                    Text(event.formattedDate)
                        .font(.headline)
                    Text(event.timeRange)
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                }
            }

            // Location
            if let location = event.location {
                HStack(spacing: 12) {
                    Image(systemName: "mappin.circle.fill")
                        .foregroundStyle(.red)
                        .frame(width: 24)

                    VStack(alignment: .leading, spacing: 2) {
                        Text(location)
                            .font(.subheadline)
                        if event.hasCoordinates {
                            Button("Open in Maps") {
                                openInMaps()
                            }
                            .font(.caption)
                        }
                    }
                }
            }

            // Description
            if let description = event.description, !description.isEmpty {
                Text(description)
                    .font(.body)
                    .foregroundStyle(.secondary)
            }

            // Capacity info
            if let capacity = event.capacity {
                HStack(spacing: 12) {
                    Image(systemName: "person.3")
                        .foregroundStyle(.orange)
                        .frame(width: 24)

                    if event.isFull {
                        Text("Full (\(capacity) spots)")
                            .font(.subheadline)
                            .foregroundStyle(.orange)
                    } else {
                        Text("\(event.spotsRemaining ?? 0) of \(capacity) spots remaining")
                            .font(.subheadline)
                            .foregroundStyle(.secondary)
                    }
                }
            }
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding()
        .background(.ultraThinMaterial)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }

    private func openInMaps() {
        guard let lat = event.locationLat, let lng = event.locationLng else { return }
        let url = URL(string: "maps://?ll=\(lat),\(lng)")!
        UIApplication.shared.open(url)
    }
}

// MARK: - RSVP Section

struct RSVPSection: View {
    let details: EventWithDetails

    @Environment(EventService.self) private var eventService
    @State private var isUpdating = false

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            // Section header with counts
            HStack {
                Text("RSVPs")
                    .font(.headline)

                Spacer()

                HStack(spacing: 12) {
                    Label("\(details.goingCount)", systemImage: "checkmark.circle.fill")
                        .foregroundStyle(.green)
                    Label("\(details.maybeCount)", systemImage: "questionmark.circle.fill")
                        .foregroundStyle(.orange)
                }
                .font(.subheadline)
            }

            // My RSVP buttons
            if details.event.status == .published && !details.event.isPast {
                HStack(spacing: 12) {
                    RSVPButton(
                        status: .going,
                        isSelected: details.myRsvp?.status == .going,
                        isLoading: isUpdating
                    ) {
                        await updateRSVP(.going)
                    }

                    RSVPButton(
                        status: .maybe,
                        isSelected: details.myRsvp?.status == .maybe,
                        isLoading: isUpdating
                    ) {
                        await updateRSVP(.maybe)
                    }

                    RSVPButton(
                        status: .notGoing,
                        isSelected: details.myRsvp?.status == .notGoing,
                        isLoading: isUpdating
                    ) {
                        await updateRSVP(.notGoing)
                    }
                }
            }

            // RSVP list
            if !details.rsvps.isEmpty {
                Divider()

                VStack(spacing: 8) {
                    // Going
                    if !details.goingRsvps.isEmpty {
                        RSVPList(title: "Going", rsvps: details.goingRsvps, color: .green)
                    }

                    // Maybe
                    if !details.maybeRsvps.isEmpty {
                        RSVPList(title: "Maybe", rsvps: details.maybeRsvps, color: .orange)
                    }
                }
            }
        }
        .padding()
        .background(.ultraThinMaterial)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }

    private func updateRSVP(_ status: RSVPStatus) async {
        isUpdating = true
        do {
            if details.myRsvp?.status == status {
                try await eventService.cancelRSVP(eventId: details.event.id)
            } else {
                _ = try await eventService.rsvp(eventId: details.event.id, status: status)
            }
        } catch {
            // Error handling done in service
        }
        isUpdating = false
    }
}

// MARK: - RSVP Button

struct RSVPButton: View {
    let status: RSVPStatus
    let isSelected: Bool
    let isLoading: Bool
    let action: () async -> Void

    var body: some View {
        Button {
            Task { await action() }
        } label: {
            HStack {
                if isLoading && isSelected {
                    ProgressView()
                        .scaleEffect(0.8)
                } else {
                    Image(systemName: status.iconName)
                }
                Text(status.displayName)
            }
            .font(.subheadline.weight(isSelected ? .semibold : .regular))
            .frame(maxWidth: .infinity)
            .padding(.vertical, 10)
            .background(isSelected ? statusColor.opacity(0.2) : Color.gray.opacity(0.1))
            .foregroundStyle(isSelected ? statusColor : .primary)
            .clipShape(RoundedRectangle(cornerRadius: 8))
        }
        .disabled(isLoading)
    }

    private var statusColor: Color {
        switch status {
        case .going: return .green
        case .maybe: return .orange
        case .notGoing: return .red
        }
    }
}

// MARK: - RSVP List

struct RSVPList: View {
    let title: String
    let rsvps: [RSVP]
    let color: Color

    var body: some View {
        VStack(alignment: .leading, spacing: 4) {
            Text(title)
                .font(.caption.bold())
                .foregroundStyle(color)

            // Show user IDs for now (would show names in real app)
            Text(rsvps.map { String($0.userId.suffix(8)) }.joined(separator: ", "))
                .font(.caption)
                .foregroundStyle(.secondary)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
    }
}

// MARK: - Roles Section

struct RolesSection: View {
    let roles: [EventRole]
    let canManage: Bool
    let onAddRole: () -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Text("Roles")
                    .font(.headline)

                Spacer()

                if canManage {
                    Button {
                        onAddRole()
                    } label: {
                        Image(systemName: "plus.circle")
                    }
                }
            }

            ForEach(roles) { role in
                RoleRow(role: role)
            }
        }
        .padding()
        .background(.ultraThinMaterial)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }
}

// MARK: - Role Row

struct RoleRow: View {
    let role: EventRole

    @Environment(EventService.self) private var eventService
    @State private var isAssigning = false

    var body: some View {
        HStack {
            VStack(alignment: .leading, spacing: 2) {
                Text(role.name)
                    .font(.subheadline.bold())

                if let description = role.description {
                    Text(description)
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }

                Text(role.availabilityText)
                    .font(.caption2)
                    .foregroundStyle(role.isFull ? .orange : .secondary)
            }

            Spacer()

            if !role.isFull {
                Button {
                    Task { await assignSelf() }
                } label: {
                    if isAssigning {
                        ProgressView()
                            .scaleEffect(0.8)
                    } else {
                        Text("Sign Up")
                            .font(.caption.bold())
                    }
                }
                .buttonStyle(.bordered)
                .disabled(isAssigning)
            } else {
                Text("Full")
                    .font(.caption)
                    .foregroundStyle(.orange)
            }
        }
        .padding(.vertical, 4)
    }

    private func assignSelf() async {
        isAssigning = true
        do {
            _ = try await eventService.assignRole(eventId: role.eventId, roleId: role.id)
        } catch {
            // Handle error
        }
        isAssigning = false
    }
}

// MARK: - Host Actions Section

struct HostActionsSection: View {
    let event: Event
    let onEdit: () -> Void
    let onCancel: () -> Void

    var body: some View {
        VStack(spacing: 12) {
            Text("Host Actions")
                .font(.headline)
                .frame(maxWidth: .infinity, alignment: .leading)

            HStack(spacing: 12) {
                Button {
                    onEdit()
                } label: {
                    Label("Edit", systemImage: "pencil")
                        .frame(maxWidth: .infinity)
                }
                .buttonStyle(.bordered)

                Button(role: .destructive) {
                    onCancel()
                } label: {
                    Label("Cancel Event", systemImage: "xmark.circle")
                        .frame(maxWidth: .infinity)
                }
                .buttonStyle(.bordered)
            }
        }
        .padding()
        .background(.ultraThinMaterial)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }
}

// MARK: - Edit Event Sheet

struct EditEventSheet: View {
    let event: Event

    @Environment(\.dismiss) private var dismiss
    @Environment(EventService.self) private var eventService

    @State private var title: String
    @State private var description: String
    @State private var location: String
    @State private var startTime: Date
    @State private var endTime: Date
    @State private var hasEndTime: Bool
    @State private var capacity: String
    @State private var visibility: EventVisibility
    @State private var isSaving = false
    @State private var errorMessage: String?

    init(event: Event) {
        self.event = event
        _title = State(initialValue: event.title)
        _description = State(initialValue: event.description ?? "")
        _location = State(initialValue: event.location ?? "")
        _startTime = State(initialValue: event.startTime)
        _endTime = State(initialValue: event.endTime ?? event.startTime.addingTimeInterval(3600))
        _hasEndTime = State(initialValue: event.endTime != nil)
        _capacity = State(initialValue: event.capacity.map(String.init) ?? "")
        _visibility = State(initialValue: event.visibility)
    }

    var body: some View {
        NavigationStack {
            Form {
                Section("Event Details") {
                    TextField("Title", text: $title)
                    TextField("Description", text: $description, axis: .vertical)
                        .lineLimit(3...6)
                }

                Section("Location") {
                    TextField("Location", text: $location)
                }

                Section("Date & Time") {
                    DatePicker("Start", selection: $startTime)

                    Toggle("End Time", isOn: $hasEndTime)

                    if hasEndTime {
                        DatePicker("End", selection: $endTime)
                    }
                }

                Section("Capacity") {
                    TextField("Max attendees (optional)", text: $capacity)
                        .keyboardType(.numberPad)

                    Picker("Visibility", selection: $visibility) {
                        ForEach(EventVisibility.allCases, id: \.self) { vis in
                            Label(vis.displayName, systemImage: vis.iconName)
                                .tag(vis)
                        }
                    }
                }

                if let error = errorMessage {
                    Section {
                        Text(error)
                            .foregroundStyle(.red)
                    }
                }
            }
            .navigationTitle("Edit Event")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Save") {
                        Task { await save() }
                    }
                    .disabled(title.isEmpty || isSaving)
                }
            }
        }
    }

    private func save() async {
        isSaving = true
        errorMessage = nil

        let request = UpdateEventRequest(
            title: title,
            description: description.isEmpty ? nil : description,
            location: location.isEmpty ? nil : location,
            startTime: startTime,
            endTime: hasEndTime ? endTime : nil,
            capacity: Int(capacity),
            visibility: visibility
        )

        do {
            _ = try await eventService.updateEvent(eventId: event.id, request)
            dismiss()
        } catch {
            errorMessage = error.localizedDescription
        }

        isSaving = false
    }
}

// MARK: - Add Role Sheet

struct AddRoleSheet: View {
    let eventId: String

    @Environment(\.dismiss) private var dismiss
    @Environment(EventService.self) private var eventService

    @State private var name = ""
    @State private var description = ""
    @State private var maxSlots = 1
    @State private var isCreating = false
    @State private var errorMessage: String?

    var body: some View {
        NavigationStack {
            Form {
                Section("Role Details") {
                    TextField("Name", text: $name)
                    TextField("Description (optional)", text: $description)
                }

                Section("Slots") {
                    Stepper("Max slots: \(maxSlots)", value: $maxSlots, in: 1...50)
                }

                if let error = errorMessage {
                    Section {
                        Text(error)
                            .foregroundStyle(.red)
                    }
                }
            }
            .navigationTitle("Add Role")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Add") {
                        Task { await createRole() }
                    }
                    .disabled(name.isEmpty || isCreating)
                }
            }
        }
    }

    private func createRole() async {
        isCreating = true
        errorMessage = nil

        do {
            _ = try await eventService.createRole(
                eventId: eventId,
                name: name,
                description: description.isEmpty ? nil : description,
                maxSlots: maxSlots
            )
            dismiss()
        } catch {
            errorMessage = error.localizedDescription
        }

        isCreating = false
    }
}

#Preview {
    NavigationStack {
        EventDetailView(event: Event(
            id: "event:1",
            guildId: "guild:1",
            hostId: "user:1",
            title: "Board Game Night",
            description: "Let's play some board games!",
            location: "123 Main St",
            locationLat: nil,
            locationLng: nil,
            startTime: Date().addingTimeInterval(86400),
            endTime: Date().addingTimeInterval(86400 + 10800),
            capacity: 10,
            rsvpCount: 5,
            status: .published,
            visibility: .guild,
            createdOn: Date(),
            updatedOn: Date()
        ))
        .environment(EventService.shared)
        .environment(AuthService.shared)
    }
}
