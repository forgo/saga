import SwiftUI

struct TimerCard: View {
    let timer: BabyTimer
    let activity: Activity
    let familyId: String
    let babyId: String
    let onUpdate: () async -> Void

    @State private var elapsed: TimeInterval = 0
    @State private var isResetting = false

    private let updateTimer = Timer.publish(every: 1, on: .main, in: .common).autoconnect()
    private let apiClient = APIClient.shared

    var timerState: TimerState {
        if elapsed >= activity.critical {
            return .critical
        } else if elapsed >= activity.warn {
            return .warning
        }
        return .normal
    }

    var body: some View {
        VStack(spacing: 0) {
            // Timer header
            HStack {
                Text(activity.icon)
                    .font(.title)

                VStack(alignment: .leading) {
                    Text(activity.name)
                        .font(.headline)
                    Text(timer.enabled ? "Active" : "Paused")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }

                Spacer()

                // Timer state indicator
                Circle()
                    .fill(timerState.color)
                    .frame(width: 12, height: 12)
            }
            .padding()

            Divider()

            // Elapsed time display
            HStack {
                VStack(alignment: .leading) {
                    Text("Time Since Last")
                        .font(.caption)
                        .foregroundStyle(.secondary)

                    Text(formatElapsed(elapsed))
                        .font(.system(.title, design: .monospaced))
                        .fontWeight(.semibold)
                        .foregroundStyle(timerState.color)
                }

                Spacer()

                // Reset button
                Button {
                    Task { await resetTimer() }
                } label: {
                    if isResetting {
                        ProgressView()
                            .frame(width: 44, height: 44)
                    } else {
                        Image(systemName: "arrow.counterclockwise.circle.fill")
                            .font(.system(size: 44))
                            .foregroundStyle(.blue)
                    }
                }
                .disabled(isResetting)
            }
            .padding()

            // Threshold info
            HStack {
                Label("\(activity.warnFormatted)", systemImage: "exclamationmark.triangle")
                    .font(.caption)
                    .foregroundStyle(.orange)

                Spacer()

                Label("\(activity.criticalFormatted)", systemImage: "exclamationmark.octagon")
                    .font(.caption)
                    .foregroundStyle(.red)
            }
            .padding(.horizontal)
            .padding(.bottom)
        }
        .background(timerState.backgroundColor)
        .cornerRadius(16)
        .overlay(
            RoundedRectangle(cornerRadius: 16)
                .stroke(timerState.color.opacity(0.3), lineWidth: 1)
        )
        .onReceive(updateTimer) { _ in
            updateElapsed()
        }
        .onAppear {
            updateElapsed()
        }
    }

    private func updateElapsed() {
        if timer.enabled {
            elapsed = Date().timeIntervalSince(timer.resetDate)
        }
    }

    private func formatElapsed(_ seconds: TimeInterval) -> String {
        let totalSeconds = Int(max(0, seconds))
        let hours = totalSeconds / 3600
        let minutes = (totalSeconds % 3600) / 60
        let secs = totalSeconds % 60
        return String(format: "%02d:%02d:%02d", hours, minutes, secs)
    }

    private func resetTimer() async {
        isResetting = true
        defer { isResetting = false }

        do {
            _ = try await apiClient.resetTimer(familyId: familyId, babyId: babyId, timerId: timer.id)
            await onUpdate()
        } catch {
            print("Failed to reset timer: \(error)")
        }
    }
}

enum TimerState {
    case normal
    case warning
    case critical

    var color: Color {
        switch self {
        case .normal: return .green
        case .warning: return .orange
        case .critical: return .red
        }
    }

    var backgroundColor: Color {
        switch self {
        case .normal:
            #if os(iOS)
            return Color(.systemBackground)
            #else
            return Color.white
            #endif
        case .warning: return Color.orange.opacity(0.1)
        case .critical: return Color.red.opacity(0.1)
        }
    }
}

#Preview {
    VStack(spacing: 16) {
        TimerCard(
            timer: BabyTimer(
                id: "timer:1",
                babyId: "baby:1",
                activityId: "act:1",
                resetDate: Date().addingTimeInterval(-3600), // 1 hour ago
                enabled: true,
                push: true,
                createdOn: .now,
                updatedOn: .now
            ),
            activity: Activity(
                id: "act:1",
                familyId: "family:1",
                name: "Feeding",
                icon: "🍼",
                warn: 10800,
                critical: 14400,
                createdOn: .now,
                updatedOn: .now
            ),
            familyId: "family:1",
            babyId: "baby:1"
        ) { }

        TimerCard(
            timer: BabyTimer(
                id: "timer:2",
                babyId: "baby:1",
                activityId: "act:2",
                resetDate: Date().addingTimeInterval(-12000), // ~3.3 hours ago (warning)
                enabled: true,
                push: true,
                createdOn: .now,
                updatedOn: .now
            ),
            activity: Activity(
                id: "act:2",
                familyId: "family:1",
                name: "Diaper",
                icon: "🧷",
                warn: 10800,
                critical: 14400,
                createdOn: .now,
                updatedOn: .now
            ),
            familyId: "family:1",
            babyId: "baby:1"
        ) { }
    }
    .padding()
}
