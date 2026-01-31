import SwiftUI

/// View for casting a ballot on a vote
struct CastBallotView: View {
    let vote: Vote

    @Environment(\.dismiss) private var dismiss

    // Single choice / Multi-select
    @State private var selectedOptions: Set<String> = []

    // Ranked choice
    @State private var rankedOptions: [VoteOption] = []
    @State private var unrankedOptions: [VoteOption] = []

    // Approval
    @State private var approvals: [String: Bool] = [:]

    @State private var isAbstaining = false
    @State private var isSubmitting = false
    @State private var error: Error?

    private let apiClient = APIClient.shared

    var canSubmit: Bool {
        if isAbstaining { return true }

        switch vote.voteType {
        case .fptp:
            return selectedOptions.count == 1
        case .multiSelect:
            let max = vote.settings.maxSelections ?? vote.options.count
            return !selectedOptions.isEmpty && selectedOptions.count <= max
        case .ranked:
            if vote.settings.requireAllRanked {
                return rankedOptions.count == vote.options.count
            }
            return !rankedOptions.isEmpty
        case .approval:
            return !approvals.isEmpty
        }
    }

    var body: some View {
        Form {
            // Vote Info
            Section {
                VStack(alignment: .leading, spacing: 8) {
                    Text(vote.title)
                        .font(.headline)
                    if let description = vote.description {
                        Text(description)
                            .font(.subheadline)
                            .foregroundStyle(.secondary)
                    }
                    Label(vote.voteType.description, systemImage: vote.voteType.iconName)
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }
            }

            // Voting UI based on type
            switch vote.voteType {
            case .fptp:
                singleChoiceSection
            case .multiSelect:
                multiSelectSection
            case .ranked:
                rankedSection
            case .approval:
                approvalSection
            }

            // Abstain option
            if vote.settings.allowAbstain {
                Section {
                    Toggle("Abstain from this vote", isOn: $isAbstaining)
                }
            }
        }
        .navigationTitle("Cast Vote")
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .cancellationAction) {
                Button("Cancel") {
                    dismiss()
                }
            }
            ToolbarItem(placement: .confirmationAction) {
                Button("Submit") {
                    Task { await submitBallot() }
                }
                .disabled(!canSubmit || isSubmitting)
            }
        }
        .alert("Error", isPresented: .constant(error != nil)) {
            Button("OK") { error = nil }
        } message: {
            if let error = error {
                Text(error.localizedDescription)
            }
        }
        .onAppear {
            setupForVoteType()
        }
    }

    // MARK: - Single Choice (FPTP)

    @ViewBuilder
    private var singleChoiceSection: some View {
        Section("Choose One") {
            ForEach(vote.options) { option in
                Button {
                    selectedOptions = [option.id]
                } label: {
                    HStack {
                        VStack(alignment: .leading) {
                            Text(option.text)
                                .foregroundStyle(.primary)
                            if let description = option.description {
                                Text(description)
                                    .font(.caption)
                                    .foregroundStyle(.secondary)
                            }
                        }
                        Spacer()
                        if selectedOptions.contains(option.id) {
                            Image(systemName: "checkmark.circle.fill")
                                .foregroundStyle(.green)
                        } else {
                            Image(systemName: "circle")
                                .foregroundStyle(.secondary)
                        }
                    }
                }
                .disabled(isAbstaining)
            }
        }
    }

    // MARK: - Multi-Select

    @ViewBuilder
    private var multiSelectSection: some View {
        Section {
            ForEach(vote.options) { option in
                Button {
                    if selectedOptions.contains(option.id) {
                        selectedOptions.remove(option.id)
                    } else {
                        selectedOptions.insert(option.id)
                    }
                } label: {
                    HStack {
                        VStack(alignment: .leading) {
                            Text(option.text)
                                .foregroundStyle(.primary)
                            if let description = option.description {
                                Text(description)
                                    .font(.caption)
                                    .foregroundStyle(.secondary)
                            }
                        }
                        Spacer()
                        if selectedOptions.contains(option.id) {
                            Image(systemName: "checkmark.square.fill")
                                .foregroundStyle(.green)
                        } else {
                            Image(systemName: "square")
                                .foregroundStyle(.secondary)
                        }
                    }
                }
                .disabled(isAbstaining)
            }
        } header: {
            Text("Select Multiple")
        } footer: {
            if let max = vote.settings.maxSelections {
                Text("You can select up to \(max) options")
            }
        }
    }

    // MARK: - Ranked Choice

    @ViewBuilder
    private var rankedSection: some View {
        if !rankedOptions.isEmpty {
            Section("Your Rankings") {
                ForEach(Array(rankedOptions.enumerated()), id: \.element.id) { index, option in
                    HStack {
                        Text("#\(index + 1)")
                            .font(.caption.bold())
                            .padding(8)
                            .background(Color.blue.opacity(0.2))
                            .clipShape(Circle())

                        Text(option.text)

                        Spacer()

                        Button {
                            rankedOptions.remove(at: index)
                            unrankedOptions.append(option)
                        } label: {
                            Image(systemName: "xmark.circle.fill")
                                .foregroundStyle(.red)
                        }
                    }
                }
                .onMove { from, to in
                    rankedOptions.move(fromOffsets: from, toOffset: to)
                }
            }
        }

        if !unrankedOptions.isEmpty {
            Section("Tap to Rank") {
                ForEach(unrankedOptions) { option in
                    Button {
                        if let index = unrankedOptions.firstIndex(where: { $0.id == option.id }) {
                            unrankedOptions.remove(at: index)
                            rankedOptions.append(option)
                        }
                    } label: {
                        HStack {
                            Text(option.text)
                                .foregroundStyle(.primary)
                            Spacer()
                            Image(systemName: "plus.circle")
                                .foregroundStyle(.blue)
                        }
                    }
                    .disabled(isAbstaining)
                }
            }
        }
    }

    // MARK: - Approval

    @ViewBuilder
    private var approvalSection: some View {
        Section("Approve or Disapprove Each") {
            ForEach(vote.options) { option in
                HStack {
                    VStack(alignment: .leading) {
                        Text(option.text)
                        if let description = option.description {
                            Text(description)
                                .font(.caption)
                                .foregroundStyle(.secondary)
                        }
                    }

                    Spacer()

                    HStack(spacing: 16) {
                        Button {
                            approvals[option.id] = true
                        } label: {
                            Image(systemName: approvals[option.id] == true ? "hand.thumbsup.fill" : "hand.thumbsup")
                                .foregroundStyle(approvals[option.id] == true ? .green : .secondary)
                        }

                        Button {
                            approvals[option.id] = false
                        } label: {
                            Image(systemName: approvals[option.id] == false ? "hand.thumbsdown.fill" : "hand.thumbsdown")
                                .foregroundStyle(approvals[option.id] == false ? .red : .secondary)
                        }
                    }
                    .buttonStyle(.plain)
                    .disabled(isAbstaining)
                }
            }
        }
    }

    // MARK: - Setup

    private func setupForVoteType() {
        switch vote.voteType {
        case .ranked:
            unrankedOptions = vote.options.sorted { $0.sortOrder < $1.sortOrder }
        case .approval:
            // Initialize all as undecided
            break
        default:
            break
        }
    }

    // MARK: - Submit

    private func submitBallot() async {
        isSubmitting = true

        var selections: [BallotSelection] = []

        if !isAbstaining {
            switch vote.voteType {
            case .fptp, .multiSelect:
                selections = selectedOptions.map { BallotSelection(optionId: $0, rank: nil, approved: nil) }

            case .ranked:
                selections = rankedOptions.enumerated().map { index, option in
                    BallotSelection(optionId: option.id, rank: index + 1, approved: nil)
                }

            case .approval:
                selections = approvals.map { optionId, approved in
                    BallotSelection(optionId: optionId, rank: nil, approved: approved)
                }
            }
        }

        do {
            _ = try await apiClient.castBallot(voteId: vote.id, selections: selections, abstain: isAbstaining)
            dismiss()
        } catch {
            self.error = error
        }

        isSubmitting = false
    }
}

#Preview {
    let sampleVote = Vote(
        id: "1",
        guildId: "g1",
        creatorId: "u1",
        title: "Best Pizza Topping",
        description: "Vote for your favorite",
        voteType: .fptp,
        options: [
            VoteOption(id: "o1", voteId: "1", text: "Pepperoni", description: nil, sortOrder: 0),
            VoteOption(id: "o2", voteId: "1", text: "Mushrooms", description: nil, sortOrder: 1),
            VoteOption(id: "o3", voteId: "1", text: "Pineapple", description: "Controversial!", sortOrder: 2)
        ],
        settings: VoteSettings(
            allowAbstain: true,
            showResultsBeforeEnd: false,
            maxSelections: nil,
            anonymousVoting: false,
            requireAllRanked: false
        ),
        status: .active,
        startTime: Date(),
        endTime: Date().addingTimeInterval(86400),
        totalVoters: 5,
        createdOn: Date()
    )

    return NavigationStack {
        CastBallotView(vote: sampleVote)
    }
}
