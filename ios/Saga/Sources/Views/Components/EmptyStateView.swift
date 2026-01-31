import SwiftUI

/// Reusable empty state view
struct EmptyStateView: View {
    let title: String
    let systemImage: String
    var description: String? = nil
    var actionTitle: String? = nil
    var action: (() -> Void)? = nil

    var body: some View {
        ContentUnavailableView {
            Label(title, systemImage: systemImage)
        } description: {
            if let description = description {
                Text(description)
            }
        } actions: {
            if let actionTitle = actionTitle, let action = action {
                Button(actionTitle, action: action)
                    .buttonStyle(.borderedProminent)
            }
        }
    }
}

// MARK: - Common Empty States

extension EmptyStateView {
    static var noGuilds: EmptyStateView {
        EmptyStateView(
            title: "No Guilds",
            systemImage: "person.3.fill",
            description: "Create a guild to start organizing your community"
        )
    }

    static var noEvents: EmptyStateView {
        EmptyStateView(
            title: "No Events",
            systemImage: "calendar",
            description: "Create an event to bring people together"
        )
    }

    static var noPeople: EmptyStateView {
        EmptyStateView(
            title: "No People",
            systemImage: "person.fill",
            description: "Add people to track your relationships"
        )
    }

    static var noResults: EmptyStateView {
        EmptyStateView(
            title: "No Results",
            systemImage: "magnifyingglass",
            description: "Try adjusting your search or filters"
        )
    }

    static var noConnection: EmptyStateView {
        EmptyStateView(
            title: "No Connection",
            systemImage: "wifi.slash",
            description: "Check your internet connection and try again"
        )
    }
}

// MARK: - Illustration Empty State

struct IllustratedEmptyState: View {
    let title: String
    let description: String
    let illustration: String
    var actionTitle: String? = nil
    var action: (() -> Void)? = nil

    var body: some View {
        VStack(spacing: 24) {
            Image(systemName: illustration)
                .font(.system(size: 60))
                .foregroundStyle(.secondary)

            VStack(spacing: 8) {
                Text(title)
                    .font(.title2.bold())

                Text(description)
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
                    .multilineTextAlignment(.center)
                    .padding(.horizontal)
            }

            if let actionTitle = actionTitle, let action = action {
                Button(actionTitle, action: action)
                    .buttonStyle(.borderedProminent)
            }
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .padding()
    }
}

#Preview {
    VStack {
        EmptyStateView.noGuilds

        IllustratedEmptyState(
            title: "Welcome to Saga",
            description: "Start by creating your first guild to organize your community",
            illustration: "sparkles",
            actionTitle: "Create Guild"
        ) {
            // Action
        }
    }
}
