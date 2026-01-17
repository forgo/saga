import SwiftUI

/// Sheet for creating a new availability posting
struct CreateAvailabilitySheet: View {
    @Environment(ProfileService.self) private var profileService
    @Environment(\.dismiss) private var dismiss

    @State private var hangoutType: HangoutType = .meetAnyone
    @State private var title: String = ""
    @State private var description: String = ""
    @State private var startTime: Date = Date()
    @State private var endTime: Date = Date().addingTimeInterval(3600)
    @State private var location: String = ""
    @State private var radiusKm: Double = 5

    @State private var isSaving = false
    @State private var errorMessage: String?

    var body: some View {
        Form {
            // Hangout Type Section
            Section("What kind of hangout?") {
                ForEach(HangoutType.allCases, id: \.self) { type in
                    Button {
                        hangoutType = type
                    } label: {
                        HStack {
                            Image(systemName: type.iconName)
                                .frame(width: 28)
                                .foregroundStyle(hangoutType == type ? .blue : .secondary)

                            VStack(alignment: .leading, spacing: 2) {
                                Text(type.displayName)
                                    .font(.subheadline.bold())
                                    .foregroundStyle(.primary)
                                Text(type.description)
                                    .font(.caption)
                                    .foregroundStyle(.secondary)
                            }

                            Spacer()

                            if hangoutType == type {
                                Image(systemName: "checkmark.circle.fill")
                                    .foregroundStyle(.blue)
                            }
                        }
                    }
                    .buttonStyle(.plain)
                }
            }

            // Details Section
            Section("Details (Optional)") {
                TextField("Title", text: $title)

                TextField("Description", text: $description, axis: .vertical)
                    .lineLimit(2...4)
            }

            // Time Section
            Section("When are you available?") {
                DatePicker("Start", selection: $startTime, in: Date()...)
                DatePicker("End", selection: $endTime, in: startTime...)
            }

            // Location Section
            Section("Location (Optional)") {
                TextField("Location name", text: $location)

                VStack(alignment: .leading, spacing: 8) {
                    Text("Search radius: \(Int(radiusKm)) km")
                        .font(.subheadline)

                    Slider(value: $radiusKm, in: 1...50, step: 1)
                }
            }

            // Error Section
            if let errorMessage = errorMessage {
                Section {
                    Text(errorMessage)
                        .foregroundStyle(.red)
                }
            }
        }
        .navigationTitle("Post Availability")
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .cancellationAction) {
                Button("Cancel") {
                    dismiss()
                }
            }

            ToolbarItem(placement: .confirmationAction) {
                Button("Post") {
                    Task {
                        await createAvailability()
                    }
                }
                .disabled(isSaving || !isValid)
            }
        }
        .disabled(isSaving)
        .overlay {
            if isSaving {
                ProgressView()
                    .scaleEffect(1.5)
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
                    .background(.ultraThinMaterial)
            }
        }
    }

    private var isValid: Bool {
        endTime > startTime
    }

    private func createAvailability() async {
        isSaving = true
        errorMessage = nil

        let request = CreateAvailabilityRequest(
            hangoutType: hangoutType,
            title: title.isEmpty ? nil : title,
            description: description.isEmpty ? nil : description,
            startTime: startTime,
            endTime: endTime,
            location: location.isEmpty ? nil : location,
            locationLat: nil, // Would come from location services
            locationLng: nil,
            radiusKm: radiusKm
        )

        do {
            _ = try await profileService.createAvailability(request)
            dismiss()
        } catch {
            errorMessage = error.localizedDescription
        }

        isSaving = false
    }
}

#Preview {
    NavigationStack {
        CreateAvailabilitySheet()
            .environment(ProfileService.shared)
    }
}
