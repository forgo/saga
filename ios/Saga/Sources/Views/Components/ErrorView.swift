import SwiftUI

/// Reusable error display view
struct ErrorView: View {
    let error: Error
    let retryAction: (() async -> Void)?

    init(_ error: Error, retryAction: (() async -> Void)? = nil) {
        self.error = error
        self.retryAction = retryAction
    }

    private var errorMessage: String {
        if let apiError = error as? APIError {
            return apiError.userMessage
        }
        return error.localizedDescription
    }

    private var errorIcon: String {
        if let apiError = error as? APIError {
            switch apiError {
            case .networkError:
                return "wifi.slash"
            case .unauthorized:
                return "lock.fill"
            case .notFound:
                return "questionmark.circle"
            case .serverError:
                return "exclamationmark.icloud"
            case .rateLimited:
                return "clock.fill"
            case .forbidden:
                return "hand.raised.fill"
            default:
                return "exclamationmark.triangle"
            }
        }
        return "exclamationmark.triangle"
    }

    var body: some View {
        ContentUnavailableView {
            Label("Something went wrong", systemImage: errorIcon)
        } description: {
            Text(errorMessage)
        } actions: {
            if let retryAction = retryAction {
                Button("Try Again") {
                    Task { await retryAction() }
                }
                .buttonStyle(.borderedProminent)
            }
        }
    }
}


// MARK: - Inline Error Banner

struct ErrorBanner: View {
    let message: String
    let onDismiss: () -> Void

    var body: some View {
        HStack {
            Image(systemName: "exclamationmark.triangle.fill")
                .foregroundStyle(.white)

            Text(message)
                .font(.subheadline)
                .foregroundStyle(.white)
                .lineLimit(2)

            Spacer()

            Button {
                onDismiss()
            } label: {
                Image(systemName: "xmark")
                    .foregroundStyle(.white.opacity(0.8))
            }
        }
        .padding()
        .background(.red)
        .clipShape(RoundedRectangle(cornerRadius: 12))
        .padding(.horizontal)
    }
}

// MARK: - Toast Modifier

struct ToastModifier: ViewModifier {
    @Binding var message: String?
    var isError: Bool = false

    func body(content: Content) -> some View {
        ZStack {
            content

            if let message = message {
                VStack {
                    Spacer()

                    HStack {
                        if isError {
                            Image(systemName: "exclamationmark.circle.fill")
                        } else {
                            Image(systemName: "checkmark.circle.fill")
                        }
                        Text(message)
                            .font(.subheadline)
                    }
                    .padding()
                    .background(isError ? .red : .green)
                    .foregroundStyle(.white)
                    .clipShape(Capsule())
                    .shadow(radius: 4)
                    .padding(.bottom, 100)
                }
                .transition(.move(edge: .bottom).combined(with: .opacity))
                .onAppear {
                    DispatchQueue.main.asyncAfter(deadline: .now() + 3) {
                        withAnimation {
                            self.message = nil
                        }
                    }
                }
            }
        }
        .animation(.spring(), value: message)
    }
}

extension View {
    func toast(message: Binding<String?>, isError: Bool = false) -> some View {
        modifier(ToastModifier(message: message, isError: isError))
    }
}

#Preview {
    ErrorView(APIError.unknown) {
        // Retry
    }
}
