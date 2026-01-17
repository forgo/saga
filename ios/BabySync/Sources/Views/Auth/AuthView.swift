import SwiftUI
import AuthenticationServices

struct AuthView: View {
    @Environment(AuthService.self) private var authService

    @State private var email = ""
    @State private var password = ""
    @State private var isRegistering = false
    @State private var showError = false
    @State private var errorMessage = ""

    var body: some View {
        NavigationStack {
            VStack(spacing: 32) {
                // Logo and title
                VStack(spacing: 8) {
                    Image(systemName: "clock.badge.checkmark.fill")
                        .font(.system(size: 64))
                        .foregroundStyle(.blue)

                    Text("BabySync")
                        .font(.largeTitle)
                        .fontWeight(.bold)

                    Text("Track baby care together")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                }
                .padding(.top, 40)

                Spacer()

                // Social sign-in buttons
                VStack(spacing: 12) {
                    SignInWithAppleButton(.signIn) { request in
                        request.requestedScopes = [.email, .fullName]
                    } onCompletion: { result in
                        handleAppleSignIn(result)
                    }
                    .frame(height: 50)
                    .cornerRadius(8)

                    Button {
                        handleGoogleSignIn()
                    } label: {
                        HStack {
                            Image(systemName: "g.circle.fill")
                            Text("Sign in with Google")
                        }
                        .frame(maxWidth: .infinity)
                        .frame(height: 50)
                        #if os(iOS)
                        .background(Color(.systemGray6))
                        #else
                        .background(Color.gray.opacity(0.1))
                        #endif
                        .cornerRadius(8)
                    }
                    .buttonStyle(.plain)
                }
                .padding(.horizontal)

                // Divider
                HStack {
                    Rectangle()
                        .fill(Color.gray.opacity(0.3))
                        .frame(height: 1)
                    Text("or")
                        .foregroundStyle(.secondary)
                        .font(.footnote)
                    Rectangle()
                        .fill(Color.gray.opacity(0.3))
                        .frame(height: 1)
                }
                .padding(.horizontal)

                // Email/password form
                VStack(spacing: 16) {
                    TextField("Email", text: $email)
                        .textContentType(.emailAddress)
                        .keyboardType(.emailAddress)
                        .autocapitalization(.none)
                        .textFieldStyle(.roundedBorder)

                    SecureField("Password", text: $password)
                        .textContentType(isRegistering ? .newPassword : .password)
                        .textFieldStyle(.roundedBorder)

                    Button {
                        Task { await handleEmailAuth() }
                    } label: {
                        if authService.isLoading {
                            ProgressView()
                                .frame(maxWidth: .infinity)
                                .frame(height: 50)
                        } else {
                            Text(isRegistering ? "Create Account" : "Sign In")
                                .frame(maxWidth: .infinity)
                                .frame(height: 50)
                        }
                    }
                    .buttonStyle(.borderedProminent)
                    .disabled(email.isEmpty || password.isEmpty || authService.isLoading)

                    Button {
                        withAnimation {
                            isRegistering.toggle()
                        }
                    } label: {
                        Text(isRegistering ? "Already have an account? Sign In" : "Don't have an account? Register")
                            .font(.footnote)
                    }
                }
                .padding(.horizontal)

                Spacer()

                // Passkey button
                Button {
                    Task { await handlePasskeyLogin() }
                } label: {
                    Label("Sign in with Passkey", systemImage: "person.badge.key.fill")
                        .font(.footnote)
                }
                .padding(.bottom)
            }
            .alert("Error", isPresented: $showError) {
                Button("OK") { }
            } message: {
                Text(errorMessage)
            }
        }
    }

    private func handleEmailAuth() async {
        do {
            if isRegistering {
                try await authService.register(email: email, password: password)
            } else {
                try await authService.login(email: email, password: password)
            }
        } catch {
            errorMessage = error.localizedDescription
            showError = true
        }
    }

    private func handleAppleSignIn(_ result: Result<ASAuthorization, Error>) {
        switch result {
        case .success(let authorization):
            guard let credential = authorization.credential as? ASAuthorizationAppleIDCredential,
                  let authCodeData = credential.authorizationCode,
                  let authCode = String(data: authCodeData, encoding: .utf8) else {
                errorMessage = "Failed to get Apple authorization code"
                showError = true
                return
            }

            // Note: In production, you'd use PKCE here
            Task {
                do {
                    try await authService.loginWithApple(code: authCode, codeVerifier: "")
                } catch {
                    await MainActor.run {
                        errorMessage = error.localizedDescription
                        showError = true
                    }
                }
            }

        case .failure(let error):
            errorMessage = error.localizedDescription
            showError = true
        }
    }

    private func handleGoogleSignIn() {
        // In production, implement Google Sign-In with PKCE
        errorMessage = "Google Sign-In not implemented in this demo"
        showError = true
    }

    private func handlePasskeyLogin() async {
        // In production, get challenge from server and perform passkey auth
        errorMessage = "Passkey login coming soon"
        showError = true
    }
}

#Preview {
    AuthView()
        .environment(AuthService.shared)
}
