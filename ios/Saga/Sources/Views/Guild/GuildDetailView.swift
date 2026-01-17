import SwiftUI

/// Detail view for a guild showing people and their timers
struct GuildDetailView: View {
    let guild: Guild

    @Environment(GuildService.self) private var guildService
    @State private var showingAddPerson = false
    @State private var searchText = ""
    @State private var errorMessage: String?
    @State private var isShowingError = false

    private var filteredPeople: [Person] {
        guard let data = guildService.currentGuild else { return [] }
        if searchText.isEmpty {
            return data.people
        }
        return data.people.filter {
            $0.name.localizedCaseInsensitiveContains(searchText) ||
            ($0.nickname?.localizedCaseInsensitiveContains(searchText) ?? false)
        }
    }

    var body: some View {
        List {
            // Guild info section
            Section {
                HStack(spacing: 16) {
                    ZStack {
                        Circle()
                            .fill(guild.displayColor.gradient)
                            .frame(width: 60, height: 60)
                        Image(systemName: guild.iconName)
                            .font(.title)
                            .foregroundStyle(.white)
                    }

                    VStack(alignment: .leading, spacing: 4) {
                        Text(guild.name)
                            .font(.title2.bold())
                        if let description = guild.description {
                            Text(description)
                                .font(.subheadline)
                                .foregroundStyle(.secondary)
                        }
                        if let data = guildService.currentGuild {
                            HStack(spacing: 12) {
                                Label("\(data.members.count)", systemImage: "person.2.fill")
                                Label("\(data.people.count)", systemImage: "person.fill")
                            }
                            .font(.caption)
                            .foregroundStyle(.secondary)
                        }
                    }
                }
                .padding(.vertical, 8)
            }

            // Connection status
            if guildService.isConnected {
                HStack {
                    Circle()
                        .fill(.green)
                        .frame(width: 8, height: 8)
                    Text("Live updates enabled")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }
                .listRowBackground(Color.clear)
            }

            // Features section
            Section("Features") {
                NavigationLink(value: GuildDestination.events) {
                    Label("Events", systemImage: "calendar")
                }

                NavigationLink(value: GuildDestination.adventures) {
                    Label("Adventures", systemImage: "figure.hiking")
                }

                NavigationLink(value: GuildDestination.pools) {
                    Label("Matching Pools", systemImage: "person.2.circle")
                }

                NavigationLink(value: GuildDestination.votes) {
                    Label("Votes", systemImage: "chart.bar.xaxis")
                }
            }

            // People section
            if filteredPeople.isEmpty && !guildService.isLoading {
                if searchText.isEmpty {
                    ContentUnavailableView(
                        "No People Yet",
                        systemImage: "person.fill.badge.plus",
                        description: Text("Add people to track your relationships")
                    )
                    .listRowBackground(Color.clear)
                } else {
                    ContentUnavailableView.search(text: searchText)
                        .listRowBackground(Color.clear)
                }
            } else {
                Section("People") {
                    ForEach(filteredPeople) { person in
                        NavigationLink(value: GuildDestination.person(person: person)) {
                            PersonRow(person: person)
                        }
                    }
                    .onDelete(perform: deletePeople)
                }
            }
        }
        .searchable(text: $searchText, prompt: "Search people")
        .navigationTitle(guild.name)
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .primaryAction) {
                Button {
                    showingAddPerson = true
                } label: {
                    Image(systemName: "person.badge.plus")
                }
            }
        }
        .refreshable {
            try? await guildService.selectGuild(id: guild.id)
        }
        .task {
            if guildService.currentGuild?.guild.id != guild.id {
                try? await guildService.selectGuild(id: guild.id)
            }
        }
        .sheet(isPresented: $showingAddPerson) {
            CreatePersonSheet()
        }
        .alert("Error", isPresented: $isShowingError) {
            Button("OK", role: .cancel) { }
        } message: {
            Text(errorMessage ?? "An error occurred")
        }
    }

    private func deletePeople(at offsets: IndexSet) {
        Task {
            for index in offsets {
                let person = filteredPeople[index]
                do {
                    try await guildService.deletePerson(id: person.id)
                } catch {
                    errorMessage = error.localizedDescription
                    isShowingError = true
                }
            }
        }
    }
}

// MARK: - Person Row

struct PersonRow: View {
    let person: Person

    var body: some View {
        HStack(spacing: 12) {
            // Avatar
            ZStack {
                Circle()
                    .fill(.gray.opacity(0.2))
                    .frame(width: 44, height: 44)
                Text(person.initials)
                    .font(.headline)
                    .foregroundStyle(.secondary)
            }

            // Info
            VStack(alignment: .leading, spacing: 2) {
                Text(person.displayName)
                    .font(.headline)

                if let days = person.daysUntilBirthday {
                    if days == 0 {
                        Text("Birthday today!")
                            .font(.caption)
                            .foregroundStyle(.orange)
                    } else if days <= 7 {
                        Text("Birthday in \(days) days")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                }
            }

            Spacer()

            // Timer status indicator could go here
        }
        .padding(.vertical, 4)
    }
}

// MARK: - Create Person Sheet

struct CreatePersonSheet: View {
    @Environment(\.dismiss) private var dismiss
    @Environment(GuildService.self) private var guildService

    @State private var name = ""
    @State private var nickname = ""
    @State private var birthday = Date()
    @State private var hasBirthday = false
    @State private var notes = ""
    @State private var isCreating = false
    @State private var errorMessage: String?

    var body: some View {
        NavigationStack {
            Form {
                Section {
                    TextField("Name", text: $name)
                    TextField("Nickname (optional)", text: $nickname)
                }

                Section {
                    Toggle("Birthday", isOn: $hasBirthday)
                    if hasBirthday {
                        DatePicker("Birthday", selection: $birthday, displayedComponents: .date)
                    }
                }

                Section {
                    TextField("Notes (optional)", text: $notes, axis: .vertical)
                        .lineLimit(3...6)
                }

                if let error = errorMessage {
                    Section {
                        Text(error)
                            .foregroundStyle(.red)
                    }
                }
            }
            .navigationTitle("Add Person")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") {
                        dismiss()
                    }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Add") {
                        createPerson()
                    }
                    .disabled(name.isEmpty || isCreating)
                }
            }
        }
    }

    private func createPerson() {
        guard !name.isEmpty else { return }

        isCreating = true
        errorMessage = nil

        let formatter = DateFormatter()
        formatter.dateFormat = "yyyy-MM-dd"
        let birthdayString = hasBirthday ? formatter.string(from: birthday) : nil

        Task {
            do {
                _ = try await guildService.createPerson(
                    name: name,
                    nickname: nickname.isEmpty ? nil : nickname,
                    birthday: birthdayString,
                    notes: notes.isEmpty ? nil : notes
                )
                dismiss()
            } catch {
                errorMessage = error.localizedDescription
            }
            isCreating = false
        }
    }
}

#Preview {
    NavigationStack {
        GuildDetailView(guild: Guild(
            id: "guild:1",
            name: "Close Friends",
            description: "My closest friends",
            icon: "person.3.fill",
            color: "#6B46C1",
            createdOn: Date(),
            updatedOn: Date()
        ))
        .environment(GuildService.shared)
    }
}
