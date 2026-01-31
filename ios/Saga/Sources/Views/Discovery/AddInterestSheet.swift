import SwiftUI

/// Sheet for adding a new interest
struct AddInterestSheet: View {
    @Environment(DiscoveryService.self) private var discoveryService
    @Environment(\.dismiss) private var dismiss

    @State private var searchText = ""
    @State private var selectedCategory: InterestCategory?
    @State private var selectedInterest: Interest?
    @State private var skillLevel: SkillLevel = .beginner
    @State private var intent: InterestIntent = .both
    @State private var notes = ""

    @State private var categoryInterests: [Interest] = []
    @State private var searchResults: [Interest] = []
    @State private var isSearching = false
    @State private var isSaving = false
    @State private var errorMessage: String?

    var body: some View {
        Form {
            // Search Section
            Section {
                TextField("Search interests...", text: $searchText)
                    .textInputAutocapitalization(.never)
                    .autocorrectionDisabled()
                    .onChange(of: searchText) { _, newValue in
                        Task {
                            await searchInterests(query: newValue)
                        }
                    }
            }

            // Search Results
            if !searchText.isEmpty && !searchResults.isEmpty {
                Section("Search Results") {
                    ForEach(searchResults) { interest in
                        Button {
                            selectedInterest = interest
                            searchText = ""
                            searchResults = []
                        } label: {
                            HStack {
                                Text(interest.name)
                                    .foregroundStyle(.primary)
                                Spacer()
                                if selectedInterest?.id == interest.id {
                                    Image(systemName: "checkmark")
                                        .foregroundStyle(.blue)
                                }
                            }
                        }
                    }
                }
            }

            // Categories Section
            if searchText.isEmpty && selectedInterest == nil {
                Section("Browse Categories") {
                    if discoveryService.interestCategories.isEmpty {
                        HStack {
                            Spacer()
                            ProgressView()
                            Spacer()
                        }
                    } else {
                        ForEach(discoveryService.interestCategories) { category in
                            Button {
                                selectedCategory = category
                                Task {
                                    await loadCategoryInterests(categoryId: category.id)
                                }
                            } label: {
                                HStack {
                                    if let iconName = category.iconName {
                                        Image(systemName: iconName)
                                            .frame(width: 24)
                                    }
                                    Text(category.name)
                                        .foregroundStyle(.primary)
                                    Spacer()
                                    Image(systemName: "chevron.right")
                                        .foregroundStyle(.secondary)
                                }
                            }
                        }
                    }
                }
            }

            // Category Interests
            if let category = selectedCategory, searchText.isEmpty && selectedInterest == nil {
                Section(category.name) {
                    if categoryInterests.isEmpty {
                        HStack {
                            Spacer()
                            ProgressView()
                            Spacer()
                        }
                    } else {
                        ForEach(categoryInterests) { interest in
                            Button {
                                selectedInterest = interest
                            } label: {
                                HStack {
                                    Text(interest.name)
                                        .foregroundStyle(.primary)
                                    Spacer()
                                    if selectedInterest?.id == interest.id {
                                        Image(systemName: "checkmark")
                                            .foregroundStyle(.blue)
                                    }
                                }
                            }
                        }
                    }

                    Button("Back to Categories") {
                        selectedCategory = nil
                        categoryInterests = []
                    }
                    .font(.caption)
                }
            }

            // Selected Interest Details
            if let interest = selectedInterest {
                Section("Selected Interest") {
                    HStack {
                        Text(interest.name)
                            .font(.headline)
                        Spacer()
                        Button("Change") {
                            selectedInterest = nil
                        }
                        .font(.caption)
                    }

                    if let description = interest.description {
                        Text(description)
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                }

                // Skill Level
                Section("Your Skill Level") {
                    ForEach(SkillLevel.allCases, id: \.self) { level in
                        Button {
                            skillLevel = level
                        } label: {
                            HStack {
                                Image(systemName: level.iconName)
                                    .frame(width: 24)
                                    .foregroundStyle(skillLevel == level ? .blue : .secondary)

                                VStack(alignment: .leading, spacing: 2) {
                                    Text(level.displayName)
                                        .foregroundStyle(.primary)
                                    Text(level.description)
                                        .font(.caption)
                                        .foregroundStyle(.secondary)
                                }

                                Spacer()

                                if skillLevel == level {
                                    Image(systemName: "checkmark")
                                        .foregroundStyle(.blue)
                                }
                            }
                        }
                    }
                }

                // Intent
                Section("What do you want to do?") {
                    ForEach(InterestIntent.allCases, id: \.self) { intentOption in
                        Button {
                            intent = intentOption
                        } label: {
                            HStack {
                                Image(systemName: intentOption.iconName)
                                    .frame(width: 24)
                                    .foregroundStyle(intent == intentOption ? .blue : .secondary)

                                VStack(alignment: .leading, spacing: 2) {
                                    Text(intentOption.displayName)
                                        .foregroundStyle(.primary)
                                    Text(intentOption.description)
                                        .font(.caption)
                                        .foregroundStyle(.secondary)
                                }

                                Spacer()

                                if intent == intentOption {
                                    Image(systemName: "checkmark")
                                        .foregroundStyle(.blue)
                                }
                            }
                        }
                    }
                }

                // Notes
                Section("Notes (Optional)") {
                    TextField("Any additional details...", text: $notes, axis: .vertical)
                        .lineLimit(2...4)
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
        .navigationTitle("Add Interest")
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .cancellationAction) {
                Button("Cancel") {
                    dismiss()
                }
            }

            ToolbarItem(placement: .confirmationAction) {
                Button("Add") {
                    Task {
                        await addInterest()
                    }
                }
                .disabled(selectedInterest == nil || isSaving)
            }
        }
        .disabled(isSaving)
        .task {
            await discoveryService.loadInterestCategories()
        }
    }

    private func searchInterests(query: String) async {
        guard query.count >= 2 else {
            searchResults = []
            return
        }

        isSearching = true
        do {
            searchResults = try await discoveryService.searchInterests(query: query)
        } catch {
            searchResults = []
        }
        isSearching = false
    }

    private func loadCategoryInterests(categoryId: String) async {
        do {
            categoryInterests = try await discoveryService.getInterests(categoryId: categoryId)
        } catch {
            categoryInterests = []
        }
    }

    private func addInterest() async {
        guard let interest = selectedInterest else { return }

        isSaving = true
        errorMessage = nil

        let request = AddInterestRequest(
            interestId: interest.id,
            skillLevel: skillLevel,
            intent: intent,
            notes: notes.isEmpty ? nil : notes
        )

        do {
            _ = try await discoveryService.addInterest(request)
            dismiss()
        } catch {
            errorMessage = error.localizedDescription
        }

        isSaving = false
    }
}

#Preview {
    NavigationStack {
        AddInterestSheet()
            .environment(DiscoveryService.shared)
    }
}
