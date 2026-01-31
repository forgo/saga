import SwiftUI

/// View for editing user's profile
struct ProfileEditView: View {
    @Environment(ProfileService.self) private var profileService
    @Environment(\.dismiss) private var dismiss

    @State private var displayName: String = ""
    @State private var bio: String = ""
    @State private var location: String = ""
    @State private var visibility: ProfileVisibility = .public
    @State private var showDistance: Bool = true
    @State private var showOnline: Bool = true

    @State private var isSaving = false
    @State private var errorMessage: String?

    var body: some View {
        Form {
            // Basic Info Section
            Section("Basic Info") {
                TextField("Display Name", text: $displayName)
                    .textContentType(.name)

                TextField("Bio", text: $bio, axis: .vertical)
                    .lineLimit(3...6)

                TextField("Location", text: $location)
                    .textContentType(.addressCity)
            }

            // Privacy Section
            Section("Privacy") {
                Picker("Profile Visibility", selection: $visibility) {
                    ForEach(ProfileVisibility.allCases, id: \.self) { vis in
                        Label(vis.displayName, systemImage: vis.iconName)
                            .tag(vis)
                    }
                }

                Toggle("Show Distance to Others", isOn: $showDistance)

                Toggle("Show Online Status", isOn: $showOnline)
            }

            // Visibility Info Section
            Section {
                HStack(spacing: 12) {
                    Image(systemName: visibility.iconName)
                        .font(.title2)
                        .foregroundStyle(.secondary)
                        .frame(width: 32)

                    VStack(alignment: .leading, spacing: 4) {
                        Text(visibility.displayName)
                            .font(.subheadline.bold())
                        Text(visibility.description)
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                }
                .padding(.vertical, 4)
            }

            // Error Section
            if let errorMessage = errorMessage {
                Section {
                    Text(errorMessage)
                        .foregroundStyle(.red)
                }
            }
        }
        .navigationTitle("Edit Profile")
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .cancellationAction) {
                Button("Cancel") {
                    dismiss()
                }
            }

            ToolbarItem(placement: .confirmationAction) {
                Button("Save") {
                    Task {
                        await saveProfile()
                    }
                }
                .disabled(isSaving || displayName.isEmpty)
            }
        }
        .disabled(isSaving)
        .overlay {
            if isSaving {
                ProgressView()
                    .scaleEffect(1.5)
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
                    .background(.ultraThinMaterial)
            }
        }
        .onAppear {
            loadCurrentProfile()
        }
    }

    private func loadCurrentProfile() {
        guard let profile = profileService.myProfile else { return }
        displayName = profile.displayName ?? ""
        bio = profile.bio ?? ""
        location = profile.location ?? ""
        visibility = profile.visibility
        showDistance = profile.showDistance
        showOnline = profile.showOnline
    }

    private func saveProfile() async {
        isSaving = true
        errorMessage = nil

        let request = UpdateProfileRequest(
            displayName: displayName.isEmpty ? nil : displayName,
            bio: bio.isEmpty ? nil : bio,
            avatarUrl: nil,
            location: location.isEmpty ? nil : location,
            locationLat: nil,
            locationLng: nil,
            visibility: visibility,
            showDistance: showDistance,
            showOnline: showOnline
        )

        do {
            _ = try await profileService.updateProfile(request)
            dismiss()
        } catch {
            errorMessage = error.localizedDescription
        }

        isSaving = false
    }
}

#Preview {
    NavigationStack {
        ProfileEditView()
            .environment(ProfileService.shared)
    }
}
