import SwiftUI

/// Detail view for a person showing their timers
struct PersonDetailView: View {
    let person: Person

    @Environment(GuildService.self) private var guildService
    @State private var timers: [TimerWithActivity] = []
    @State private var isLoading = false
    @State private var showingAddTimer = false
    @State private var errorMessage: String?
    @State private var isShowingError = false

    var body: some View {
        List {
            // Person info section
            Section {
                HStack(spacing: 16) {
                    ZStack {
                        Circle()
                            .fill(.gray.opacity(0.2))
                            .frame(width: 70, height: 70)
                        Text(person.initials)
                            .font(.title)
                            .foregroundStyle(.secondary)
                    }

                    VStack(alignment: .leading, spacing: 4) {
                        Text(person.name)
                            .font(.title2.bold())
                        if let nickname = person.nickname {
                            Text("\"\(nickname)\"")
                                .font(.subheadline)
                                .foregroundStyle(.secondary)
                        }
                        if let birthdayDate = person.birthdayDate {
                            HStack {
                                Image(systemName: "gift.fill")
                                    .foregroundStyle(.orange)
                                Text(birthdayDate, format: .dateTime.month().day())
                                if let days = person.daysUntilBirthday {
                                    if days == 0 {
                                        Text("(Today!)")
                                            .foregroundStyle(.orange)
                                    } else {
                                        Text("(in \(days) days)")
                                            .foregroundStyle(.secondary)
                                    }
                                }
                            }
                            .font(.caption)
                        }
                    }
                }
                .padding(.vertical, 8)

                if let notes = person.notes, !notes.isEmpty {
                    VStack(alignment: .leading, spacing: 4) {
                        Text("Notes")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                        Text(notes)
                            .font(.body)
                    }
                }
            }

            // Timers section
            Section("Activity Timers") {
                if timers.isEmpty && !isLoading {
                    ContentUnavailableView(
                        "No Timers",
                        systemImage: "timer",
                        description: Text("Add a timer to track when you last did an activity")
                    )
                    .listRowBackground(Color.clear)
                } else {
                    ForEach(timers) { timerWithActivity in
                        TimerCard(
                            timer: timerWithActivity.timer,
                            activity: timerWithActivity.activity,
                            onReset: {
                                await resetTimer(timerWithActivity.timer.id)
                            }
                        )
                    }
                    .onDelete(perform: deleteTimers)
                }
            }
        }
        .navigationTitle(person.displayName)
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .primaryAction) {
                Button {
                    showingAddTimer = true
                } label: {
                    Image(systemName: "plus")
                }
                .disabled(availableActivities.isEmpty)
            }
        }
        .refreshable {
            await loadTimers()
        }
        .task {
            await loadTimers()
        }
        .onReceive(NotificationCenter.default.publisher(for: .guildTimerEvent)) { notification in
            if let event = notification.object as? GuildEvent {
                handleTimerEvent(event)
            }
        }
        .sheet(isPresented: $showingAddTimer) {
            AddTimerSheet(person: person, availableActivities: availableActivities) { timer, activity in
                timers.append(TimerWithActivity(timer: timer, activity: activity))
            }
        }
        .alert("Error", isPresented: $isShowingError) {
            Button("OK", role: .cancel) { }
        } message: {
            Text(errorMessage ?? "An error occurred")
        }
    }

    private var availableActivities: [Activity] {
        guard let data = guildService.currentGuild else { return [] }
        let usedActivityIds = Set(timers.map { $0.activity.id })
        return data.activities.filter { !usedActivityIds.contains($0.id) }
    }

    private func loadTimers() async {
        guard let guildId = guildService.currentGuild?.guild.id else { return }

        isLoading = true
        defer { isLoading = false }

        do {
            let response = try await APIClient.shared.listTimers(guildId: guildId, personId: person.id)
            timers = response.data
        } catch {
            errorMessage = error.localizedDescription
            isShowingError = true
        }
    }

    private func resetTimer(_ timerId: String) async {
        do {
            let updatedTimer = try await guildService.resetTimer(personId: person.id, timerId: timerId)
            if let index = timers.firstIndex(where: { $0.timer.id == timerId }) {
                timers[index] = TimerWithActivity(timer: updatedTimer, activity: timers[index].activity)
            }
        } catch {
            errorMessage = error.localizedDescription
            isShowingError = true
        }
    }

    private func deleteTimers(at offsets: IndexSet) {
        Task {
            for index in offsets {
                let timer = timers[index]
                do {
                    try await guildService.deleteTimer(personId: person.id, timerId: timer.timer.id)
                    timers.remove(at: index)
                } catch {
                    errorMessage = error.localizedDescription
                    isShowingError = true
                }
            }
        }
    }

    private func handleTimerEvent(_ event: GuildEvent) {
        switch event {
        case .timerReset(let timer), .timerUpdated(let timer):
            if let index = timers.firstIndex(where: { $0.timer.id == timer.id }) {
                timers[index] = TimerWithActivity(timer: timer, activity: timers[index].activity)
            }
        case .timerDeleted(let id):
            timers.removeAll { $0.timer.id == id }
        default:
            break
        }
    }
}

