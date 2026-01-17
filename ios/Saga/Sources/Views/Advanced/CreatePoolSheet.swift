import SwiftUI

/// Sheet for creating a new matching pool
struct CreatePoolSheet: View {
    let guildId: String

    @Environment(\.dismiss) private var dismiss

    @State private var name = ""
    @State private var description = ""
    @State private var matchingFrequency: MatchingFrequency = .weekly
    @State private var matchSize = 2
    @State private var isCreating = false
    @State private var error: Error?

    private let apiClient = APIClient.shared

    var body: some View {
        Form {
            // Basic Info
            Section("Details") {
                TextField("Pool Name", text: $name)
                TextField("Description (optional)", text: $description, axis: .vertical)
                    .lineLimit(3...6)
            }

            // Matching Settings
            Section {
                Picker("Frequency", selection: $matchingFrequency) {
                    ForEach(MatchingFrequency.allCases, id: \.self) { freq in
                        Text(freq.displayName).tag(freq)
                    }
                }

                Stepper("Match Size: \(matchSize)", value: $matchSize, in: 2...8)
            } header: {
                Text("Matching Settings")
            } footer: {
                Text("Match size is the number of people matched together each cycle.")
            }
        }
        .navigationTitle("New Pool")
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .cancellationAction) {
                Button("Cancel") {
                    dismiss()
                }
            }
            ToolbarItem(placement: .confirmationAction) {
                Button("Create") {
                    Task { await createPool() }
                }
                .disabled(name.isEmpty || isCreating)
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

    private func createPool() async {
        isCreating = true

        let request = CreatePoolRequest(
            guildId: guildId,
            name: name,
            description: description.isEmpty ? nil : description,
            matchingFrequency: matchingFrequency,
            matchSize: matchSize
        )

        do {
            _ = try await apiClient.createPool(request)
            dismiss()
        } catch {
            self.error = error
        }

        isCreating = false
    }
}

#Preview {
    NavigationStack {
        CreatePoolSheet(guildId: "test")
    }
}
