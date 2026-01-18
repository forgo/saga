import SwiftUI

/// Card displaying a timer with its status and reset action
struct TimerCard: View {
    let timer: ActivityTimer
    let activity: Activity
    let onReset: () async -> Void

    @State private var isResetting = false
    @State private var elapsedTime: TimeInterval = 0
    @State private var displayTimer: Timer?

    private var status: TimerStatus {
        timer.status(for: activity)
    }

    var body: some View {
        VStack(spacing: 12) {
            // Header with activity info
            HStack {
                Image(systemName: activity.icon)
                    .font(.title2)
                    .foregroundStyle(status.color)
                    .frame(width: 32)

                VStack(alignment: .leading, spacing: 2) {
                    Text(activity.name)
                        .font(.headline)
                    Text(formatElapsed(Int(elapsedTime)))
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                }

                Spacer()

                // Status indicator
                Image(systemName: status.iconName)
                    .foregroundStyle(status.color)
                    .font(.title2)
            }

            // Progress bar
            GeometryReader { geometry in
                ZStack(alignment: .leading) {
                    // Background
                    RoundedRectangle(cornerRadius: 4)
                        .fill(.gray.opacity(0.2))
                        .frame(height: 8)

                    // Warning zone
                    let warnWidth = geometry.size.width * (Double(activity.warn) / Double(activity.critical))
                    RoundedRectangle(cornerRadius: 4)
                        .fill(.orange.opacity(0.3))
                        .frame(width: warnWidth, height: 8)

                    // Progress
                    let progress = min(1.0, elapsedTime / Double(activity.critical))
                    RoundedRectangle(cornerRadius: 4)
                        .fill(status.color)
                        .frame(width: geometry.size.width * progress, height: 8)
                }
            }
            .frame(height: 8)

            // Thresholds and reset button
            HStack {
                VStack(alignment: .leading, spacing: 2) {
                    HStack(spacing: 4) {
                        Circle()
                            .fill(.orange)
                            .frame(width: 6, height: 6)
                        Text(activity.warnDescription)
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                    HStack(spacing: 4) {
                        Circle()
                            .fill(.red)
                            .frame(width: 6, height: 6)
                        Text(activity.criticalDescription)
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                }

                Spacer()

                Button {
                    Task {
                        isResetting = true
                        await onReset()
                        elapsedTime = 0
                        isResetting = false
                    }
                } label: {
                    HStack(spacing: 4) {
                        if isResetting {
                            ProgressView()
                                .scaleEffect(0.8)
                        } else {
                            Image(systemName: "arrow.counterclockwise")
                        }
                        Text("Reset")
                    }
                    .font(.subheadline.weight(.medium))
                    .padding(.horizontal, 12)
                    .padding(.vertical, 6)
                    .background(status.color.opacity(0.15))
                    .foregroundStyle(status.color)
                    .clipShape(Capsule())
                }
                .disabled(isResetting)
            }
        }
        .padding()
        .background(.background)
        .clipShape(RoundedRectangle(cornerRadius: 12))
        .shadow(color: .black.opacity(0.05), radius: 2, y: 1)
        .onAppear {
            elapsedTime = timer.elapsed
            startDisplayTimer()
        }
        .onDisappear {
            stopDisplayTimer()
        }
        .onChange(of: timer.resetDate) {
            elapsedTime = timer.elapsed
        }
    }

    private func startDisplayTimer() {
        displayTimer = Timer.scheduledTimer(withTimeInterval: 60, repeats: true) { [self] _ in
            Task { @MainActor in
                elapsedTime = timer.elapsed
            }
        }
    }

    private func stopDisplayTimer() {
        displayTimer?.invalidate()
        displayTimer = nil
    }

    private func formatElapsed(_ seconds: Int) -> String {
        let days = seconds / 86400
        let hours = (seconds % 86400) / 3600
        let minutes = (seconds % 3600) / 60

        if days > 0 {
            if hours > 0 {
                return "\(days)d \(hours)h ago"
            }
            return days == 1 ? "1 day ago" : "\(days) days ago"
        } else if hours > 0 {
            if minutes > 0 {
                return "\(hours)h \(minutes)m ago"
            }
            return hours == 1 ? "1 hour ago" : "\(hours) hours ago"
        } else if minutes > 0 {
            return minutes == 1 ? "1 minute ago" : "\(minutes) minutes ago"
        } else {
            return "Just now"
        }
    }
}

// MARK: - Compact Timer Row

/// A more compact timer display for list views
struct TimerRow: View {
    let timer: ActivityTimer
    let activity: Activity

    private var status: TimerStatus {
        timer.status(for: activity)
    }

    var body: some View {
        HStack(spacing: 12) {
            Image(systemName: activity.icon)
                .font(.body)
                .foregroundStyle(status.color)
                .frame(width: 24)

            Text(activity.name)
                .font(.subheadline)

            Spacer()

            Text(timer.elapsedDescription)
                .font(.caption)
                .foregroundStyle(.secondary)

            Image(systemName: status.iconName)
                .foregroundStyle(status.color)
                .font(.caption)
        }
    }
}

#Preview {
    VStack(spacing: 16) {
        TimerCard(
            timer: ActivityTimer(
                id: "timer:1",
                resetDate: Date().addingTimeInterval(-86400 * 5),
                enabled: true,
                push: true,
                createdOn: Date(),
                updatedOn: Date()
            ),
            activity: Activity(
                id: "activity:1",
                name: "Hung out",
                icon: "person.2.fill",
                warn: 14 * 86400,
                critical: 30 * 86400,
                createdOn: Date(),
                updatedOn: Date()
            ),
            onReset: {}
        )

        TimerCard(
            timer: ActivityTimer(
                id: "timer:2",
                resetDate: Date().addingTimeInterval(-86400 * 20),
                enabled: true,
                push: false,
                createdOn: Date(),
                updatedOn: Date()
            ),
            activity: Activity(
                id: "activity:2",
                name: "Called",
                icon: "phone.fill",
                warn: 7 * 86400,
                critical: 14 * 86400,
                createdOn: Date(),
                updatedOn: Date()
            ),
            onReset: {}
        )
    }
    .padding()
    .background(.gray.opacity(0.1))
}
