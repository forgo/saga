import SwiftUI

/// View for managing passkeys
struct PasskeySetupView: View {
    @Environment(AuthService.self) private var authService
    @Environment(PasskeyService.self) private var passkeyService

    @State private var isRegistering = false
    @State private var errorMessage: String?
    @State private var isShowingError = false
    @State private var showingDeleteAlert = false
    @State private var passkeyToDelete: Passkey?

    var body: some View {
        List {
            // Info section
            Section {
                VStack(alignment: .leading, spacing: 12) {
                    HStack {
                        Image(systemName: "person.badge.key.fill")
                            .font(.title)
                            .foregroundStyle(.blue)
                        VStack(alignment: .leading) {
                            Text("Passkeys")
                                .font(.headline)
                            Text("Sign in securely without a password")
                                .font(.subheadline)
                                .foregroundStyle(.secondary)
                        }
                    }

                    Text("Passkeys use Face ID, Touch ID, or your device passcode for secure authentication. They're phishing-resistant and never leave your device.")
                        .font(.callout)
                        .foregroundStyle(.secondary)
                }
                .padding(.vertical, 8)
            }

            // Existing passkeys
            if !authService.passkeys.isEmpty {
                Section("Your Passkeys") {
                    ForEach(authService.passkeys) { passkey in
                        PasskeyRow(passkey: passkey) {
                            passkeyToDelete = passkey
                            showingDeleteAlert = true
                        }
                    }
                }
            }

            // Register new passkey
            Section {
                Button {
                    Task { await registerPasskey() }
                } label: {
                    HStack {
                        if isRegistering {
                            ProgressView()
                        } else {
                            Image(systemName: "plus.circle.fill")
                                .foregroundStyle(.blue)
                        }
                        Text(authService.passkeys.isEmpty ? "Set Up Passkey" : "Add Another Passkey")
                    }
                }
                .disabled(isRegistering)
            } footer: {
                Text("Each device can have its own passkey for convenient access.")
            }
        }
        .navigationTitle("Passkeys")
        .navigationBarTitleDisplayMode(.inline)
        .alert("Error", isPresented: $isShowingError) {
            Button("OK", role: .cancel) { }
        } message: {
            Text(errorMessage ?? "An error occurred")
        }
        .alert("Delete Passkey", isPresented: $showingDeleteAlert) {
            Button("Cancel", role: .cancel) {
                passkeyToDelete = nil
            }
            Button("Delete", role: .destructive) {
                if let passkey = passkeyToDelete {
                    Task { await deletePasskey(passkey) }
                }
            }
        } message: {
            Text("Are you sure you want to delete this passkey? You won't be able to sign in with it anymore.")
        }
    }

    // MARK: - Register Passkey

    private func registerPasskey() async {
        isRegistering = true
        defer { isRegistering = false }

        do {
            // Get challenge from server
            let challengeResponse = try await APIClient.shared.beginPasskeyRegistration()
            guard let challengeData = Data(base64URLEncoded: challengeResponse.data.challenge) else {
                throw PasskeyError.invalidCredential
            }

            // Create user ID from current user
            guard let user = authService.currentUser else {
                throw APIError.unauthorized
            }
            guard let userIdData = user.id.data(using: .utf8) else {
                throw PasskeyError.invalidCredential
            }

            // Perform WebAuthn registration
            let registration = try await passkeyService.registerPasskey(
                challenge: challengeData,
                userId: userIdData,
                userName: user.email ?? user.id,
                userDisplayName: user.displayName
            )

            // Complete registration with API
            _ = try await authService.registerPasskey(credential: registration.toPasskeyCredential())

        } catch PasskeyError.cancelled {
            // User cancelled, ignore
        } catch {
            showError(error)
        }
    }

    // MARK: - Delete Passkey

    private func deletePasskey(_ passkey: Passkey) async {
        do {
            try await authService.deletePasskey(id: passkey.id)
        } catch {
            showError(error)
        }
        passkeyToDelete = nil
    }

    // MARK: - Error Handling

    private func showError(_ error: Error) {
        if let apiError = error as? APIError {
            errorMessage = apiError.userMessage
        } else {
            errorMessage = error.localizedDescription
        }
        isShowingError = true
    }
}

// MARK: - Passkey Row

struct PasskeyRow: View {
    let passkey: Passkey
    let onDelete: () -> Void

    var body: some View {
        HStack {
            Image(systemName: "person.badge.key.fill")
                .foregroundStyle(.secondary)
                .frame(width: 32)

            VStack(alignment: .leading, spacing: 4) {
                Text(passkey.name)
                    .font(.body)

                if let lastUsed = passkey.lastUsedAt {
                    Text("Last used \(lastUsed.formatted(.relative(presentation: .named)))")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                } else if let created = passkey.createdAt {
                    Text("Created \(created.formatted(.relative(presentation: .named)))")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }
            }

            Spacer()

            Button(role: .destructive) {
                onDelete()
            } label: {
                Image(systemName: "trash")
                    .foregroundStyle(.red)
            }
            .buttonStyle(.plain)
        }
        .padding(.vertical, 4)
    }
}

#Preview {
    NavigationStack {
        PasskeySetupView()
            .environment(AuthService.shared)
            .environment(PasskeyService.shared)
    }
}
