import SwiftUI

/// View displaying user's availability postings
struct AvailabilityListView: View {
    @Environment(ProfileService.self) private var profileService

    @State private var showCreateSheet = false
    @State private var selectedFilter: AvailabilityFilter = .active

    enum AvailabilityFilter: String, CaseIterable {
        case active = "Active"
        case all = "All"
    }

    var filteredAvailabilities: [Availability] {
        switch selectedFilter {
        case .active:
            return profileService.myAvailabilities.filter { $0.isActive }
        case .all:
            return profileService.myAvailabilities
        }
    }

    var body: some View {
        List {
            // Filter Picker
            Picker("Filter", selection: $selectedFilter) {
                ForEach(AvailabilityFilter.allCases, id: \.self) { filter in
                    Text(filter.rawValue).tag(filter)
                }
            }
            .pickerStyle(.segmented)
            .listRowBackground(Color.clear)
            .listRowInsets(EdgeInsets())
            .padding(.horizontal)

            // Availabilities
            if filteredAvailabilities.isEmpty {
                ContentUnavailableView {
                    Label("No Availabilities", systemImage: "calendar.badge.plus")
                } description: {
                    Text("Post your availability to meet new people")
                } actions: {
                    Button("Post Availability") {
                        showCreateSheet = true
                    }
                    .buttonStyle(.borderedProminent)
                }
                .listRowBackground(Color.clear)
            } else {
                ForEach(filteredAvailabilities) { availability in
                    AvailabilityRow(availability: availability)
                        .swipeActions(edge: .trailing, allowsFullSwipe: true) {
                            if availability.status == .active {
                                Button(role: .destructive) {
                                    Task {
                                        try? await profileService.cancelAvailability(availabilityId: availability.id)
                                    }
                                } label: {
                                    Label("Cancel", systemImage: "xmark")
                                }
                            }
                        }
                }
            }
        }
        .navigationTitle("My Availability")
        .toolbar {
            ToolbarItem(placement: .primaryAction) {
                Button {
                    showCreateSheet = true
                } label: {
                    Image(systemName: "plus")
                }
            }
        }
        .refreshable {
            await profileService.loadMyAvailabilities()
        }
        .sheet(isPresented: $showCreateSheet) {
            NavigationStack {
                CreateAvailabilitySheet()
            }
        }
        .task {
            await profileService.loadMyAvailabilities()
        }
    }
}

// MARK: - Availability Row

struct AvailabilityRow: View {
    let availability: Availability

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            // Header with type and status
            HStack {
                Label(availability.hangoutType.displayName, systemImage: availability.hangoutType.iconName)
                    .font(.headline)

                Spacer()

                AvailabilityStatusBadge(status: availability.status)
            }

            // Title if present
            if let title = availability.title {
                Text(title)
                    .font(.subheadline)
            }

            // Time range
            HStack {
                Image(systemName: "clock")
                    .foregroundStyle(.secondary)
                Text(availability.timeRange)
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }

            // Location if present
            if let location = availability.location {
                HStack {
                    Image(systemName: "mappin")
                        .foregroundStyle(.secondary)
                    Text(location)
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }
            }
        }
        .padding(.vertical, 4)
    }
}

// MARK: - Availability Status Badge

struct AvailabilityStatusBadge: View {
    let status: AvailabilityStatus

    var color: Color {
        switch status {
        case .active: return .green
        case .matched: return .blue
        case .expired: return .gray
        case .cancelled: return .red
        }
    }

    var body: some View {
        Text(status.displayName)
            .font(.caption2.bold())
            .padding(.horizontal, 8)
            .padding(.vertical, 4)
            .background(color.opacity(0.2))
            .foregroundStyle(color)
            .clipShape(Capsule())
    }
}

#Preview {
    NavigationStack {
        AvailabilityListView()
            .environment(ProfileService.shared)
    }
}
