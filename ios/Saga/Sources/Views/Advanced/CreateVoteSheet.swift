import SwiftUI

/// Sheet for creating a new vote
struct CreateVoteSheet: View {
    let guildId: String

    @Environment(\.dismiss) private var dismiss

    @State private var title = ""
    @State private var description = ""
    @State private var voteType: VoteType = .fptp
    @State private var options: [OptionInput] = [OptionInput(), OptionInput()]
    @State private var startTime = Date()
    @State private var endTime: Date?
    @State private var hasEndTime = true

    // Settings
    @State private var allowAbstain = true
    @State private var showResultsBeforeEnd = false
    @State private var anonymousVoting = false
    @State private var maxSelections: Int?
    @State private var requireAllRanked = false

    @State private var isCreating = false
    @State private var error: Error?

    private let apiClient = APIClient.shared

    var canCreate: Bool {
        !title.isEmpty && options.filter { !$0.text.isEmpty }.count >= 2
    }

    var body: some View {
        Form {
            // Basic Info
            Section("Details") {
                TextField("Question/Title", text: $title)
                TextField("Description (optional)", text: $description, axis: .vertical)
                    .lineLimit(3...6)
            }

            // Vote Type
            Section {
                Picker("Vote Type", selection: $voteType) {
                    ForEach(VoteType.allCases, id: \.self) { type in
                        Label(type.displayName, systemImage: type.iconName)
                            .tag(type)
                    }
                }
            } header: {
                Text("Type")
            } footer: {
                Text(voteType.description)
            }

            // Options
            Section {
                ForEach($options) { $option in
                    HStack {
                        TextField("Option", text: $option.text)
                        if options.count > 2 {
                            Button {
                                options.removeAll { $0.id == option.id }
                            } label: {
                                Image(systemName: "minus.circle.fill")
                                    .foregroundStyle(.red)
                            }
                            .buttonStyle(.plain)
                        }
                    }
                }

                Button {
                    options.append(OptionInput())
                } label: {
                    Label("Add Option", systemImage: "plus")
                }
            } header: {
                Text("Options")
            } footer: {
                Text("At least 2 options required")
            }

            // Schedule
            Section("Schedule") {
                DatePicker("Start", selection: $startTime, in: Date()...)

                Toggle("Set End Time", isOn: $hasEndTime)

                if hasEndTime {
                    DatePicker("End", selection: Binding(
                        get: { endTime ?? startTime.addingTimeInterval(86400) },
                        set: { endTime = $0 }
                    ), in: startTime...)
                }
            }

            // Settings
            Section("Settings") {
                Toggle("Allow Abstaining", isOn: $allowAbstain)
                Toggle("Show Results Before End", isOn: $showResultsBeforeEnd)
                Toggle("Anonymous Voting", isOn: $anonymousVoting)

                if voteType == .multiSelect {
                    Stepper(
                        "Max Selections: \(maxSelections ?? options.count)",
                        value: Binding(
                            get: { maxSelections ?? options.count },
                            set: { maxSelections = $0 }
                        ),
                        in: 1...options.count
                    )
                }

                if voteType == .ranked {
                    Toggle("Require All Options Ranked", isOn: $requireAllRanked)
                }
            }
        }
        .navigationTitle("New Vote")
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .cancellationAction) {
                Button("Cancel") {
                    dismiss()
                }
            }
            ToolbarItem(placement: .confirmationAction) {
                Button("Create") {
                    Task { await createVote() }
                }
                .disabled(!canCreate || isCreating)
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

    private func createVote() async {
        isCreating = true

        let validOptions = options.filter { !$0.text.isEmpty }

        let settings = VoteSettings(
            allowAbstain: allowAbstain,
            showResultsBeforeEnd: showResultsBeforeEnd,
            maxSelections: voteType == .multiSelect ? maxSelections : nil,
            anonymousVoting: anonymousVoting,
            requireAllRanked: voteType == .ranked ? requireAllRanked : false
        )

        let request = CreateVoteRequest(
            guildId: guildId,
            title: title,
            description: description.isEmpty ? nil : description,
            voteType: voteType,
            options: validOptions.map { CreateVoteOptionRequest(text: $0.text, description: nil) },
            settings: settings,
            startTime: startTime,
            endTime: hasEndTime ? endTime : nil
        )

        do {
            _ = try await apiClient.createVote(request)
            dismiss()
        } catch {
            self.error = error
        }

        isCreating = false
    }
}

// MARK: - Option Input

struct OptionInput: Identifiable {
    let id = UUID()
    var text = ""
}

#Preview {
    NavigationStack {
        CreateVoteSheet(guildId: "test")
    }
}
