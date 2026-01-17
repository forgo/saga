import SwiftUI

/// Main discovery view for finding people and events
struct DiscoverView: View {
    @Environment(DiscoveryService.self) private var discoveryService

    @State private var selectedTab: DiscoverTab = .people
    @State private var minCompatibility: Double = 0.3
    @State private var radiusKm: Double = 25

    enum DiscoverTab: String, CaseIterable {
        case people = "People"
        case events = "Events"
        case interests = "Interests"
    }

    var body: some View {
        VStack(spacing: 0) {
            // Tab Picker
            Picker("Tab", selection: $selectedTab) {
                ForEach(DiscoverTab.allCases, id: \.self) { tab in
                    Text(tab.rawValue).tag(tab)
                }
            }
            .pickerStyle(.segmented)
            .padding()

            // Content
            Group {
                switch selectedTab {
                case .people:
                    peopleDiscoveryView
                case .events:
                    eventsDiscoveryView
                case .interests:
                    interestMatchesView
                }
            }
        }
        .navigationTitle("Discover")
        .toolbar {
            ToolbarItem(placement: .primaryAction) {
                NavigationLink {
                    InterestsView()
                } label: {
                    Image(systemName: "star.fill")
                }
            }
        }
    }

    // MARK: - People Discovery

    @ViewBuilder
    private var peopleDiscoveryView: some View {
        List {
            // Filters Section
            Section {
                VStack(alignment: .leading, spacing: 8) {
                    Text("Min Compatibility: \(Int(minCompatibility * 100))%")
                        .font(.subheadline)
                    Slider(value: $minCompatibility, in: 0...1, step: 0.1)
                }

                VStack(alignment: .leading, spacing: 8) {
                    Text("Search Radius: \(Int(radiusKm)) km")
                        .font(.subheadline)
                    Slider(value: $radiusKm, in: 5...100, step: 5)
                }

                Button("Search") {
                    Task {
                        await discoveryService.discoverPeople(
                            lat: 37.7749, // Mock location
                            lng: -122.4194,
                            radiusKm: radiusKm,
                            minCompatibility: minCompatibility
                        )
                    }
                }
                .frame(maxWidth: .infinity)
                .buttonStyle(.borderedProminent)
            }

            // Results Section
            if discoveryService.isLoadingDiscovery {
                Section {
                    HStack {
                        Spacer()
                        ProgressView()
                        Spacer()
                    }
                    .padding(.vertical, 20)
                }
            } else if discoveryService.discoveryResults.isEmpty {
                Section {
                    ContentUnavailableView {
                        Label("No People Found", systemImage: "person.slash")
                    } description: {
                        Text("Adjust your filters or try again later")
                    }
                }
                .listRowBackground(Color.clear)
            } else {
                Section("Found \(discoveryService.discoveryResults.count) People") {
                    ForEach(discoveryService.discoveryResults, id: \.user.userId) { result in
                        NavigationLink {
                            PublicProfileView(userId: result.user.userId)
                        } label: {
                            DiscoveryResultRow(result: result)
                        }
                    }
                }
            }
        }
        .task {
            await discoveryService.discoverPeople(
                lat: 37.7749,
                lng: -122.4194,
                radiusKm: radiusKm,
                minCompatibility: minCompatibility
            )
        }
    }

    // MARK: - Events Discovery

    @ViewBuilder
    private var eventsDiscoveryView: some View {
        List {
            if discoveryService.isLoadingDiscovery && discoveryService.nearbyEvents.isEmpty {
                Section {
                    HStack {
                        Spacer()
                        ProgressView()
                        Spacer()
                    }
                    .padding(.vertical, 20)
                }
            } else if discoveryService.nearbyEvents.isEmpty {
                Section {
                    ContentUnavailableView {
                        Label("No Events Nearby", systemImage: "calendar.badge.exclamationmark")
                    } description: {
                        Text("No events found in your area")
                    }
                }
                .listRowBackground(Color.clear)
            } else {
                ForEach(discoveryService.nearbyEvents) { event in
                    NavigationLink {
                        EventDetailView(event: event)
                    } label: {
                        EventDiscoveryRow(event: event)
                    }
                }
            }
        }
        .refreshable {
            await discoveryService.discoverEventsNearby(lat: 37.7749, lng: -122.4194)
        }
        .task {
            await discoveryService.discoverEventsNearby(lat: 37.7749, lng: -122.4194)
        }
    }

    // MARK: - Interest Matches

