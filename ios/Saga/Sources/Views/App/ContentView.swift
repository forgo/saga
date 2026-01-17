import SwiftUI

/// Root content view that handles authentication routing
struct ContentView: View {
    @Environment(AuthService.self) private var authService
    @Environment(GuildService.self) private var guildService
    @Environment(EventService.self) private var eventService
    @Environment(ProfileService.self) private var profileService
    @Environment(DiscoveryService.self) private var discoveryService

    var body: some View {
        Group {
            if authService.isAuthenticated {
                MainTabView()
            } else {
                AuthView()
            }
        }
        .animation(.easeInOut, value: authService.isAuthenticated)
        .onChange(of: authService.isAuthenticated) { _, isAuthenticated in
            if !isAuthenticated {
                // Clear data on logout
                Task {
                    await guildService.clear()
                    await eventService.clear()
                    profileService.clear()
                    discoveryService.clear()
                }
            }
        }
    }
}

/// Main tab view for authenticated users
struct MainTabView: View {
    @Environment(GuildService.self) private var guildService

    var body: some View {
        TabView {
            // Guilds tab
            NavigationStack {
                GuildListView()
                    .navigationDestination(for: GuildDestination.self) { destination in
                        guildDestinationView(for: destination)
                    }
            }
            .tabItem {
                Label("Guilds", systemImage: "person.3.fill")
            }

            // Events tab
            NavigationStack {
                EventListView()
                    .navigationDestination(for: EventDestination.self) { destination in
                        switch destination {
                        case .detail(let event):
                            EventDetailView(event: event)
                        case .create:
                            Text("Create Event")
                        }
                    }
            }
            .tabItem {
                Label("Events", systemImage: "calendar")
            }

            // Discover tab
            NavigationStack {
                DiscoverView()
            }
            .tabItem {
                Label("Discover", systemImage: "sparkle.magnifyingglass")
            }

            // Profile tab
            NavigationStack {
                ProfileView()
            }
            .tabItem {
                Label("Profile", systemImage: "person.crop.circle")
            }
        }
    }

    @ViewBuilder
    private func guildDestinationView(for destination: GuildDestination) -> some View {
        switch destination {
        case .detail(let guild):
            GuildDetailView(guild: guild)
        case .person(let person):
            PersonDetailView(person: person)
        case .createPerson:
            CreatePersonSheet()
        case .settings:
            Text("Guild Settings")
        case .events:
            if let guildId = guildService.currentGuild?.guild.id {
                GuildEventListView(guildId: guildId)
            } else {
                Text("Select a guild first")
            }
        case .adventures:
            if let guildId = guildService.currentGuild?.guild.id {
                AdventureListView(guildId: guildId)
            } else {
                Text("Select a guild first")
            }
        case .pools:
            if let guildId = guildService.currentGuild?.guild.id {
                PoolListView(guildId: guildId)
            } else {
                Text("Select a guild first")
            }
        case .votes:
            if let guildId = guildService.currentGuild?.guild.id {
                VoteListView(guildId: guildId)
            } else {
                Text("Select a guild first")
            }
        }
    }
}

// MARK: - Profile View

struct ProfileView: View {
    @Environment(AuthService.self) private var authService
    @Environment(ProfileService.self) private var profileService
    @State private var showingLogoutAlert = false
    @State private var showEditProfile = false

