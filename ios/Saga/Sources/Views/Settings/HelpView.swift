import SwiftUI

/// Help and support view
struct HelpView: View {
    @State private var searchText = ""

    private let faqItems: [(question: String, answer: String)] = [
        ("How do I create a guild?", "Tap the + button on the Guilds tab to create a new guild. Give it a name, description, and choose an icon and color."),
        ("What are adventures?", "Adventures are group activities with admission control. You can create open events or require approval, invitations, or specific criteria for joining."),
        ("How does the matching pool work?", "Matching pools automatically pair members together at regular intervals (daily, weekly, etc.) for one-on-one connections."),
        ("What is Resonance?", "Resonance is a gamification system that rewards community participation. Earn points for attending events, making connections, and being an active member."),
        ("How do I report someone?", "On any user's profile, tap the ... menu and select 'Report'. Choose a reason and provide details about the issue."),
        ("Can I delete my account?", "Yes, go to Settings > Account and scroll to the bottom to find the Delete Account option. This action is permanent."),
        ("How do passkeys work?", "Passkeys are a secure, passwordless way to sign in using your device's biometrics (Face ID, Touch ID). Go to Settings > Security to set one up."),
        ("What trust levels are there?", "There are three trust levels: Basic (default), Elevated (for closer connections), and Full (for your closest friends)."),
    ]

    var filteredFAQ: [(question: String, answer: String)] {
        if searchText.isEmpty {
            return faqItems
        }
        return faqItems.filter {
            $0.question.localizedCaseInsensitiveContains(searchText) ||
            $0.answer.localizedCaseInsensitiveContains(searchText)
        }
    }

    var body: some View {
        List {
            // Contact support
            Section {
                Link(destination: URL(string: "mailto:support@example.com")!) {
                    HStack {
                        Image(systemName: "envelope.fill")
                            .foregroundStyle(.blue)
                            .frame(width: 28)
                        VStack(alignment: .leading) {
                            Text("Email Support")
                            Text("support@example.com")
                                .font(.caption)
                                .foregroundStyle(.secondary)
                        }
                        Spacer()
                        Image(systemName: "arrow.up.right")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                }

                Link(destination: URL(string: "https://example.com/help")!) {
                    HStack {
                        Image(systemName: "globe")
                            .foregroundStyle(.green)
                            .frame(width: 28)
                        Text("Help Center")
                        Spacer()
                        Image(systemName: "arrow.up.right")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                }
            } header: {
                Text("Contact Us")
            }

            // FAQ
            Section("Frequently Asked Questions") {
                ForEach(filteredFAQ, id: \.question) { item in
                    DisclosureGroup {
                        Text(item.answer)
                            .font(.subheadline)
                            .foregroundStyle(.secondary)
                            .padding(.vertical, 4)
                    } label: {
                        Text(item.question)
                            .font(.subheadline)
                    }
                }
            }

            // Community
            Section("Community") {
                Link(destination: URL(string: "https://example.com/community")!) {
                    HStack {
                        Image(systemName: "bubble.left.and.bubble.right.fill")
                            .foregroundStyle(.purple)
                            .frame(width: 28)
                        Text("Community Forum")
                        Spacer()
                        Image(systemName: "arrow.up.right")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                }

                Link(destination: URL(string: "https://twitter.com/sagaapp")!) {
                    HStack {
                        Image(systemName: "at")
                            .foregroundStyle(.blue)
                            .frame(width: 28)
                        Text("Follow us on Twitter")
                        Spacer()
                        Image(systemName: "arrow.up.right")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                }
            }
        }
        .navigationTitle("Help & Support")
        .navigationBarTitleDisplayMode(.inline)
        .searchable(text: $searchText, prompt: "Search help")
    }
}

#Preview {
    NavigationStack {
        HelpView()
    }
}
