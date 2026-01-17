import SwiftUI

/// About the app view
struct AboutView: View {
    private var appVersion: String {
        Bundle.main.infoDictionary?["CFBundleShortVersionString"] as? String ?? "1.0"
    }

    private var buildNumber: String {
        Bundle.main.infoDictionary?["CFBundleVersion"] as? String ?? "1"
    }

    var body: some View {
        List {
            // App info
            Section {
                VStack(spacing: 16) {
                    // App icon
                    Image(systemName: "sparkles")
                        .font(.system(size: 60))
                        .foregroundStyle(.purple.gradient)

                    Text("Saga")
                        .font(.title.bold())

                    Text("Version \(appVersion) (\(buildNumber))")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)

                    Text("Connect with your community")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                        .multilineTextAlignment(.center)
                }
                .frame(maxWidth: .infinity)
                .padding(.vertical, 24)
            }
            .listRowBackground(Color.clear)

            // Links
            Section {
                Link(destination: URL(string: "https://example.com")!) {
                    HStack {
                        Text("Website")
                        Spacer()
                        Image(systemName: "arrow.up.right")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                }

                Link(destination: URL(string: "https://example.com/privacy")!) {
                    HStack {
                        Text("Privacy Policy")
                        Spacer()
                        Image(systemName: "arrow.up.right")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                }

                Link(destination: URL(string: "https://example.com/terms")!) {
                    HStack {
                        Text("Terms of Service")
                        Spacer()
                        Image(systemName: "arrow.up.right")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                }
            }

            // Acknowledgements
            Section("Acknowledgements") {
                NavigationLink("Open Source Licenses") {
                    LicensesView()
                }
            }

            // Copyright
            Section {
                Text("© 2024 Saga. All rights reserved.")
                    .font(.caption)
                    .foregroundStyle(.secondary)
                    .frame(maxWidth: .infinity)
            }
            .listRowBackground(Color.clear)
        }
        .navigationTitle("About")
        .navigationBarTitleDisplayMode(.inline)
    }
}

// MARK: - Licenses View

struct LicensesView: View {
    private let licenses: [(name: String, license: String)] = [
        ("KeychainAccess", "MIT License\n\nCopyright (c) 2014 kishikawa katsumi"),
        ("Swift", "Apache License 2.0\n\nCopyright (c) Apple Inc."),
    ]

    var body: some View {
        List {
            ForEach(licenses, id: \.name) { item in
                DisclosureGroup {
                    Text(item.license)
                        .font(.caption.monospaced())
                        .foregroundStyle(.secondary)
                        .padding(.vertical, 4)
                } label: {
                    Text(item.name)
                }
            }
        }
        .navigationTitle("Licenses")
        .navigationBarTitleDisplayMode(.inline)
    }
}

#Preview {
    NavigationStack {
        AboutView()
    }
}
