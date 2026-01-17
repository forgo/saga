import SwiftUI

/// View for taking a questionnaire
struct QuestionnaireView: View {
    let questionnaireId: String

    @Environment(DiscoveryService.self) private var discoveryService
    @Environment(\.dismiss) private var dismiss

    @State private var currentQuestionIndex = 0
    @State private var responses: [String: SubmitResponseRequest] = [:]
    @State private var isSaving = false
    @State private var errorMessage: String?

    var questionnaire: Questionnaire? {
        discoveryService.currentQuestionnaire
    }

    var questions: [Question] {
        questionnaire?.questions ?? []
    }

    var currentQuestion: Question? {
        guard currentQuestionIndex < questions.count else { return nil }
        return questions[currentQuestionIndex]
    }

    var progress: Double {
        guard !questions.isEmpty else { return 0 }
        return Double(currentQuestionIndex) / Double(questions.count)
    }

    var body: some View {
        Group {
            if discoveryService.isLoadingQuestionnaires {
                ProgressView("Loading questionnaire...")
            } else if let questionnaire = questionnaire {
                questionnaireContent(questionnaire)
            } else {
                ContentUnavailableView {
                    Label("Questionnaire Not Found", systemImage: "questionmark.circle")
                } description: {
                    Text("Unable to load this questionnaire")
                }
            }
        }
        .navigationTitle(questionnaire?.name ?? "Questionnaire")
        .navigationBarTitleDisplayMode(.inline)
        .task {
            await discoveryService.loadQuestionnaire(questionnaireId: questionnaireId)
            loadExistingResponses()
        }
    }

    @ViewBuilder
    private func questionnaireContent(_ questionnaire: Questionnaire) -> some View {
        VStack(spacing: 0) {
            // Progress Bar
            ProgressView(value: progress)
                .tint(.blue)
                .padding(.horizontal)
                .padding(.top)

            Text("\(currentQuestionIndex + 1) of \(questions.count)")
                .font(.caption)
                .foregroundStyle(.secondary)
                .padding(.top, 4)

            // Question Content
            if let question = currentQuestion {
                ScrollView {
                    VStack(alignment: .leading, spacing: 24) {
                        Text(question.text)
                            .font(.title3.bold())
                            .padding(.top, 24)

                        questionInput(for: question)
                    }
                    .padding(.horizontal)
                }
            } else {
                // Completion view
                completionView
            }

            // Navigation Buttons
            HStack(spacing: 16) {
                Button {
                    withAnimation {
                        currentQuestionIndex -= 1
                    }
                } label: {
                    Label("Previous", systemImage: "chevron.left")
                }
                .disabled(currentQuestionIndex == 0)

                Spacer()

                if currentQuestionIndex < questions.count {
                    Button {
                        withAnimation {
                            currentQuestionIndex += 1
                        }
                    } label: {
                        Label("Next", systemImage: "chevron.right")
                            .labelStyle(.trailingIcon)
                    }
                } else {
                    Button {
                        Task {
                            await submitAllResponses()
                        }
                    } label: {
                        if isSaving {
                            ProgressView()
                        } else {
                            Text("Submit")
                                .bold()
                        }
                    }
                    .buttonStyle(.borderedProminent)
                    .disabled(isSaving)
                }
            }
            .padding()
            .background(.bar)
        }
    }

    @ViewBuilder
    private func questionInput(for question: Question) -> some View {
        switch question.questionType {
        case .multipleChoice:
            multipleChoiceInput(for: question)
        case .scale:
            scaleInput(for: question)
        case .text:
            textInput(for: question)
        case .multiSelect:
            multiSelectInput(for: question)
        }
    }

    @ViewBuilder
    private func multipleChoiceInput(for question: Question) -> some View {
        VStack(spacing: 12) {
            ForEach(question.options ?? []) { option in
                Button {
                    responses[question.id] = SubmitResponseRequest(
                        questionId: question.id,
                        optionId: option.id,
                        scaleValue: nil,
                        textValue: nil,
                        selectedOptions: nil
                    )
                } label: {
                    HStack {
                        Text(option.text)
                            .foregroundStyle(.primary)
                        Spacer()
                        if responses[question.id]?.optionId == option.id {
                            Image(systemName: "checkmark.circle.fill")
                                .foregroundStyle(.blue)
                        } else {
                            Image(systemName: "circle")
                                .foregroundStyle(.secondary)
                        }
                    }
                    .padding()
                    .background(
                        RoundedRectangle(cornerRadius: 12)
                            .fill(responses[question.id]?.optionId == option.id ? Color.blue.opacity(0.1) : Color(.systemGray6))
                    )
                }
                .buttonStyle(.plain)
            }
        }
    }

