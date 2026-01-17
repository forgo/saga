import SwiftUI

/// View listing all questionnaires
struct QuestionnaireListView: View {
    @Environment(DiscoveryService.self) private var discoveryService

    var body: some View {
        List {
            // Progress Overview
            Section {
                VStack(spacing: 12) {
                    // Progress Circle
                    ZStack {
                        Circle()
                            .stroke(.gray.opacity(0.3), lineWidth: 8)
                            .frame(width: 80, height: 80)
                        Circle()
                            .trim(from: 0, to: discoveryService.overallProgress)
                            .stroke(.blue, style: StrokeStyle(lineWidth: 8, lineCap: .round))
                            .frame(width: 80, height: 80)
                            .rotationEffect(.degrees(-90))
                        Text("\(Int(discoveryService.overallProgress * 100))%")
                            .font(.title2.bold())
                    }

                    Text("Overall Completion")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)

                    Text("Complete questionnaires to improve your compatibility matching")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                        .multilineTextAlignment(.center)
                }
                .frame(maxWidth: .infinity)
                .padding(.vertical)
            }

            // Questionnaires by Category
            let grouped = Dictionary(grouping: discoveryService.questionnaires) { $0.category }

            ForEach(QuestionnaireCategory.allCases, id: \.self) { category in
                if let questionnaires = grouped[category], !questionnaires.isEmpty {
                    Section {
                        ForEach(questionnaires) { questionnaire in
                            NavigationLink {
                                QuestionnaireView(questionnaireId: questionnaire.id)
                            } label: {
                                QuestionnaireRow(
                                    questionnaire: questionnaire,
                                    progress: discoveryService.progress(for: questionnaire.id)
                                )
                            }
                        }
                    } header: {
                        Label(category.displayName, systemImage: category.iconName)
                    }
                }
            }
        }
        .navigationTitle("Questionnaires")
        .refreshable {
            await discoveryService.loadQuestionnaires()
        }
        .task {
            await discoveryService.loadQuestionnaires()
        }
    }
}

// MARK: - Questionnaire Row

struct QuestionnaireRow: View {
    let questionnaire: Questionnaire
    let progress: QuestionnaireProgress?

    var body: some View {
        HStack(spacing: 12) {
            // Status indicator
            ZStack {
                Circle()
                    .fill(statusColor.opacity(0.2))
                    .frame(width: 40, height: 40)
                Image(systemName: statusIcon)
                    .foregroundStyle(statusColor)
            }

            VStack(alignment: .leading, spacing: 4) {
                Text(questionnaire.name)
                    .font(.headline)

                if let description = questionnaire.description {
                    Text(description)
                        .font(.caption)
                        .foregroundStyle(.secondary)
                        .lineLimit(2)
                }

                // Progress bar
                if let progress = progress {
                    HStack(spacing: 8) {
                        ProgressView(value: progress.progressPercentage)
                            .tint(statusColor)

                        Text("\(progress.answeredQuestions)/\(progress.totalQuestions)")
                            .font(.caption2)
                            .foregroundStyle(.secondary)
                    }
                }
            }

            if questionnaire.isRequired {
                Text("Required")
                    .font(.caption2.bold())
                    .padding(.horizontal, 6)
                    .padding(.vertical, 2)
                    .background(Color.red.opacity(0.1))
                    .foregroundStyle(.red)
                    .clipShape(Capsule())
            }
        }
        .padding(.vertical, 4)
    }

    private var statusColor: Color {
        guard let progress = progress else { return .gray }
        if progress.isComplete { return .green }
        if progress.answeredQuestions > 0 { return .orange }
        return .gray
    }

    private var statusIcon: String {
        guard let progress = progress else { return "circle" }
        if progress.isComplete { return "checkmark.circle.fill" }
        if progress.answeredQuestions > 0 { return "circle.lefthalf.filled" }
        return "circle"
    }
}

#Preview {
    NavigationStack {
        QuestionnaireListView()
            .environment(DiscoveryService.shared)
    }
}