// MARK: - Add Timer Sheet

struct AddTimerSheet: View {
    let person: Person
    let availableActivities: [Activity]
    let onAdd: (ActivityTimer, Activity) -> Void

    @Environment(\.dismiss) private var dismiss
    @Environment(GuildService.self) private var guildService

    @State private var selectedActivity: Activity?
    @State private var pushEnabled = false
    @State private var isCreating = false
    @State private var errorMessage: String?

    var body: some View {
        NavigationStack {
            Form {
                Section("Activity") {
                    ForEach(availableActivities) { activity in
                        Button {
                            selectedActivity = activity
                        } label: {
                            HStack {
                                Image(systemName: activity.icon)
                                    .frame(width: 24)
                                Text(activity.name)
                                Spacer()
                                if selectedActivity?.id == activity.id {
                                    Image(systemName: "checkmark")
                                        .foregroundStyle(.blue)
                                }
                            }
                            .foregroundStyle(.primary)
                        }
                    }
                }

                if let activity = selectedActivity {
                    Section("Thresholds") {
                        HStack {
                            Label("Warning", systemImage: "exclamationmark.triangle.fill")
                                .foregroundStyle(.orange)
                            Spacer()
                            Text(activity.warnDescription)
                                .foregroundStyle(.secondary)
                        }
                        HStack {
                            Label("Critical", systemImage: "exclamationmark.octagon.fill")
                                .foregroundStyle(.red)
                            Spacer()
                            Text(activity.criticalDescription)
                                .foregroundStyle(.secondary)
                        }
                    }

                    Section {
                        Toggle("Push notifications", isOn: $pushEnabled)
                    } footer: {
                        Text("Get notified when this timer reaches warning or critical status")
                    }
                }

                if let error = errorMessage {
                    Section {
                        Text(error)
                            .foregroundStyle(.red)
                    }
                }
            }
            .navigationTitle("Add Timer")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") {
                        dismiss()
                    }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Add") {
                        createTimer()
                    }
                    .disabled(selectedActivity == nil || isCreating)
                }
            }
        }
    }

    private func createTimer() {
        guard let activity = selectedActivity else { return }

        isCreating = true
        errorMessage = nil

        Task {
            do {
                let timer = try await guildService.createTimer(
                    personId: person.id,
                    activityId: activity.id,
                    push: pushEnabled
                )
                onAdd(timer, activity)
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
        PersonDetailView(person: Person(
            id: "person:1",
            name: "Alex Chen",
            nickname: "Alex",
            birthday: "1990-03-15",
            notes: "Met at college, loves hiking",
            createdOn: Date(),
            updatedOn: Date()
        ))
        .environment(GuildService.shared)
    }
}