    @ViewBuilder
    private func scaleInput(for question: Question) -> some View {
        let minValue = question.minValue ?? 1
        let maxValue = question.maxValue ?? 5
        let currentValue = responses[question.id]?.scaleValue ?? minValue

        VStack(spacing: 16) {
            HStack {
                Text("\(minValue)")
                    .font(.caption)
                    .foregroundStyle(.secondary)
                Spacer()
                Text("\(maxValue)")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }

            HStack(spacing: 8) {
                ForEach(minValue...maxValue, id: \.self) { value in
                    Button {
                        responses[question.id] = SubmitResponseRequest(
                            questionId: question.id,
                            optionId: nil,
                            scaleValue: value,
                            textValue: nil,
                            selectedOptions: nil
                        )
                    } label: {
                        Text("\(value)")
                            .font(.headline)
                            .frame(width: 44, height: 44)
                            .background(
                                Circle()
                                    .fill(currentValue == value ? Color.blue : Color(.systemGray5))
                            )
                            .foregroundStyle(currentValue == value ? .white : .primary)
                    }
                    .buttonStyle(.plain)
                }
            }
        }
    }

    @ViewBuilder
    private func textInput(for question: Question) -> some View {
        let binding = Binding<String>(
            get: { responses[question.id]?.textValue ?? "" },
            set: { newValue in
                responses[question.id] = SubmitResponseRequest(
                    questionId: question.id,
                    optionId: nil,
                    scaleValue: nil,
                    textValue: newValue.isEmpty ? nil : newValue,
                    selectedOptions: nil
                )
            }
        )

        TextField("Your answer...", text: binding, axis: .vertical)
            .lineLimit(3...6)
            .textFieldStyle(.roundedBorder)
    }

    @ViewBuilder
    private func multiSelectInput(for question: Question) -> some View {
        let selectedOptions = responses[question.id]?.selectedOptions ?? []

        VStack(spacing: 12) {
            ForEach(question.options ?? []) { option in
                Button {
                    var current = selectedOptions
                    if current.contains(option.id) {
                        current.removeAll { $0 == option.id }
                    } else {
                        current.append(option.id)
                    }
                    responses[question.id] = SubmitResponseRequest(
                        questionId: question.id,
                        optionId: nil,
                        scaleValue: nil,
                        textValue: nil,
                        selectedOptions: current.isEmpty ? nil : current
                    )
                } label: {
                    HStack {
                        Text(option.text)
                            .foregroundStyle(.primary)
                        Spacer()
                        Image(systemName: selectedOptions.contains(option.id) ? "checkmark.square.fill" : "square")
                            .foregroundStyle(selectedOptions.contains(option.id) ? .blue : .secondary)
                    }
                    .padding()
                    .background(
                        RoundedRectangle(cornerRadius: 12)
                            .fill(selectedOptions.contains(option.id) ? Color.blue.opacity(0.1) : Color(.systemGray6))
                    )
                }
                .buttonStyle(.plain)
            }
        }
    }

    @ViewBuilder
    private var completionView: some View {
        VStack(spacing: 24) {
            Spacer()

            Image(systemName: "checkmark.circle.fill")
                .font(.system(size: 80))
                .foregroundStyle(.green)

            Text("All Done!")
                .font(.title.bold())

            Text("You've answered all \(questions.count) questions. Submit your responses to improve your compatibility matching.")
                .font(.body)
                .foregroundStyle(.secondary)
                .multilineTextAlignment(.center)
                .padding(.horizontal)

            if let errorMessage = errorMessage {
                Text(errorMessage)
                    .foregroundStyle(.red)
                    .font(.caption)
            }

            Spacer()
        }
    }

    private func loadExistingResponses() {
        for response in discoveryService.currentResponses {
            responses[response.questionId] = SubmitResponseRequest(
                questionId: response.questionId,
                optionId: response.optionId,
                scaleValue: response.scaleValue,
                textValue: response.textValue,
                selectedOptions: response.selectedOptions
            )
        }

        // Find first unanswered question
        if let firstUnanswered = questions.firstIndex(where: { responses[$0.id] == nil }) {
            currentQuestionIndex = firstUnanswered
        } else if !questions.isEmpty {
            currentQuestionIndex = questions.count // Go to completion
        }
    }

    private func submitAllResponses() async {
        isSaving = true
        errorMessage = nil

        let requestsToSubmit = Array(responses.values)

        do {
            _ = try await discoveryService.submitResponses(requestsToSubmit)
            dismiss()
        } catch {
            errorMessage = error.localizedDescription
        }

        isSaving = false
    }
}

// MARK: - Trailing Icon Label Style

struct TrailingIconLabelStyle: LabelStyle {
    func makeBody(configuration: Configuration) -> some View {
        HStack {
            configuration.title
            configuration.icon
        }
    }
}

extension LabelStyle where Self == TrailingIconLabelStyle {
    static var trailingIcon: TrailingIconLabelStyle { TrailingIconLabelStyle() }
}

#Preview {
    NavigationStack {
        QuestionnaireView(questionnaireId: "test")
            .environment(DiscoveryService.shared)
    }
}
