import SwiftUI

/// View for browsing nearby availabilities
struct NearbyAvailabilityView: View {
    @Environment(ProfileService.self) private var profileService

    @State private var selectedType: HangoutType?
    @State private var selectedAvailability: NearbyAvailability?
    @State private var showRequestSheet = false

    var body: some View {
        List {
            // Type Filter
            Section {
                ScrollView(.horizontal, showsIndicators: false) {
                    HStack(spacing: 8) {
                        FilterChip(
                            title: "All",
                            isSelected: selectedType == nil
                        ) {
                            selectedType = nil
                        }

                        ForEach(HangoutType.allCases, id: \.self) { type in
                            FilterChip(
                                title: type.displayName,
                                icon: type.iconName,
                                isSelected: selectedType == type
                            ) {
                                selectedType = type
                            }
                        }
                    }
                    .padding(.horizontal)
                }
            }
            .listRowInsets(EdgeInsets())
            .listRowBackground(Color.clear)

            // Nearby Availabilities
            if profileService.isLoadingAvailability && profileService.nearbyAvailabilities.isEmpty {
                Section {
                    HStack {
                        Spacer()
                        ProgressView()
                        Spacer()
                    }
                    .padding(.vertical, 40)
                }
                .listRowBackground(Color.clear)
            } else if filteredAvailabilities.isEmpty {
                ContentUnavailableView {
                    Label("No Nearby Availabilities", systemImage: "location.slash")
                } description: {
                    Text("No one is available nearby right now. Check back later!")
                }
                .listRowBackground(Color.clear)
            } else {
                ForEach(filteredAvailabilities, id: \.availability.id) { nearby in
                    NearbyAvailabilityRow(nearby: nearby)
                        .contentShape(Rectangle())
                        .onTapGesture {
                            selectedAvailability = nearby
                            showRequestSheet = true
                        }
                }
            }
        }
        .navigationTitle("Nearby")
        .refreshable {
            // Would use actual location in production
            await profileService.searchNearbyAvailabilities(
                lat: 37.7749,
                lng: -122.4194,
                hangoutType: selectedType
            )
        }
        .sheet(isPresented: $showRequestSheet) {
            if let nearby = selectedAvailability {
                NavigationStack {
                    HangoutRequestSheet(nearby: nearby)
                }
            }
        }
        .task {
            // Load with mock location - in production use CoreLocation
            await profileService.searchNearbyAvailabilities(
                lat: 37.7749,
                lng: -122.4194,
                hangoutType: selectedType
            )
        }
        .onChange(of: selectedType) { _, newType in
            Task {
                await profileService.searchNearbyAvailabilities(
                    lat: 37.7749,
                    lng: -122.4194,
                    hangoutType: newType
                )
            }
        }
    }

    private var filteredAvailabilities: [NearbyAvailability] {
        guard let type = selectedType else {
            return profileService.nearbyAvailabilities
        }
        return profileService.nearbyAvailabilities.filter { $0.availability.hangoutType == type }
    }
}

// MARK: - Filter Chip

struct FilterChip: View {
    let title: String
    var icon: String?
    let isSelected: Bool
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            HStack(spacing: 4) {
                if let icon = icon {
                    Image(systemName: icon)
                        .font(.caption)
                }
                Text(title)
                    .font(.subheadline)
            }
            .padding(.horizontal, 12)
            .padding(.vertical, 8)
            .background(isSelected ? Color.blue : Color(.systemGray5))
            .foregroundStyle(isSelected ? .white : .primary)
            .clipShape(Capsule())
        }
        .buttonStyle(.plain)
    }
}

// MARK: - Nearby Availability Row

