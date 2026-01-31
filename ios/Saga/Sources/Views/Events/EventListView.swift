import SwiftUI

/// List view displaying events for the current guild
struct EventListView: View {
    @Environment(GuildService.self) private var guildService
    @Environment(EventService.self) private var eventService

    @State private var showingCreateSheet = false
    @State private var selectedFilter: EventFilter = .upcoming

    enum EventFilter: String, CaseIterable {
        case upcoming = "Upcoming"
        case past = "Past"
        case all = "All"
    }

    private var filteredEvents: [Event] {
        switch selectedFilter {
        case .upcoming:
            return eventService.upcomingEvents
        case .past:
            return eventService.pastEvents
        case .all:
            return eventService.guildEvents
        }
    }

    var body: some View {
        List {
            // Filter picker
            Picker("Filter", selection: $selectedFilter) {
                ForEach(EventFilter.allCases, id: \.self) { filter in
                    Text(filter.rawValue).tag(filter)
                }
            }
            .pickerStyle(.segmented)
            .listRowBackground(Color.clear)
            .listRowInsets(EdgeInsets(top: 8, leading: 16, bottom: 8, trailing: 16))
            .accessibilityIdentifier("event_filter_picker")

            if filteredEvents.isEmpty && !eventService.isLoading {
                ContentUnavailableView(
                    selectedFilter == .upcoming ? "No Upcoming Events" : "No Events",
                    systemImage: "calendar",
                    description: Text(selectedFilter == .upcoming
                        ? "Create an event to get started"
                        : "No events found")
                )
                .listRowBackground(Color.clear)
            } else {
                // Today's events section
                let todayEvents = filteredEvents.filter { $0.isToday }
                if !todayEvents.isEmpty {
                    Section("Today") {
                        ForEach(todayEvents) { event in
                            NavigationLink(value: EventDestination.detail(event)) {
                                EventRow(event: event)
                            }
                            .accessibilityIdentifier("event_row_\(event.id)")
                        }
                    }
                }

                // Grouped by date
                let groupedEvents = Dictionary(grouping: filteredEvents.filter { !$0.isToday }) { event in
                    Calendar.current.startOfDay(for: event.startTime)
                }
                .sorted { $0.key < $1.key }

                ForEach(groupedEvents, id: \.key) { date, events in
                    Section(header: Text(formatSectionDate(date))) {
                        ForEach(events) { event in
                            NavigationLink(value: EventDestination.detail(event)) {
                                EventRow(event: event)
                            }
                            .accessibilityIdentifier("event_row_\(event.id)")
                        }
                    }
                }
            }
        }
        .accessibilityIdentifier("event_list")
        .navigationTitle("Events")
        .toolbar {
            ToolbarItem(placement: .primaryAction) {
                Button {
                    showingCreateSheet = true
                } label: {
                    Image(systemName: "plus")
                }
                .disabled(guildService.currentGuild == nil)
                .accessibilityIdentifier("event_create_button")
            }
        }
        .refreshable {
            if let guildId = guildService.currentGuild?.guild.id {
                await eventService.loadGuildEvents(guildId: guildId)
            }
        }
        .task {
            if let guildId = guildService.currentGuild?.guild.id {
                await eventService.loadGuildEvents(guildId: guildId)
            }
        }
        .sheet(isPresented: $showingCreateSheet) {
            if let guildId = guildService.currentGuild?.guild.id {
                CreateEventSheet(guildId: guildId)
            }
        }
    }

    private func formatSectionDate(_ date: Date) -> String {
        let formatter = DateFormatter()
        if Calendar.current.isDate(date, equalTo: Date(), toGranularity: .weekOfYear) {
            formatter.dateFormat = "EEEE"
        } else {
            formatter.dateStyle = .medium
        }
        return formatter.string(from: date)
    }
}

// MARK: - Event Row

struct EventRow: View {
    let event: Event

    var body: some View {
        HStack(spacing: 12) {
            // Time column
            VStack(alignment: .center, spacing: 2) {
                Text(formatTime(event.startTime))
                    .font(.subheadline.bold())
                if event.endTime != nil {
                    Text("-")
                        .font(.caption2)
                        .foregroundStyle(.secondary)
                    Text(formatTime(event.endTime!))
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }
            }
            .frame(width: 50)

            // Event info
            VStack(alignment: .leading, spacing: 4) {
                Text(event.title)
                    .font(.headline)
                    .lineLimit(1)

                if let location = event.location {
                    HStack(spacing: 4) {
                        Image(systemName: "mappin")
                            .font(.caption2)
                        Text(location)
                            .font(.caption)
                    }
                    .foregroundStyle(.secondary)
                    .lineLimit(1)
                }

                HStack(spacing: 8) {
                    // RSVP count
                    Label("\(event.rsvpCount)", systemImage: "person.2")
                        .font(.caption)
                        .foregroundStyle(.secondary)

                    // Capacity
                    if let capacity = event.capacity {
                        Text("/\(capacity)")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }

                    Spacer()

                    // Status badge
                    if event.status != .published {
                        StatusBadge(status: event.status)
                    }

                    // Visibility
                    if event.visibility != .guild {
                        Image(systemName: event.visibility.iconName)
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                }
            }
        }
        .padding(.vertical, 4)
        .opacity(event.isPast ? 0.6 : 1.0)
    }

    private func formatTime(_ date: Date) -> String {
        let formatter = DateFormatter()
        formatter.dateFormat = "h:mm"
        return formatter.string(from: date)
    }
}

// MARK: - Status Badge

struct StatusBadge: View {
    let status: EventStatus

    var body: some View {
        Text(status.displayName)
            .font(.caption2.bold())
            .padding(.horizontal, 6)
            .padding(.vertical, 2)
            .background(backgroundColor)
            .foregroundStyle(foregroundColor)
            .clipShape(Capsule())
    }

    private var backgroundColor: Color {
        switch status {
        case .draft: return .gray.opacity(0.2)
        case .published: return .green.opacity(0.2)
        case .cancelled: return .red.opacity(0.2)
        case .completed: return .blue.opacity(0.2)
        }
    }

    private var foregroundColor: Color {
        switch status {
        case .draft: return .gray
        case .published: return .green
        case .cancelled: return .red
        case .completed: return .blue
        }
    }
}

// MARK: - Event Destination

enum EventDestination: Hashable {
    case detail(Event)
    case create
}

#Preview {
    NavigationStack {
        EventListView()
            .environment(GuildService.shared)
            .environment(EventService.shared)
    }
}
