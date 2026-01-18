import SwiftUI
import AuthenticationServices

/// Main authentication view with login and registration options
struct AuthView: View {
    @Environment(AuthService.self) private var authService
    @Environment(PasskeyService.self) private var passkeyService

    @State private var mode: AuthMode = .login
    @State private var email = ""
    @State private var password = ""
    @State private var firstname = ""
    @State private var lastname = ""
    @State private var errorMessage: String?
    @State private var isShowingError = false
    @State private var pkceHelper: PKCEHelper?

    enum AuthMode {
        case login
        case register
    }

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(spacing: 32) {
                    // Logo/Title
                    VStack(spacing: 8) {
                        Image(systemName: "person.3.fill")
                            .font(.system(size: 60))
                            .foregroundStyle(.blue)
                        Text("Saga")
                            .font(.largeTitle.bold())
                        Text("Build meaningful connections")
                            .foregroundStyle(.secondary)
                    }
                    .padding(.top, 40)

                    // Mode picker
                    Picker("Mode", selection: $mode) {
                        Text("Sign In").tag(AuthMode.login)
                        Text("Create Account").tag(AuthMode.register)
                    }
                    .pickerStyle(.segmented)
                    .padding(.horizontal)
                    .accessibilityIdentifier("auth_mode_picker")

                    // Email/Password form
                    VStack(spacing: 16) {
                        if mode == .register {
                            HStack(spacing: 12) {
                                TextField("First name", text: $firstname)
                                    .textContentType(.givenName)
                                    .textInputAutocapitalization(.words)
                                    .padding()
                                    .background(.fill.tertiary)
                                    .clipShape(RoundedRectangle(cornerRadius: 12))
                                    .accessibilityIdentifier("auth_firstname_field")

                                TextField("Last name", text: $lastname)
                                    .textContentType(.familyName)
                                    .textInputAutocapitalization(.words)
                                    .padding()
                                    .background(.fill.tertiary)
                                    .clipShape(RoundedRectangle(cornerRadius: 12))
                                    .accessibilityIdentifier("auth_lastname_field")
                            }
                        }

                        TextField("Email", text: $email)
                            .textContentType(.emailAddress)
                            .textInputAutocapitalization(.never)
                            .keyboardType(.emailAddress)
                            .autocorrectionDisabled()
                            .padding()
                            .background(.fill.tertiary)
                            .clipShape(RoundedRectangle(cornerRadius: 12))
                            .accessibilityIdentifier("auth_email_field")

                        SecureField("Password", text: $password)
                            .textContentType(mode == .login ? .password : .newPassword)
                            .padding()
                            .background(.fill.tertiary)
                            .clipShape(RoundedRectangle(cornerRadius: 12))
                            .accessibilityIdentifier("auth_password_field")

                        if let errorMessage, isShowingError {
                            Text(errorMessage)
                                .foregroundStyle(.red)
                                .font(.subheadline)
                                .accessibilityIdentifier("auth_error_message")
                        }

                        Button {
                            Task { await handleEmailAuth() }
                        } label: {
                            HStack {
                                if authService.isLoading {
                                    ProgressView()
                                        .tint(.white)
                                } else {
                                    Text(mode == .login ? "Sign In" : "Create Account")
                                }
                            }
                            .frame(maxWidth: .infinity)
                            .padding()
                            .background(.blue)
                            .foregroundStyle(.white)
                            .clipShape(RoundedRectangle(cornerRadius: 12))
                        }
                        .disabled(authService.isLoading || !isFormValid)
                        .accessibilityIdentifier("auth_submit_button")
                    }
                    .padding(.horizontal)

                    // Divider
                    HStack {
                        Rectangle()
                            .fill(.secondary.opacity(0.3))
                            .frame(height: 1)
                        Text("or")
                            .foregroundStyle(.secondary)
                            .font(.subheadline)
                        Rectangle()
                            .fill(.secondary.opacity(0.3))
                            .frame(height: 1)
                    }
                    .padding(.horizontal)

                    // Social auth buttons
                    VStack(spacing: 12) {
                        // Sign in with Apple
                        SignInWithAppleButton(
                            mode == .login ? .signIn : .signUp,
                            onRequest: configureAppleRequest,
                            onCompletion: handleAppleResult
                        )
                        .signInWithAppleButtonStyle(.black)
                        .frame(height: 50)
                        .clipShape(RoundedRectangle(cornerRadius: 12))

                        // Sign in with Google
                        Button {
                            startGoogleAuth()
                        } label: {
                            HStack {
                                Image(systemName: "g.circle.fill")
                                Text(mode == .login ? "Sign in with Google" : "Sign up with Google")
                            }
                            .frame(maxWidth: .infinity)
                            .padding()
                            .background(.white)
                            .foregroundStyle(.black)
                            .clipShape(RoundedRectangle(cornerRadius: 12))
                            .overlay(
                                RoundedRectangle(cornerRadius: 12)
                                    .stroke(.secondary.opacity(0.3), lineWidth: 1)
                            )
                        }
                        .accessibilityIdentifier("auth_google_button")

                        // Passkey login (only in login mode)
                        if mode == .login {
                            Button {
                                Task { await handlePasskeyLogin() }
                            } label: {
                                HStack {
                                    Image(systemName: "person.badge.key.fill")
                                    Text("Sign in with Passkey")
                                }
                                .frame(maxWidth: .infinity)
                                .padding()
                                .background(.fill.tertiary)
                                .clipShape(RoundedRectangle(cornerRadius: 12))
                            }
                            .accessibilityIdentifier("auth_passkey_button")
                        }
                    }
                    .padding(.horizontal)

                    Spacer(minLength: 40)
                }
            }
            .navigationBarTitleDisplayMode(.inline)
            .alert("Error", isPresented: $isShowingError) {
                Button("OK", role: .cancel) { }
            } message: {
                Text(errorMessage ?? "An error occurred")
            }
        }
    }

    // MARK: - Form Validation

    private var isFormValid: Bool {
        let emailValid = email.contains("@") && email.contains(".")
        let passwordValid = password.count >= 8
        return emailValid && passwordValid
    }

    // MARK: - Email/Password Auth

    private func handleEmailAuth() async {
        do {
            if mode == .login {
                try await authService.login(email: email, password: password)
            } else {
                try await authService.register(
                    email: email,
                    password: password,
                    firstname: firstname.isEmpty ? nil : firstname,
                    lastname: lastname.isEmpty ? nil : lastname
                )
            }
        } catch {
            showError(error)
        }
    }

    // MARK: - Apple Auth

    private func configureAppleRequest(_ request: ASAuthorizationAppleIDRequest) {
        request.requestedScopes = [.email, .fullName]
        // Note: Apple doesn't use PKCE in the standard way, the code is passed directly
    }

    private func handleAppleResult(_ result: Result<ASAuthorization, Error>) {
        switch result {
        case .success(let authorization):
            guard let credential = authorization.credential as? ASAuthorizationAppleIDCredential,
                  let codeData = credential.authorizationCode,
                  let code = String(data: codeData, encoding: .utf8) else {
                showError(PasskeyError.invalidCredential)
                return
            }

            Task {
                do {
                    // Apple OAuth doesn't use PKCE, but our API expects a code_verifier
                    // We send an empty string as a placeholder
                    try await authService.loginWithApple(code: code, codeVerifier: "")
                } catch {
                    showError(error)
                }
            }

        case .failure(let error):
            if (error as? ASAuthorizationError)?.code != .canceled {
                showError(error)
            }
        }
    }

    // MARK: - Google Auth

    private func startGoogleAuth() {
        pkceHelper = PKCEHelper()
        guard let helper = pkceHelper else { return }

        let clientID = currentEnvironment.googleClientID
        let redirectURI = "app.saga://oauth/callback"

        guard let url = helper.googleAuthURL(clientID: clientID, redirectURI: redirectURI) else {
            showError(APIError.invalidURL)
            return
        }

        // Open Google auth in Safari
        // In a real app, you'd use ASWebAuthenticationSession
        #if os(iOS)
        UIApplication.shared.open(url)
        #endif
    }

    // MARK: - Passkey Auth

    private func handlePasskeyLogin() async {
        do {
            // Get challenge from server
            let challengeResponse = try await APIClient.shared.beginPasskeyLogin(email: nil)
            guard let challengeData = Data(base64URLEncoded: challengeResponse.data.challenge) else {
                throw PasskeyError.invalidCredential
            }

            // Get allowed credentials if any
            let allowedCredentials = challengeResponse.data.allowCredentials?.compactMap {
                Data(base64URLEncoded: $0.id)
            }

            // Perform WebAuthn assertion
            let assertion = try await passkeyService.loginWithPasskey(
                challenge: challengeData,
                allowedCredentials: allowedCredentials
            )

            // Complete login with API
            try await authService.loginWithPasskey(credential: assertion.toPasskeyCredential())

        } catch PasskeyError.cancelled {
            // User cancelled, ignore
        } catch {
            showError(error)
        }
    }

    // MARK: - Error Handling

    private func showError(_ error: Error) {
        if let apiError = error as? APIError {
            errorMessage = apiError.userMessage
        } else {
            errorMessage = error.localizedDescription
        }
        isShowingError = true
    }
}

#Preview {
    AuthView()
        .environment(AuthService.shared)
        .environment(PasskeyService.shared)
}
