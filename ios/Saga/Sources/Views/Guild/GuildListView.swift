import SwiftUI

/// List of user's guilds
struct GuildListView: View {
    @Environment(GuildService.self) private var guildService
    @State private var showingCreateSheet = false
    @State private var errorMessage: String?
    @State private var isShowingError = false

    var body: some View {
        List {
            if guildService.guilds.isEmpty && !guildService.isLoading {
                ContentUnavailableView(
                    "No Guilds Yet",
                    systemImage: "person.3.fill",
                    description: Text("Create a guild to start tracking your relationships")
                )
                .listRowBackground(Color.clear)
            } else {
                ForEach(guildService.guilds) { guild in
                    NavigationLink(value: GuildDestination.detail(guild: guild)) {
                        GuildRow(guild: guild)
                    }
                    .accessibilityIdentifier("guild_row_\(guild.id)")
                }
                .onDelete(perform: deleteGuilds)
            }
        }
        .accessibilityIdentifier("guild_list")
        .navigationTitle("Guilds")
        .toolbar {
            ToolbarItem(placement: .primaryAction) {
                Button {
                    showingCreateSheet = true
                } label: {
                    Image(systemName: "plus")
                }
                .accessibilityIdentifier("create_guild_button")
            }
        }
        .refreshable {
            try? await guildService.fetchGuilds()
        }
        .task {
            if guildService.guilds.isEmpty {
                try? await guildService.fetchGuilds()
            }
        }
        .sheet(isPresented: $showingCreateSheet) {
            CreateGuildSheet()
        }
        .alert("Error", isPresented: $isShowingError) {
            Button("OK", role: .cancel) { }
        } message: {
            Text(errorMessage ?? "An error occurred")
        }
    }

    private func deleteGuilds(at offsets: IndexSet) {
        Task {
            for index in offsets {
                let guild = guildService.guilds[index]
                do {
                    try await guildService.deleteGuild(id: guild.id)
                } catch {
                    errorMessage = error.localizedDescription
                    isShowingError = true
                }
            }
        }
    }
}

// MARK: - Guild Row

struct GuildRow: View {
    let guild: Guild

    var body: some View {
        HStack(spacing: 12) {
            // Guild icon
            ZStack {
                Circle()
                    .fill(guild.displayColor.gradient)
                    .frame(width: 44, height: 44)
                Image(systemName: guild.iconName)
                    .font(.title3)
                    .foregroundStyle(.white)
            }

            // Guild info
            VStack(alignment: .leading, spacing: 2) {
                Text(guild.name)
                    .font(.headline)
                if let description = guild.description {
                    Text(description)
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                        .lineLimit(1)
                }
            }
        }
        .padding(.vertical, 4)
    }
}

// MARK: - Create Guild Sheet

struct CreateGuildSheet: View {
    @Environment(\.dismiss) private var dismiss
    @Environment(GuildService.self) private var guildService

    @State private var name = ""
    @State private var description = ""
    @State private var selectedIcon = "person.3.fill"
    @State private var selectedColor = Color.blue
    @State private var isCreating = false
    @State private var errorMessage: String?

    private let icons = [
        "person.3.fill", "figure.2.arms.open", "heart.fill",
        "house.fill", "briefcase.fill", "graduationcap.fill",
        "sportscourt.fill", "music.note", "gamecontroller.fill"
    ]

    var body: some View {
        NavigationStack {
            Form {
                Section {
                    TextField("Guild name", text: $name)
                    TextField("Description (optional)", text: $description, axis: .vertical)
                        .lineLimit(2...4)
                }

                Section("Icon") {
                    LazyVGrid(columns: Array(repeating: GridItem(.flexible()), count: 5), spacing: 16) {
                        ForEach(icons, id: \.self) { icon in
                            Button {
                                selectedIcon = icon
                            } label: {
                                ZStack {
                                    Circle()
                                        .fill(icon == selectedIcon ? selectedColor : Color.gray.opacity(0.2))
                                        .frame(width: 44, height: 44)
                                    Image(systemName: icon)
                                        .foregroundStyle(icon == selectedIcon ? .white : .primary)
                                }
                            }
                            .buttonStyle(.plain)
                        }
                    }
                    .padding(.vertical, 8)
                }

                Section("Color") {
                    ColorPicker("Guild color", selection: $selectedColor, supportsOpacity: false)
                }

                if let error = errorMessage {
                    Section {
                        Text(error)
                            .foregroundStyle(.red)
                    }
                }
            }
            .navigationTitle("New Guild")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") {
                        dismiss()
                    }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Create") {
                        createGuild()
                    }
                    .disabled(name.isEmpty || isCreating)
                }
            }
        }
    }

    private func createGuild() {
        guard !name.isEmpty else { return }

        isCreating = true
        errorMessage = nil

        Task {
            do {
                _ = try await guildService.createGuild(
                    name: name,
                    description: description.isEmpty ? nil : description,
                    icon: selectedIcon,
                    color: selectedColor.hexString
                )
                dismiss()
            } catch {
                errorMessage = error.localizedDescription
            }
            isCreating = false
        }
    }
}

// MARK: - Navigation Destination

enum GuildDestination: Hashable {
    case detail(guild: Guild)
    case person(person: Person)
    case createPerson
    case settings
    case events
    case adventures
    case pools
    case votes
}

#Preview {
    NavigationStack {
        GuildListView()
            .environment(GuildService.shared)
    }
}