    var body: some View {
        List {
            // User info section
            if let user = authService.currentUser {
                Section {
                    HStack {
                        // Avatar
                        ZStack {
                            Circle()
                                .fill(.blue.gradient)
                                .frame(width: 60, height: 60)
                            Text(user.initials)
                                .font(.title2.bold())
                                .foregroundStyle(.white)
                        }

                        VStack(alignment: .leading, spacing: 4) {
                            Text(user.displayName)
                                .font(.headline)
                            if let email = user.email {
                                Text(email)
                                    .font(.subheadline)
                                    .foregroundStyle(.secondary)
                            }
                            if let profile = profileService.myProfile, let bio = profile.bio {
                                Text(bio)
                                    .font(.caption)
                                    .foregroundStyle(.secondary)
                                    .lineLimit(2)
                            }
                        }
                        .padding(.leading, 8)
                    }
                    .padding(.vertical, 8)

                    Button {
                        showEditProfile = true
                    } label: {
                        Label("Edit Profile", systemImage: "pencil")
                    }
                }
            }

            // Social section
            Section("Social") {
                NavigationLink {
                    AvailabilityListView()
                } label: {
                    HStack {
                        Image(systemName: "calendar.badge.clock")
                            .foregroundStyle(.green)
                            .frame(width: 24)
                        Text("My Availability")
                    }
                }

                NavigationLink {
                    NearbyAvailabilityView()
                } label: {
                    HStack {
                        Image(systemName: "person.wave.2")
                            .foregroundStyle(.orange)
                            .frame(width: 24)
                        Text("Find Nearby")
                    }
                }

                NavigationLink {
                    TrustManagementView()
                } label: {
                    HStack {
                        Image(systemName: "shield.fill")
                            .foregroundStyle(.blue)
                            .frame(width: 24)
                        Text("Trust & Connections")
                    }
                }
            }

            // Linked accounts section
            Section("Linked Accounts") {
                ForEach(authService.identities) { identity in
                    HStack {
                        Image(systemName: identity.provider.iconName)
                            .foregroundStyle(.secondary)
                            .frame(width: 24)
                        Text(identity.provider.displayName)
                        Spacer()
                        if identity.verified {
                            Image(systemName: "checkmark.circle.fill")
                                .foregroundStyle(.green)
                        }
                    }
                }

                if !authService.passkeys.isEmpty {
                    ForEach(authService.passkeys) { passkey in
                        HStack {
                            Image(systemName: "person.badge.key.fill")
                                .foregroundStyle(.secondary)
                                .frame(width: 24)
                            Text(passkey.name)
                            Spacer()
                            Image(systemName: "checkmark.circle.fill")
                                .foregroundStyle(.green)
                        }
                    }
                }
            }

            // Resonance section
            Section("Resonance") {
                NavigationLink {
                    ResonanceView()
                } label: {
                    HStack {
                        Image(systemName: "sparkles")
                            .foregroundStyle(.purple)
                            .frame(width: 24)
                        Text("My Resonance")
                    }
                }

                NavigationLink {
                    ResonanceLeaderboardView()
                } label: {
                    HStack {
                        Image(systemName: "trophy.fill")
                            .foregroundStyle(.yellow)
                            .frame(width: 24)
                        Text("Leaderboard")
                    }
                }
            }

            // Privacy & Safety section
            Section("Privacy & Safety") {
                NavigationLink {
                    BlockedUsersView()
                } label: {
                    HStack {
                        Image(systemName: "person.slash.fill")
                            .foregroundStyle(.red)
                            .frame(width: 24)
                        Text("Blocked Users")
                    }
                }

                NavigationLink {
                    ModerationStatusView()
                } label: {
                    HStack {
                        Image(systemName: "shield.checkered")
                            .foregroundStyle(.orange)
                            .frame(width: 24)
                        Text("Account Status")
                    }
                }
            }

            // Security section
            Section("Security") {
                NavigationLink {
                    PasskeySetupView()
                } label: {
                    HStack {
                        Image(systemName: "person.badge.key.fill")
                            .foregroundStyle(.blue)
                            .frame(width: 24)
                        Text("Manage Passkeys")
                    }
                }
            }

            // Settings section
            Section {
                NavigationLink {
                    SettingsView()
                } label: {
                    HStack {
                        Image(systemName: "gearshape.fill")
                            .foregroundStyle(.gray)
                            .frame(width: 24)
                        Text("Settings")
                    }
                }
            }

            // Logout section
            Section {
                Button(role: .destructive) {
                    showingLogoutAlert = true
                } label: {
                    HStack {
                        Spacer()
                        if authService.isLoading {
                            ProgressView()
                        } else {
                            Text("Sign Out")
                        }
                        Spacer()
                    }
                }
                .disabled(authService.isLoading)
            }
        }
        .navigationTitle("Profile")
        .task {
            await profileService.loadMyProfile()
        }
        .sheet(isPresented: $showEditProfile) {
            NavigationStack {
                ProfileEditView()
            }
        }
        .alert("Sign Out", isPresented: $showingLogoutAlert) {
            Button("Cancel", role: .cancel) { }
            Button("Sign Out", role: .destructive) {
                Task {
                    try? await authService.logout()
                }
            }
        } message: {
            Text("Are you sure you want to sign out?")
        }
    }
}

#Preview {
    ContentView()
        .environment(AuthService.shared)
        .environment(GuildService.shared)
        .environment(EventService.shared)
        .environment(ProfileService.shared)
        .environment(DiscoveryService.shared)
}
