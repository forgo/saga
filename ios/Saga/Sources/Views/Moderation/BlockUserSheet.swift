import SwiftUI

/// Sheet for blocking a user
struct BlockUserSheet: View {
    let userId: String
    let userName: String

    @Environment(\.dismiss) private var dismiss

    @State private var reason = ""
    @State private var isBlocking = false
    @State private var error: Error?
    @State private var showingConfirmation = false

    private let apiClient = APIClient.shared

    var body: some View {
        NavigationStack {
            Form {
                // User info
                Section {
                    HStack {
                        Image(systemName: "person.slash.fill")
                            .foregroundStyle(.red)
                        Text("Block \(userName)")
                            .font(.headline)
                    }
                }

                // What blocking does
                Section {
                    VStack(alignment: .leading, spacing: 8) {
                        blockingEffect("They won't be able to see your profile", icon: "eye.slash")
                        blockingEffect("They can't message you", icon: "message.badge.slash")
                        blockingEffect("They won't appear in your discovery", icon: "magnifyingglass.circle")
                        blockingEffect("You won't see their content", icon: "xmark.circle")
                    }
                } header: {
                    Text("When you block someone")
                }

                // Reason
                Section {
                    TextField("Reason (optional)", text: $reason, axis: .vertical)
                        .lineLimit(3...6)
                } header: {
                    Text("Reason")
                } footer: {
                    Text("This is for your reference only")
                }

                // Warning
                Section {
                    HStack {
                        Image(systemName: "info.circle.fill")
                            .foregroundStyle(.blue)
                        Text("You can unblock this person at any time from your settings.")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                }
            }
            .navigationTitle("Block User")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") {
                        dismiss()
                    }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Block") {
                        Task { await blockUser() }
                    }
                    .foregroundStyle(.red)
                    .disabled(isBlocking)
                }
            }
            .alert("User Blocked", isPresented: $showingConfirmation) {
                Button("OK") {
                    dismiss()
                }
            } message: {
                Text("\(userName) has been blocked. They will no longer be able to contact you.")
            }
            .alert("Error", isPresented: .constant(error != nil)) {
                Button("OK") { error = nil }
            } message: {
                if let error = error {
                    Text(error.localizedDescription)
                }
            }
        }
    }

    @ViewBuilder
    private func blockingEffect(_ text: String, icon: String) -> some View {
        HStack {
            Image(systemName: icon)
                .foregroundStyle(.secondary)
                .frame(width: 24)
            Text(text)
                .font(.subheadline)
        }
    }

    private func blockUser() async {
        isBlocking = true

        let request = CreateBlockRequest(
            blockedId: userId,
            reason: reason.isEmpty ? nil : reason
        )

        do {
            _ = try await apiClient.blockUser(request)
            showingConfirmation = true
        } catch {
            self.error = error
        }

        isBlocking = false
    }
}

#Preview {
    BlockUserSheet(userId: "user:123", userName: "John Doe")
}
