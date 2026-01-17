import SwiftUI

/// Sheet for editing pool preferences
struct PoolPreferencesSheet: View {
    let poolId: String
    let currentPreferences: PoolPreferences?

    @Environment(\.dismiss) private var dismiss

    @State private var availableDays: Set<Int> = []
    @State private var preferredTimes: Set<String> = []
    @State private var excludeRecentMatches = true
    @State private var notes = ""
    @State private var isSaving = false
    @State private var error: Error?

    private let apiClient = APIClient.shared

    private let dayNames = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"]
    private let timeSlots = ["morning", "afternoon", "evening"]

    var body: some View {
        Form {
            // Available Days
            Section {
                ForEach(0..<7, id: \.self) { day in
                    Toggle(dayNames[day], isOn: Binding(
                        get: { availableDays.contains(day) },
                        set: { isOn in
                            if isOn {
                                availableDays.insert(day)
                            } else {
                                availableDays.remove(day)
                            }
                        }
                    ))
                }
            } header: {
                Text("Available Days")
            } footer: {
                Text("Select the days you're available to meet")
            }

            // Preferred Times
            Section("Preferred Times") {
                ForEach(timeSlots, id: \.self) { slot in
                    Toggle(slot.capitalized, isOn: Binding(
                        get: { preferredTimes.contains(slot) },
                        set: { isOn in
                            if isOn {
                                preferredTimes.insert(slot)
                            } else {
                                preferredTimes.remove(slot)
                            }
                        }
                    ))
                }
            }

            // Options
            Section("Options") {
                Toggle("Exclude Recent Matches", isOn: $excludeRecentMatches)
            }

            // Notes
            Section("Notes") {
                TextField("Notes for matching (optional)", text: $notes, axis: .vertical)
                    .lineLimit(3...6)
            }
        }
        .navigationTitle("Preferences")
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .cancellationAction) {
                Button("Cancel") {
                    dismiss()
                }
            }
            ToolbarItem(placement: .confirmationAction) {
                Button("Save") {
                    Task { await savePreferences() }
                }
                .disabled(isSaving)
            }
        }
        .alert("Error", isPresented: .constant(error != nil)) {
            Button("OK") { error = nil }
        } message: {
            if let error = error {
                Text(error.localizedDescription)
            }
        }
        .onAppear {
            loadCurrentPreferences()
        }
    }

    private func loadCurrentPreferences() {
        guard let prefs = currentPreferences else { return }

        if let days = prefs.availableDays {
            availableDays = Set(days)
        }
        if let times = prefs.preferredTimes {
            preferredTimes = Set(times)
        }
        excludeRecentMatches = prefs.excludeRecentMatches
        notes = prefs.notes ?? ""
    }

    private func savePreferences() async {
        isSaving = true

        let preferences = PoolPreferences(
            availableDays: availableDays.isEmpty ? nil : Array(availableDays).sorted(),
            preferredTimes: preferredTimes.isEmpty ? nil : Array(preferredTimes),
            excludeRecentMatches: excludeRecentMatches,
            notes: notes.isEmpty ? nil : notes
        )

        do {
            _ = try await apiClient.updatePoolPreferences(poolId: poolId, preferences: preferences)
            dismiss()
        } catch {
            self.error = error
        }

        isSaving = false
    }
}

#Preview {
    NavigationStack {
        PoolPreferencesSheet(poolId: "test", currentPreferences: nil)
    }
}
