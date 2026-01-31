import SwiftUI

/// List of user's guilds
struct GuildListView: View {
    @Environment(GuildService.self) private var guildService
    @State private var showingCreateSheet = false
    @State private var errorMessage: String?
    @State private var isShowingError = false
    @State private var needsRefresh = false

    var body: some View {
        List {
            if guildService.guilds.isEmpty && !guildService.isLoading {
                ContentUnavailableView(
                    "No Guilds Yet",
                    systemImage: "person.3.fill",
                    description: Text("Create a guild to start tracking your relationships")
                )
                .accessibilityIdentifier("guild_empty_state")
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
        // Force list to re-render when guild count changes
        .id(guildService.guilds.count)
        .accessibilityIdentifier("guild_list")
        .navigationTitle("Guilds")
        .toolbar {
            ToolbarItem(placement: .primaryAction) {
                Button {
                    showingCreateSheet = true
                } label: {
                    Image(systemName: "plus")
                }
                .accessibilityIdentifier("guild_create_button")
            }
        }
        .refreshable {
            do {
                try await guildService.fetchGuilds()
            } catch {
                errorMessage = error.localizedDescription
                isShowingError = true
            }
        }
        .task {
            if guildService.guilds.isEmpty {
                do {
                    try await guildService.fetchGuilds()
                } catch {
                    #if DEBUG
                    print("GuildListView: Failed to fetch guilds - \(error)")
                    #endif
                    errorMessage = error.localizedDescription
                    isShowingError = true
                }
            }
        }
        .sheet(isPresented: $showingCreateSheet, onDismiss: {
            needsRefresh = true
        }) {
            CreateGuildSheet()
                .environment(guildService)
        }
        .onChange(of: needsRefresh) { _, newValue in
            if newValue {
                needsRefresh = false
                Task {
                    do {
                        try await guildService.fetchGuilds()
                    } catch {
                        errorMessage = error.localizedDescription
                        isShowingError = true
                    }
                }
            }
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
                        .accessibilityIdentifier("guild_name_field")
                    TextField("Description (optional)", text: $description, axis: .vertical)
                        .lineLimit(2...4)
                        .accessibilityIdentifier("guild_description_field")
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
                    .accessibilityIdentifier("guild_create_cancel")
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Create") {
                        createGuild()
                    }
                    .disabled(name.isEmpty || isCreating)
                    .accessibilityIdentifier("guild_create_confirm")
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
                // Fetch updated list to ensure UI is in sync
                try? await guildService.fetchGuilds()
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
    case events(guildId: String)
    case adventures(guildId: String)
    case pools(guildId: String)
    case votes(guildId: String)
}

#Preview {
    NavigationStack {
        GuildListView()
            .environment(GuildService.shared)
    }
}
