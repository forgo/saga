import SwiftUI

/// Event list view for a specific guild (used when navigating from guild detail)
struct GuildEventListView: View {
    let guildId: String

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
                            NavigationLink {
                                EventDetailView(event: event)
                            } label: {
                                EventRow(event: event)
                            }
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
                            NavigationLink {
                                EventDetailView(event: event)
                            } label: {
                                EventRow(event: event)
                            }
                        }
                    }
                }
            }
        }
        .navigationTitle("Events")
        .toolbar {
            ToolbarItem(placement: .primaryAction) {
                Button {
                    showingCreateSheet = true
                } label: {
                    Image(systemName: "plus")
                }
            }
        }
        .refreshable {
            await eventService.loadGuildEvents(guildId: guildId)
        }
        .task {
            await eventService.loadGuildEvents(guildId: guildId)
        }
        .sheet(isPresented: $showingCreateSheet) {
            CreateEventSheet(guildId: guildId)
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

#Preview {
    NavigationStack {
        GuildEventListView(guildId: "test-guild")
            .environment(EventService.shared)
    }
}
