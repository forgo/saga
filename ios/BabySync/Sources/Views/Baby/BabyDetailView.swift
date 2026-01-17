import SwiftUI

struct BabyDetailView: View {
    let baby: Baby
    let activities: [Activity]
    let familyId: String

    @State private var timers: [BabyTimer] = []
    @State private var isLoading = true
    @State private var showAddTimer = false

    private let apiClient = APIClient.shared

    var body: some View {
        ScrollView {
            LazyVStack(spacing: 16) {
                // Baby header
                BabyHeader(baby: baby)

                // Timers section
                VStack(alignment: .leading, spacing: 12) {
                    HStack {
                        Text("Timers")
                            .font(.headline)
                        Spacer()
                        Button {
                            showAddTimer = true
                        } label: {
                            Image(systemName: "plus.circle.fill")
                                .font(.title2)
                        }
                    }

                    if isLoading {
                        ProgressView()
                            .frame(maxWidth: .infinity)
                            .padding()
                    } else if timers.isEmpty {
                        Text("No timers yet. Add a timer to track activities!")
                            .foregroundStyle(.secondary)
                            .frame(maxWidth: .infinity)
                            .padding()
                    } else {
                        ForEach(timers) { timer in
                            if let activity = activities.first(where: { $0.id == timer.activityId }) {
                                TimerCard(timer: timer, activity: activity, familyId: familyId, babyId: baby.id) {
                                    await loadTimers()
                                }
                            }
                        }
                    }
                }
            }
            .padding()
        }
        .navigationTitle(baby.name)
        .sheet(isPresented: $showAddTimer) {
            AddTimerSheet(
                activities: activities,
                familyId: familyId,
                babyId: baby.id
            ) {
                await loadTimers()
            }
        }
        .task {
            await loadTimers()
        }
        .onReceive(NotificationCenter.default.publisher(for: .timerReset)) { notification in
            if let updatedTimer = notification.object as? BabyTimer,
               updatedTimer.babyId == baby.id {
                if let index = timers.firstIndex(where: { $0.id == updatedTimer.id }) {
                    timers[index] = updatedTimer
                }
            }
        }
    }

    private func loadTimers() async {
        isLoading = true
        defer { isLoading = false }

        do {
            let response = try await apiClient.listTimers(familyId: familyId, babyId: baby.id)
            timers = response.data
        } catch {
            print("Failed to load timers: \(error)")
        }
    }
}

struct BabyHeader: View {
    let baby: Baby

    var body: some View {
        VStack(spacing: 12) {
            Image(systemName: "figure.child")
                .font(.system(size: 60))
                .foregroundStyle(.pink)

            Text(baby.name)
                .font(.title)
                .fontWeight(.bold)
        }
        .frame(maxWidth: .infinity)
        .padding()
        #if os(iOS)
        .background(Color(.systemGray6))
        #else
        .background(Color.gray.opacity(0.1))
        #endif
        .cornerRadius(16)
    }
}

struct AddTimerSheet: View {
    let activities: [Activity]
    let familyId: String
    let babyId: String
    let onAdd: () async -> Void

    @Environment(\.dismiss) private var dismiss
    @State private var selectedActivityId: String?
    @State private var enablePush = false

    private let apiClient = APIClient.shared

    var body: some View {
        NavigationStack {
            Form {
                Section("Activity") {
                    Picker("Select Activity", selection: $selectedActivityId) {
                        Text("Select an activity").tag(nil as String?)
                        ForEach(activities) { activity in
                            HStack {
                                Text(activity.icon)
                                Text(activity.name)
                            }
                            .tag(activity.id as String?)
                        }
                    }
                }

                Section {
                    Toggle("Push Notifications", isOn: $enablePush)
                } footer: {
                    Text("Get notified when this timer reaches warning or critical thresholds")
                }
            }
            .navigationTitle("Add Timer")
            #if os(iOS)
            .navigationBarTitleDisplayMode(.inline)
            #endif
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Add") {
                        Task {
                            await addTimer()
                        }
                    }
                    .disabled(selectedActivityId == nil)
                }
            }
        }
        .presentationDetents([.medium])
    }

    private func addTimer() async {
        guard let activityId = selectedActivityId else { return }

        let request = CreateTimerRequest(activityId: activityId, enabled: true, push: enablePush)
        _ = try? await apiClient.createTimer(familyId: familyId, babyId: babyId, request: request)
        await onAdd()
        dismiss()
    }
}

#Preview {
    NavigationStack {
        BabyDetailView(
            baby: Baby(id: "baby:1", familyId: "family:1", name: "Emma", createdOn: .now, updatedOn: .now),
            activities: [
                Activity(id: "act:1", familyId: "family:1", name: "Feeding", icon: "🍼", warn: 10800, critical: 14400, createdOn: .now, updatedOn: .now)
            ],
            familyId: "family:1"
        )
    }
}
