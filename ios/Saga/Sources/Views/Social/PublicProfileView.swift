import SwiftUI

/// View for displaying another user's public profile
struct PublicProfileView: View {
    let userId: String

    @Environment(ProfileService.self) private var profileService

    @State private var profile: PublicProfile?
    @State private var trustSummary: TrustSummary?
    @State private var trustAggregate: TrustAggregate?
    @State private var isLoading = true
    @State private var error: Error?

    @State private var showTrustSheet = false
    @State private var showIRLSheet = false

    var body: some View {
        Group {
            if isLoading {
                ProgressView("Loading profile...")
            } else if let profile = profile {
                profileContent(profile)
            } else if let error = error {
                ContentUnavailableView {
                    Label("Couldn't Load Profile", systemImage: "person.crop.circle.badge.exclamationmark")
                } description: {
                    Text(error.localizedDescription)
                } actions: {
                    Button("Try Again") {
                        Task { await loadProfile() }
                    }
                }
            }
        }
        .navigationTitle("Profile")
        .navigationBarTitleDisplayMode(.inline)
        .task {
            await loadProfile()
        }
        .sheet(isPresented: $showTrustSheet) {
            NavigationStack {
                GrantTrustView(userId: userId)
            }
        }
        .sheet(isPresented: $showIRLSheet) {
            NavigationStack {
                RequestIRLView(userId: userId)
            }
        }
    }

    @ViewBuilder
    private func profileContent(_ profile: PublicProfile) -> some View {
        List {
            // Header Section
            Section {
                VStack(spacing: 16) {
                    // Avatar
                    ZStack {
                        Circle()
                            .fill(.blue.gradient)
                            .frame(width: 100, height: 100)

                        if let avatarUrl = profile.avatarUrl, let url = URL(string: avatarUrl) {
                            AsyncImage(url: url) { image in
                                image
                                    .resizable()
                                    .aspectRatio(contentMode: .fill)
                            } placeholder: {
                                Text(profile.initials)
                                    .font(.largeTitle.bold())
                                    .foregroundStyle(.white)
                            }
                            .frame(width: 100, height: 100)
                            .clipShape(Circle())
                        } else {
                            Text(profile.initials)
                                .font(.largeTitle.bold())
                                .foregroundStyle(.white)
                        }

                        // Online indicator
                        if profile.isOnline == true {
                            Circle()
                                .fill(.green)
                                .frame(width: 20, height: 20)
                                .overlay {
                                    Circle().stroke(.white, lineWidth: 3)
                                }
                                .offset(x: 35, y: 35)
                        }
                    }

                    // Name
                    Text(profile.displayName)
                        .font(.title2.bold())

                    // Distance
                    if let distance = profile.formattedDistance {
                        Label(distance, systemImage: "location")
                            .font(.subheadline)
                            .foregroundStyle(.secondary)
                    }
                }
                .frame(maxWidth: .infinity)
                .padding(.vertical)
            }

            // Bio Section
            if let bio = profile.bio, !bio.isEmpty {
                Section("About") {
                    Text(bio)
                        .font(.body)
                }
            }

            // Mutual Guilds Section
            if let guilds = profile.mutualGuilds, !guilds.isEmpty {
                Section("Mutual Guilds") {
                    ForEach(guilds, id: \.self) { guildId in
                        Label(guildId, systemImage: "person.3.fill")
                    }
                }
            }

            // Trust Section
            if let trust = trustSummary {
                Section("Trust") {
                    HStack {
                        TrustStatView(title: "Trusted By", count: trust.trustedByCount, icon: "shield.fill")
                        Divider()
                        TrustStatView(title: "Trusts", count: trust.trustsCount, icon: "hand.raised.fill")
                        Divider()
                        TrustStatView(title: "Mutual", count: trust.mutualTrustCount, icon: "arrow.left.arrow.right")
                        Divider()
                        TrustStatView(title: "IRL Met", count: trust.irlConfirmedCount, icon: "person.2.fill")
                    }
                    .padding(.vertical, 8)
                }
            }

            // Trust Ratings Section
            if let aggregate = trustAggregate {
                Section("Trust Ratings") {
                    HStack(spacing: 24) {
                        HStack {
                            Image(systemName: "hand.thumbsup.fill")
                                .foregroundStyle(.green)
                            Text("\(aggregate.trustCount)")
                                .font(.headline)
                        }

                        HStack {
                            Image(systemName: "hand.thumbsdown.fill")
                                .foregroundStyle(.red)
                            Text("\(aggregate.distrustCount)")
                                .font(.headline)
                        }

                        Spacer()

                        Text("\(aggregate.totalEndorsements) endorsements")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                    .padding(.vertical, 4)
                }
            }

            // Actions Section
            Section {
                Button {
                    showTrustSheet = true
                } label: {
                    Label("Grant Trust", systemImage: "shield.fill")
                }

                Button {
                    showIRLSheet = true
                } label: {
                    Label("Request IRL Confirmation", systemImage: "person.2.fill")
                }
            }
        }
    }

    private func loadProfile() async {
        isLoading = true
        error = nil

        do {
            async let profileFetch = profileService.getPublicProfile(userId: userId)
            async let trustFetch = profileService.getTrustSummary(userId: userId)
            async let aggregateFetch = profileService.getTrustAggregate(userId: userId)

            profile = try await profileFetch
            trustSummary = try? await trustFetch
            trustAggregate = try? await aggregateFetch
        } catch {
            self.error = error
        }

        isLoading = false
    }
}

// MARK: - Trust Stat View

struct TrustStatView: View {
    let title: String
    let count: Int
    let icon: String

    var body: some View {
        VStack(spacing: 4) {
            Image(systemName: icon)
                .font(.title3)
                .foregroundStyle(.blue)
            Text("\(count)")
                .font(.headline)
            Text(title)
                .font(.caption2)
                .foregroundStyle(.secondary)
        }
        .frame(maxWidth: .infinity)
    }
}

#Preview {
    NavigationStack {
        PublicProfileView(userId: "test-user-id")
            .environment(ProfileService.shared)
    }
}
