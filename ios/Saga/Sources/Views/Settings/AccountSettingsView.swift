import SwiftUI

/// Account settings and management
struct AccountSettingsView: View {
    @Environment(AuthService.self) private var authService
    @Environment(\.dismiss) private var dismiss

    @State private var showingDeleteConfirmation = false
    @State private var showingLogoutConfirmation = false
    @State private var isDeleting = false
    @State private var error: Error?

    var body: some View {
        List {
            profileSection
            linkedAccountsSection
            passkeysSection
            signOutSection
            deleteAccountSection
        }
        .navigationTitle("Account")
        .navigationBarTitleDisplayMode(.inline)
        .alert("Sign Out", isPresented: $showingLogoutConfirmation) {
            Button("Cancel", role: .cancel) { }
            Button("Sign Out", role: .destructive) {
                Task {
                    try? await authService.logout()
                }
            }
        } message: {
            Text("Are you sure you want to sign out?")
        }
        .alert("Delete Account", isPresented: $showingDeleteConfirmation) {
            Button("Cancel", role: .cancel) { }
            Button("Delete", role: .destructive) {
                Task { await deleteAccount() }
            }
        } message: {
            Text("This will permanently delete your account and all associated data. This action cannot be undone.")
        }
        .alert("Error", isPresented: .constant(error != nil)) {
            Button("OK") { error = nil }
        } message: {
            if let error {
                Text(error.localizedDescription)
            }
        }
    }

    // MARK: - Profile Section

    @ViewBuilder
    private var profileSection: some View {
        Section {
            if let user = authService.currentUser {
                HStack(spacing: 16) {
                    ZStack {
                        Circle()
                            .fill(.blue.gradient)
                            .frame(width: 60, height: 60)
                        Text(user.initials)
                            .font(.title2.bold())
                            .foregroundStyle(.white)
                    }

                    VStack(alignment: .leading, spacing: 4) {
                        Text(user.displayName)
                            .font(.headline)
                        if let email = user.email {
                            Text(email)
                                .font(.subheadline)
                                .foregroundStyle(.secondary)
                        }
                        if let createdAt = user.createdAt {
                            Text("Member since \(createdAt.formatted(date: .abbreviated, time: .omitted))")
                                .font(.caption)
                                .foregroundStyle(.secondary)
                        }
                    }
                }
                .padding(.vertical, 8)
            }
        }
    }

    // MARK: - Linked Accounts Section

    @ViewBuilder
    private var linkedAccountsSection: some View {
        Section("Linked Accounts") {
            ForEach(authService.identities) { identity in
                identityRow(identity)
            }

            Button {
                // TODO: Link Google account
            } label: {
                Label("Link Google Account", systemImage: "plus.circle")
            }

            Button {
                // TODO: Link Apple account
            } label: {
                Label("Link Apple Account", systemImage: "plus.circle")
            }
        }
    }

    private func identityRow(_ identity: Identity) -> some View {
        HStack {
            Image(systemName: identity.provider.iconName)
                .foregroundStyle(.secondary)
                .frame(width: 24)
            Text(identity.provider.displayName)
            Spacer()
            if identity.verified {
                Image(systemName: "checkmark.circle.fill")
                    .foregroundStyle(.green)
            }
        }
    }

    // MARK: - Passkeys Section

    @ViewBuilder
    private var passkeysSection: some View {
        Section("Passkeys") {
            ForEach(authService.passkeys) { passkey in
                passkeyRow(passkey)
            }

            NavigationLink {
                PasskeySetupView()
            } label: {
                Label("Add Passkey", systemImage: "plus.circle")
            }
        }
    }

    private func passkeyRow(_ passkey: Passkey) -> some View {
        HStack {
            Image(systemName: "person.badge.key.fill")
                .foregroundStyle(.blue)
                .frame(width: 24)
            VStack(alignment: .leading) {
                Text(passkey.name)
                if let createdAt = passkey.createdAt {
                    Text("Added \(createdAt.formatted(date: .abbreviated, time: .omitted))")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }
            }
            Spacer()
            Image(systemName: "checkmark.circle.fill")
                .foregroundStyle(.green)
        }
    }

    // MARK: - Sign Out Section

    @ViewBuilder
    private var signOutSection: some View {
        Section {
            Button {
                showingLogoutConfirmation = true
            } label: {
                HStack {
                    Spacer()
                    Text("Sign Out")
                    Spacer()
                }
            }
        }
    }

    // MARK: - Delete Account Section

    @ViewBuilder
    private var deleteAccountSection: some View {
        Section {
            Button(role: .destructive) {
                showingDeleteConfirmation = true
            } label: {
                HStack {
                    Spacer()
                    if isDeleting {
                        ProgressView()
                    } else {
                        Text("Delete Account")
                    }
                    Spacer()
                }
            }
            .disabled(isDeleting)
        } footer: {
            Text("Deleting your account will permanently remove all your data. This action cannot be undone.")
        }
    }

    // MARK: - Actions

    private func deleteAccount() async {
        isDeleting = true
        // TODO: Implement account deletion
        // do {
        //     try await authService.deleteAccount()
        // } catch {
        //     self.error = error
        // }
        isDeleting = false
    }
}

#Preview {
    NavigationStack {
        AccountSettingsView()
            .environment(AuthService.shared)
    }
}