    @ViewBuilder
    private var interestMatchesView: some View {
        List {
            if discoveryService.isLoadingInterests && discoveryService.interestMatches.isEmpty {
                Section {
                    HStack {
                        Spacer()
                        ProgressView()
                        Spacer()
                    }
                    .padding(.vertical, 20)
                }
            } else if discoveryService.interestMatches.isEmpty {
                Section {
                    ContentUnavailableView {
                        Label("No Matches Yet", systemImage: "arrow.left.arrow.right")
                    } description: {
                        Text("Add more interests to find matches")
                    } actions: {
                        NavigationLink("Add Interests") {
                            InterestsView()
                        }
                        .buttonStyle(.borderedProminent)
                    }
                }
                .listRowBackground(Color.clear)
            } else {
                ForEach(discoveryService.interestMatches, id: \.user.userId) { match in
                    NavigationLink {
                        PublicProfileView(userId: match.user.userId)
                    } label: {
                        InterestMatchRow(match: match)
                    }
                }
            }
        }
        .refreshable {
            await discoveryService.loadInterestMatches()
        }
        .task {
            await discoveryService.loadInterestMatches()
        }
    }
}

// MARK: - Discovery Result Row

struct DiscoveryResultRow: View {
    let result: DiscoveryResult

    var body: some View {
        HStack(spacing: 12) {
            // Avatar
            ZStack {
                Circle()
                    .fill(.blue.gradient)
                    .frame(width: 50, height: 50)
                Text(result.user.initials)
                    .font(.headline.bold())
                    .foregroundStyle(.white)
            }

            // Info
            VStack(alignment: .leading, spacing: 4) {
                Text(result.user.displayName)
                    .font(.headline)

                HStack(spacing: 12) {
                    if let distance = result.formattedDistance {
                        Label(distance, systemImage: "location")
                            .font(.caption)
                    }

                    if result.sharedInterests > 0 {
                        Label("\(result.sharedInterests) shared", systemImage: "star.fill")
                            .font(.caption)
                    }

                    if result.mutualConnections > 0 {
                        Label("\(result.mutualConnections) mutual", systemImage: "person.2")
                            .font(.caption)
                    }
                }
                .foregroundStyle(.secondary)
            }

            Spacer()

            // Compatibility Score
            if let compat = result.formattedCompatibility {
                Text(compat)
                    .font(.subheadline.bold())
                    .foregroundStyle(.green)
            }
        }
        .padding(.vertical, 4)
    }
}

// MARK: - Event Discovery Row

struct EventDiscoveryRow: View {
    let event: Event

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(event.title)
                .font(.headline)

            if let description = event.description {
                Text(description)
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
                    .lineLimit(2)
            }

            HStack {
                Label(event.startTime.formatted(date: .abbreviated, time: .shortened), systemImage: "calendar")
                    .font(.caption)

                Spacer()

                if let location = event.location {
                    Label(location, systemImage: "mappin")
                        .font(.caption)
                }
            }
            .foregroundStyle(.secondary)
        }
        .padding(.vertical, 4)
    }
}

// MARK: - Interest Match Row

struct InterestMatchRow: View {
    let match: InterestMatch

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack(spacing: 12) {
                // Avatar
                ZStack {
                    Circle()
                        .fill(.purple.gradient)
                        .frame(width: 50, height: 50)
                    Text(match.user.initials)
                        .font(.headline.bold())
                        .foregroundStyle(.white)
                }

                VStack(alignment: .leading, spacing: 2) {
                    Text(match.user.displayName)
                        .font(.headline)
                    Text(match.matchQuality)
                        .font(.subheadline)
                        .foregroundStyle(.green)
                }

                Spacer()

                Text("\(Int(match.compatibilityScore * 100))%")
                    .font(.title2.bold())
                    .foregroundStyle(.green)
            }

            // Matching interests preview
            if !match.matchingInterests.isEmpty {
                ScrollView(.horizontal, showsIndicators: false) {
                    HStack(spacing: 8) {
                        ForEach(match.matchingInterests.prefix(3), id: \.interestId) { interest in
                            HStack(spacing: 4) {
                                Image(systemName: interest.matchType.iconName)
                                Text(interest.interestName)
                            }
                            .font(.caption)
                            .padding(.horizontal, 8)
                            .padding(.vertical, 4)
                            .background(Color(.systemGray5))
                            .clipShape(Capsule())
                        }

                        if match.matchingInterests.count > 3 {
                            Text("+\(match.matchingInterests.count - 3) more")
                                .font(.caption)
                                .foregroundStyle(.secondary)
                        }
                    }
                }
            }
        }
        .padding(.vertical, 4)
    }
}

#Preview {
    NavigationStack {
        DiscoverView()
            .environment(DiscoveryService.shared)
            .environment(ProfileService.shared)
    }
}
