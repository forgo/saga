import SwiftUI

/// Reusable loading state view
struct LoadingView: View {
    var message: String = "Loading..."

    var body: some View {
        VStack(spacing: 16) {
            ProgressView()
                .scaleEffect(1.2)

            Text(message)
                .font(.subheadline)
                .foregroundStyle(.secondary)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }
}

// MARK: - Async Content View

/// A view that handles loading, error, and content states
struct AsyncContentView<T, Content: View>: View {
    let state: AsyncState<T>
    let retryAction: (() async -> Void)?
    @ViewBuilder let content: (T) -> Content

    init(
        state: AsyncState<T>,
        retryAction: (() async -> Void)? = nil,
        @ViewBuilder content: @escaping (T) -> Content
    ) {
        self.state = state
        self.retryAction = retryAction
        self.content = content
    }

    var body: some View {
        switch state {
        case .idle:
            Color.clear
        case .loading:
            LoadingView()
        case .loaded(let data):
            content(data)
        case .error(let error):
            ErrorView(error, retryAction: retryAction)
        }
    }
}

// MARK: - Async State

enum AsyncState<T> {
    case idle
    case loading
    case loaded(T)
    case error(Error)

    var isLoading: Bool {
        if case .loading = self { return true }
        return false
    }

    var data: T? {
        if case .loaded(let data) = self { return data }
        return nil
    }

    var error: Error? {
        if case .error(let error) = self { return error }
        return nil
    }
}

// MARK: - Skeleton Loading

struct SkeletonView: View {
    var body: some View {
        RoundedRectangle(cornerRadius: 4)
            .fill(.gray.opacity(0.2))
            .shimmering()
    }
}

struct ShimmerModifier: ViewModifier {
    @State private var phase: CGFloat = 0

    func body(content: Content) -> some View {
        content
            .overlay(
                GeometryReader { geometry in
                    LinearGradient(
                        gradient: Gradient(colors: [
                            .clear,
                            .white.opacity(0.4),
                            .clear
                        ]),
                        startPoint: .leading,
                        endPoint: .trailing
                    )
                    .frame(width: geometry.size.width * 2)
                    .offset(x: -geometry.size.width + phase * geometry.size.width * 2)
                }
            )
            .mask(content)
            .onAppear {
                withAnimation(.linear(duration: 1.5).repeatForever(autoreverses: false)) {
                    phase = 1
                }
            }
    }
}

extension View {
    func shimmering() -> some View {
        modifier(ShimmerModifier())
    }
}

// MARK: - Loading Overlay

struct LoadingOverlay: ViewModifier {
    let isLoading: Bool
    var message: String = "Loading..."

    func body(content: Content) -> some View {
        ZStack {
            content
                .disabled(isLoading)
                .opacity(isLoading ? 0.5 : 1)

            if isLoading {
                VStack(spacing: 12) {
                    ProgressView()
                    Text(message)
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }
                .padding(24)
                .background(.ultraThinMaterial)
                .clipShape(RoundedRectangle(cornerRadius: 12))
            }
        }
    }
}

extension View {
    func loadingOverlay(_ isLoading: Bool, message: String = "Loading...") -> some View {
        modifier(LoadingOverlay(isLoading: isLoading, message: message))
    }
}

#Preview {
    VStack(spacing: 20) {
        LoadingView()

        SkeletonView()
            .frame(height: 20)
            .padding()

        Text("Content")
            .loadingOverlay(true)
    }
}
