import SwiftUI

/// Sheet for reporting a user or content
struct ReportUserSheet: View {
    let targetType: ReportTargetType
    let targetId: String
    let targetName: String?

    @Environment(\.dismiss) private var dismiss

    @State private var selectedReason: ReportReason?
    @State private var details = ""
    @State private var isSubmitting = false
    @State private var error: Error?
    @State private var showingConfirmation = false

    private let apiClient = APIClient.shared

    var body: some View {
        NavigationStack {
            Form {
                // Target info
                Section {
                    HStack {
                        Image(systemName: targetType.iconName)
                            .foregroundStyle(.secondary)
                        VStack(alignment: .leading) {
                            Text("Reporting: \(targetType.displayName)")
                                .font(.subheadline)
                            if let name = targetName {
                                Text(name)
                                    .font(.caption)
                                    .foregroundStyle(.secondary)
                            }
                        }
                    }
                }

                // Reason selection
                Section {
                    ForEach(ReportReason.allCases, id: \.self) { reason in
                        Button {
                            selectedReason = reason
                        } label: {
                            HStack {
                                Image(systemName: reason.iconName)
                                    .foregroundStyle(.orange)
                                    .frame(width: 24)

                                VStack(alignment: .leading) {
                                    Text(reason.displayName)
                                        .foregroundStyle(.primary)
                                    Text(reason.description)
                                        .font(.caption)
                                        .foregroundStyle(.secondary)
                                }

                                Spacer()

                                if selectedReason == reason {
                                    Image(systemName: "checkmark.circle.fill")
                                        .foregroundStyle(.blue)
                                }
                            }
                        }
                    }
                } header: {
                    Text("Reason for Report")
                } footer: {
                    Text("Select the reason that best describes the issue")
                }

                // Additional details
                Section {
                    TextField("Provide additional details (optional)", text: $details, axis: .vertical)
                        .lineLimit(4...8)
                } header: {
                    Text("Additional Details")
                } footer: {
                    Text("Help us understand the situation better")
                }

                // Warning
                Section {
                    HStack {
                        Image(systemName: "exclamationmark.triangle.fill")
                            .foregroundStyle(.orange)
                        Text("False reports may result in action against your account.")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                }
            }
            .navigationTitle("Report")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") {
                        dismiss()
                    }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Submit") {
                        Task { await submitReport() }
                    }
                    .disabled(selectedReason == nil || isSubmitting)
                }
            }
            .alert("Report Submitted", isPresented: $showingConfirmation) {
                Button("OK") {
                    dismiss()
                }
            } message: {
                Text("Thank you for your report. We'll review it and take appropriate action.")
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

    private func submitReport() async {
        guard let reason = selectedReason else { return }

        isSubmitting = true

        let request = CreateReportRequest(
            targetType: targetType,
            targetId: targetId,
            reason: reason,
            details: details.isEmpty ? nil : details
        )

        do {
            _ = try await apiClient.submitReport(request)
            showingConfirmation = true
        } catch {
            self.error = error
        }

        isSubmitting = false
    }
}

#Preview {
    ReportUserSheet(
        targetType: .user,
        targetId: "user:123",
        targetName: "John Doe"
    )
}