struct NearbyAvailabilityRow: View {
    let nearby: NearbyAvailability

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            // User Info Row
            HStack(spacing: 12) {
                // Avatar
                ZStack {
                    Circle()
                        .fill(.blue.gradient)
                        .frame(width: 44, height: 44)
                    Text(nearby.user.initials)
                        .font(.subheadline.bold())
                        .foregroundStyle(.white)
                }

                VStack(alignment: .leading, spacing: 2) {
                    Text(nearby.user.displayName)
                        .font(.headline)
                    if let distance = nearby.user.formattedDistance {
                        Text(distance)
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                }

                Spacer()

                // Compatibility Score
                if let score = nearby.compatibilityScore {
                    VStack(spacing: 2) {
                        Text("\(Int(score * 100))%")
                            .font(.headline)
                            .foregroundStyle(.green)
                        Text("match")
                            .font(.caption2)
                            .foregroundStyle(.secondary)
                    }
                }
            }

            // Availability Details
            VStack(alignment: .leading, spacing: 6) {
                Label(nearby.availability.hangoutType.displayName, systemImage: nearby.availability.hangoutType.iconName)
                    .font(.subheadline.bold())
                    .foregroundStyle(.blue)

                if let title = nearby.availability.title {
                    Text(title)
                        .font(.subheadline)
                }

                HStack {
                    Image(systemName: "clock")
                        .font(.caption)
                    Text(nearby.availability.timeRange)
                        .font(.caption)
                }
                .foregroundStyle(.secondary)

                if let location = nearby.availability.location {
                    HStack {
                        Image(systemName: "mappin")
                            .font(.caption)
                        Text(location)
                            .font(.caption)
                    }
                    .foregroundStyle(.secondary)
                }
            }
            .padding(12)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(Color(.systemGray6))
            .clipShape(RoundedRectangle(cornerRadius: 8))
        }
        .padding(.vertical, 4)
    }
}

// MARK: - Hangout Request Sheet

struct HangoutRequestSheet: View {
    let nearby: NearbyAvailability

    @Environment(ProfileService.self) private var profileService
    @Environment(\.dismiss) private var dismiss

    @State private var message: String = ""
    @State private var isSending = false
    @State private var errorMessage: String?

    var body: some View {
        Form {
            // Availability Info
            Section {
                VStack(alignment: .leading, spacing: 8) {
                    HStack(spacing: 12) {
                        ZStack {
                            Circle()
                                .fill(.blue.gradient)
                                .frame(width: 50, height: 50)
                            Text(nearby.user.initials)
                                .font(.headline.bold())
                                .foregroundStyle(.white)
                        }

                        VStack(alignment: .leading) {
                            Text(nearby.user.displayName)
                                .font(.headline)
                            Text(nearby.availability.hangoutType.displayName)
                                .font(.subheadline)
                                .foregroundStyle(.secondary)
                        }
                    }

                    Divider()

                    if let title = nearby.availability.title {
                        Text(title)
                            .font(.subheadline)
                    }

                    Text(nearby.availability.timeRange)
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }
                .padding(.vertical, 4)
            }

            // Message Section
            Section("Message (Optional)") {
                TextField("Say something about yourself...", text: $message, axis: .vertical)
                    .lineLimit(3...6)
            }

            // Error Section
            if let errorMessage = errorMessage {
                Section {
                    Text(errorMessage)
                        .foregroundStyle(.red)
                }
            }
        }
        .navigationTitle("Request Hangout")
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .cancellationAction) {
                Button("Cancel") {
                    dismiss()
                }
            }

            ToolbarItem(placement: .confirmationAction) {
                Button("Send Request") {
                    Task {
                        await sendRequest()
                    }
                }
                .disabled(isSending)
            }
        }
        .disabled(isSending)
    }

    private func sendRequest() async {
        isSending = true
        errorMessage = nil

        do {
            _ = try await profileService.sendHangoutRequest(
                availabilityId: nearby.availability.id,
                message: message.isEmpty ? nil : message
            )
            dismiss()
        } catch {
            errorMessage = error.localizedDescription
        }

        isSending = false
    }
}

#Preview {
    NavigationStack {
        NearbyAvailabilityView()
            .environment(ProfileService.shared)
    }
}
