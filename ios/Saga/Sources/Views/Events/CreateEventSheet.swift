import SwiftUI

/// Sheet for creating a new event
struct CreateEventSheet: View {
    let guildId: String

    @Environment(\.dismiss) private var dismiss
    @Environment(EventService.self) private var eventService

    @State private var title = ""
    @State private var description = ""
    @State private var location = ""
    @State private var startDate = Date().addingTimeInterval(3600) // 1 hour from now
    @State private var endDate = Date().addingTimeInterval(7200) // 2 hours from now
    @State private var hasEndTime = true
    @State private var hasCapacity = false
    @State private var capacity = "20"
    @State private var visibility: EventVisibility = .guild

    @State private var isCreating = false
    @State private var errorMessage: String?

    var body: some View {
        NavigationStack {
            Form {
                // Basic info
                Section("Event Details") {
                    TextField("Title", text: $title)
                        .textContentType(.none)

                    TextField("Description (optional)", text: $description, axis: .vertical)
                        .lineLimit(3...6)
                }

                // Location
                Section("Location") {
                    TextField("Location (optional)", text: $location)
                        .textContentType(.fullStreetAddress)
                }

                // Date and time
                Section("Date & Time") {
                    DatePicker(
                        "Starts",
                        selection: $startDate,
                        in: Date()...,
                        displayedComponents: [.date, .hourAndMinute]
                    )

                    Toggle("End Time", isOn: $hasEndTime)

                    if hasEndTime {
                        DatePicker(
                            "Ends",
                            selection: $endDate,
                            in: startDate...,
                            displayedComponents: [.date, .hourAndMinute]
                        )
                    }
                }

                // Capacity and visibility
                Section("Settings") {
                    Toggle("Limit Capacity", isOn: $hasCapacity)

                    if hasCapacity {
                        HStack {
                            Text("Max attendees")
                            Spacer()
                            TextField("", text: $capacity)
                                .keyboardType(.numberPad)
                                .multilineTextAlignment(.trailing)
                                .frame(width: 60)
                        }
                    }

                    Picker("Visibility", selection: $visibility) {
                        ForEach(EventVisibility.allCases, id: \.self) { vis in
                            Label(vis.displayName, systemImage: vis.iconName)
                                .tag(vis)
                        }
                    }
                }

                // Visibility explanation
                Section {
                    switch visibility {
                    case .public:
                        Label("Anyone can discover and RSVP to this event", systemImage: "globe")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    case .guild:
                        Label("Only guild members can see and RSVP", systemImage: "person.3")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    case .private:
                        Label("Only invited people can see this event", systemImage: "lock")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                }

                // Error display
                if let error = errorMessage {
                    Section {
                        Text(error)
                            .foregroundStyle(.red)
                    }
                }
            }
            .navigationTitle("New Event")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") {
                        dismiss()
                    }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Create") {
                        Task { await createEvent() }
                    }
                    .disabled(title.trimmingCharacters(in: .whitespaces).isEmpty || isCreating)
                }
            }
            .interactiveDismissDisabled(isCreating)
        }
    }

    private func createEvent() async {
        isCreating = true
        errorMessage = nil

        do {
            _ = try await eventService.createEvent(
                guildId: guildId,
                title: title.trimmingCharacters(in: .whitespaces),
                description: description.isEmpty ? nil : description,
                location: location.isEmpty ? nil : location,
                startTime: startDate,
                endTime: hasEndTime ? endDate : nil,
                capacity: hasCapacity ? Int(capacity) : nil,
                visibility: visibility
            )
            dismiss()
        } catch {
            errorMessage = error.localizedDescription
        }

        isCreating = false
    }
}

#Preview {
    CreateEventSheet(guildId: "guild:1")
        .environment(EventService.shared)
}
