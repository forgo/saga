import SwiftUI

/// View for granting trust to another user
struct GrantTrustView: View {
    let userId: String

    @Environment(ProfileService.self) private var profileService
    @Environment(\.dismiss) private var dismiss

    @State private var trustLevel: TrustLevel = .basic
    @State private var selectedPermissions: Set<TrustPermission> = []
    @State private var notes: String = ""

    @State private var isSaving = false
    @State private var errorMessage: String?

    var body: some View {
        Form {
            // Trust Level Section
            Section("Trust Level") {
                ForEach(TrustLevel.allCases, id: \.self) { level in
                    Button {
                        trustLevel = level
                        // Set default permissions for the level
                        selectedPermissions = Set(level.defaultPermissions)
                    } label: {
                        HStack {
                            Image(systemName: level.iconName)
                                .font(.title2)
                                .frame(width: 32)
                                .foregroundStyle(trustLevel == level ? .blue : .secondary)

                            VStack(alignment: .leading, spacing: 2) {
                                Text(level.displayName)
                                    .font(.subheadline.bold())
                                    .foregroundStyle(.primary)
                                Text(level.description)
                                    .font(.caption)
                                    .foregroundStyle(.secondary)
                            }

                            Spacer()

                            if trustLevel == level {
                                Image(systemName: "checkmark.circle.fill")
                                    .foregroundStyle(.blue)
                            }
                        }
                    }
                    .buttonStyle(.plain)
                }
            }

            // Permissions Section
            Section {
                ForEach(TrustPermission.allCases, id: \.self) { permission in
                    Toggle(isOn: Binding(
                        get: { selectedPermissions.contains(permission) },
                        set: { isSelected in
                            if isSelected {
                                selectedPermissions.insert(permission)
                            } else {
                                selectedPermissions.remove(permission)
                            }
                        }
                    )) {
                        HStack {
                            Image(systemName: permission.iconName)
                                .foregroundStyle(.secondary)
                                .frame(width: 24)
                            VStack(alignment: .leading, spacing: 2) {
                                Text(permission.displayName)
                                    .font(.subheadline)
                                Text(permission.description)
                                    .font(.caption)
                                    .foregroundStyle(.secondary)
                            }
                        }
                    }
                }
            } header: {
                Text("Permissions")
            } footer: {
                Text("Choose what this person can see about you")
            }

            // Notes Section
            Section("Notes (Optional)") {
                TextField("Private note about this trust grant...", text: $notes, axis: .vertical)
                    .lineLimit(2...4)
            }

            // Error Section
            if let errorMessage = errorMessage {
                Section {
                    Text(errorMessage)
                        .foregroundStyle(.red)
                }
            }
        }
        .navigationTitle("Grant Trust")
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .cancellationAction) {
                Button("Cancel") {
                    dismiss()
                }
            }

            ToolbarItem(placement: .confirmationAction) {
                Button("Grant") {
                    Task {
                        await grantTrust()
                    }
                }
                .disabled(isSaving)
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
    }

    private func grantTrust() async {
        isSaving = true
        errorMessage = nil

        let request = CreateTrustRequest(
            granteeId: userId,
            trustLevel: trustLevel,
            permissions: selectedPermissions.isEmpty ? nil : Array(selectedPermissions),
            notes: notes.isEmpty ? nil : notes
        )

        do {
            _ = try await profileService.grantTrust(request)
            dismiss()
        } catch {
            errorMessage = error.localizedDescription
        }

        isSaving = false
    }
}

#Preview {
    NavigationStack {
        GrantTrustView(userId: "test-user")
            .environment(ProfileService.shared)
    }
}
