import SwiftUI

/// Sheet for creating a new adventure
struct CreateAdventureSheet: View {
    let guildId: String

    @Environment(\.dismiss) private var dismiss

    @State private var title = ""
    @State private var description = ""
    @State private var location = ""
    @State private var startTime = Date().addingTimeInterval(3600) // 1 hour from now
    @State private var endTime: Date?
    @State private var hasEndTime = false
    @State private var maxParticipants: Int?
    @State private var hasMaxParticipants = false
    @State private var admissionType: AdmissionType = .open
    @State private var minTrustLevel: TrustLevel?
    @State private var minCompatibility: Double?
    @State private var guildMemberOnly = false
    @State private var isCreating = false
    @State private var error: Error?

    private let apiClient = APIClient.shared

    var body: some View {
        Form {
            // Basic Info
            Section("Details") {
                TextField("Title", text: $title)
                TextField("Description (optional)", text: $description, axis: .vertical)
                    .lineLimit(3...6)
                TextField("Location (optional)", text: $location)
            }

            // Schedule
            Section("Schedule") {
                DatePicker("Start Time", selection: $startTime, in: Date()...)

                Toggle("Has End Time", isOn: $hasEndTime)

                if hasEndTime {
                    DatePicker("End Time", selection: Binding(
                        get: { endTime ?? startTime.addingTimeInterval(7200) },
                        set: { endTime = $0 }
                    ), in: startTime...)
                }
            }

            // Capacity
            Section("Capacity") {
                Toggle("Limit Participants", isOn: $hasMaxParticipants)

                if hasMaxParticipants {
                    Stepper(
                        "Max: \(maxParticipants ?? 10)",
                        value: Binding(
                            get: { maxParticipants ?? 10 },
                            set: { maxParticipants = $0 }
                        ),
                        in: 2...100
                    )
                }
            }

            // Admission
            Section {
                Picker("Admission Type", selection: $admissionType) {
                    ForEach(AdmissionType.allCases, id: \.self) { type in
                        Label(type.displayName, systemImage: type.iconName)
                            .tag(type)
                    }
                }
            } header: {
                Text("Admission")
            } footer: {
                Text(admissionType.description)
            }

            // Criteria (if applicable)
            if admissionType == .criteria {
                Section("Admission Criteria") {
                    Picker("Minimum Trust Level", selection: $minTrustLevel) {
                        Text("None").tag(nil as TrustLevel?)
                        ForEach(TrustLevel.allCases, id: \.self) { level in
                            Text(level.displayName).tag(level as TrustLevel?)
                        }
                    }

                    Toggle("Guild Members Only", isOn: $guildMemberOnly)
                }
            }
        }
        .navigationTitle("New Adventure")
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .cancellationAction) {
                Button("Cancel") {
                    dismiss()
                }
            }
            ToolbarItem(placement: .confirmationAction) {
                Button("Create") {
                    Task { await createAdventure() }
                }
                .disabled(title.isEmpty || isCreating)
            }
        }
        .alert("Error", isPresented: .constant(error != nil)) {
            Button("OK") { error = nil }
        } message: {
            if let error = error {
                Text(error.localizedDescription)
            }
        }
    }

    private func createAdventure() async {
        isCreating = true

        var criteria: AdmissionCriteria?
        if admissionType == .criteria {
            criteria = AdmissionCriteria(
                minTrustLevel: minTrustLevel,
                minCompatibility: minCompatibility,
                requiredInterests: nil,
                minIrlConfirmations: nil,
                guildMemberOnly: guildMemberOnly
            )
        }

        let request = CreateAdventureRequest(
            guildId: guildId,
            title: title,
            description: description.isEmpty ? nil : description,
            location: location.isEmpty ? nil : location,
            locationLat: nil,
            locationLng: nil,
            startTime: startTime,
            endTime: hasEndTime ? endTime : nil,
            capacity: hasMaxParticipants ? maxParticipants : nil,
            admissionType: admissionType,
            admissionCriteria: criteria
        )

        do {
            _ = try await apiClient.createAdventure(request)
            dismiss()
        } catch {
            self.error = error
        }

        isCreating = false
    }
}

#Preview {
    NavigationStack {
        CreateAdventureSheet(guildId: "test")
    }
}
