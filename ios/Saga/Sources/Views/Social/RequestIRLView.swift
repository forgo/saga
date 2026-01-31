import SwiftUI

/// View for requesting IRL confirmation from another user
struct RequestIRLView: View {
    let userId: String

    @Environment(ProfileService.self) private var profileService
    @Environment(\.dismiss) private var dismiss

    @State private var context: IRLContext = .spontaneous
    @State private var location: String = ""

    @State private var isSending = false
    @State private var errorMessage: String?

    var body: some View {
        Form {
            // Context Section
            Section {
                ForEach(IRLContext.allCases, id: \.self) { ctx in
                    Button {
                        context = ctx
                    } label: {
                        HStack {
                            Text(ctx.displayName)
                                .foregroundStyle(.primary)
                            Spacer()
                            if context == ctx {
                                Image(systemName: "checkmark")
                                    .foregroundStyle(.blue)
                            }
                        }
                    }
                    .buttonStyle(.plain)
                }
            } header: {
                Text("How did you meet?")
            }

            // Location Section
            Section("Where did you meet? (Optional)") {
                TextField("Location", text: $location)
            }

            // Info Section
            Section {
                HStack(spacing: 12) {
                    Image(systemName: "info.circle.fill")
                        .foregroundStyle(.blue)
                        .font(.title2)

                    VStack(alignment: .leading, spacing: 4) {
                        Text("IRL Confirmation")
                            .font(.subheadline.bold())
                        Text("This will send a request to confirm you've met in real life. Once confirmed, it strengthens trust between you.")
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
        .navigationTitle("Request IRL Confirmation")
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

        let request = RequestIRLRequest(
            targetId: userId,
            context: context,
            contextId: nil, // Could link to specific event
            location: location.isEmpty ? nil : location
        )

        do {
            _ = try await profileService.requestIRLConfirmation(request)
            dismiss()
        } catch {
            errorMessage = error.localizedDescription
        }

        isSending = false
    }
}

#Preview {
    NavigationStack {
        RequestIRLView(userId: "test-user")
            .environment(ProfileService.shared)
    }
}
