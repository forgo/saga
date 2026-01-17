import Foundation
import AuthenticationServices
#if os(iOS)
import UIKit
#elseif os(macOS)
import AppKit
#endif

/// Manages Passkey (WebAuthn) authentication
@Observable
final class PasskeyService: NSObject, @unchecked Sendable {
    static let shared = PasskeyService()

    private(set) var isLoading = false

    private let relyingPartyIdentifier = "babysync.app"
    private var authContinuation: CheckedContinuation<ASAuthorizationResult, Error>?

    // MARK: - Passkey Registration

    func registerPasskey(
        challenge: Data,
        userId: Data,
        userName: String,
        userDisplayName: String
    ) async throws -> ASAuthorizationPlatformPublicKeyCredentialRegistration {
        let provider = ASAuthorizationPlatformPublicKeyCredentialProvider(
            relyingPartyIdentifier: relyingPartyIdentifier
        )

        let request = provider.createCredentialRegistrationRequest(
            challenge: challenge,
            name: userName,
            userID: userId
        )

        let result = try await performAuthorization(requests: [request])

        guard let credential = result.credential as? ASAuthorizationPlatformPublicKeyCredentialRegistration else {
            throw PasskeyError.invalidCredential
        }

        return credential
    }

    // MARK: - Passkey Login

    func loginWithPasskey(
        challenge: Data,
        allowedCredentials: [Data]? = nil
    ) async throws -> ASAuthorizationPlatformPublicKeyCredentialAssertion {
        let provider = ASAuthorizationPlatformPublicKeyCredentialProvider(
            relyingPartyIdentifier: relyingPartyIdentifier
        )

        let request = provider.createCredentialAssertionRequest(challenge: challenge)

        if let credentials = allowedCredentials {
            request.allowedCredentials = credentials.map {
                ASAuthorizationPlatformPublicKeyCredentialDescriptor(credentialID: $0)
            }
        }

        let result = try await performAuthorization(requests: [request])

        guard let credential = result.credential as? ASAuthorizationPlatformPublicKeyCredentialAssertion else {
            throw PasskeyError.invalidCredential
        }

        return credential
    }

    // MARK: - AutoFill Passkey Login

    func beginAutoFillAssistedPasskeyLogin(
        challenge: Data
    ) -> ASAuthorizationRequest {
        let provider = ASAuthorizationPlatformPublicKeyCredentialProvider(
            relyingPartyIdentifier: relyingPartyIdentifier
        )

        let request = provider.createCredentialAssertionRequest(challenge: challenge)
        return request
    }

    // MARK: - Authorization

    private func performAuthorization(
        requests: [ASAuthorizationRequest]
    ) async throws -> ASAuthorizationResult {
        let controller = ASAuthorizationController(authorizationRequests: requests)

        return try await withCheckedThrowingContinuation { continuation in
            self.authContinuation = continuation
            controller.delegate = self
            controller.presentationContextProvider = self
            controller.performRequests()
        }
    }
}

// MARK: - ASAuthorizationControllerDelegate

extension PasskeyService: ASAuthorizationControllerDelegate {
    nonisolated func authorizationController(
        controller: ASAuthorizationController,
        didCompleteWithAuthorization authorization: ASAuthorization
    ) {
        Task { @MainActor in
            authContinuation?.resume(returning: ASAuthorizationResult(credential: authorization.credential))
            authContinuation = nil
        }
    }

    nonisolated func authorizationController(
        controller: ASAuthorizationController,
        didCompleteWithError error: Error
    ) {
        Task { @MainActor in
            authContinuation?.resume(throwing: error)
            authContinuation = nil
        }
    }
}

// MARK: - ASAuthorizationControllerPresentationContextProviding

extension PasskeyService: ASAuthorizationControllerPresentationContextProviding {
    nonisolated func presentationAnchor(for controller: ASAuthorizationController) -> ASPresentationAnchor {
        // Return the key window
        #if os(iOS)
        return UIApplication.shared.connectedScenes
            .compactMap { $0 as? UIWindowScene }
            .flatMap { $0.windows }
            .first { $0.isKeyWindow } ?? UIWindow()
        #else
        return NSApplication.shared.keyWindow ?? NSWindow()
        #endif
    }
}

// MARK: - Result Type

struct ASAuthorizationResult: @unchecked Sendable {
    let credential: ASAuthorizationCredential
}

// MARK: - Errors

enum PasskeyError: Error, LocalizedError {
    case invalidCredential
    case cancelled
    case failed(Error)

    var errorDescription: String? {
        switch self {
        case .invalidCredential:
            return "Invalid passkey credential"
        case .cancelled:
            return "Passkey authentication was cancelled"
        case .failed(let error):
            return "Passkey authentication failed: \(error.localizedDescription)"
        }
    }
}

// MARK: - Credential Conversion

extension ASAuthorizationPlatformPublicKeyCredentialRegistration {
    func toPasskeyCredential() -> PasskeyCredential {
        PasskeyCredential(
            id: credentialID.base64URLEncoded(),
            rawId: credentialID.base64URLEncoded(),
            type: "public-key",
            response: PasskeyResponse(
                clientDataJSON: rawClientDataJSON.base64URLEncoded(),
                authenticatorData: nil,
                signature: nil,
                attestationObject: rawAttestationObject?.base64URLEncoded(),
                userHandle: nil
            )
        )
    }
}

extension ASAuthorizationPlatformPublicKeyCredentialAssertion {
    func toPasskeyCredential() -> PasskeyCredential {
        PasskeyCredential(
            id: credentialID.base64URLEncoded(),
            rawId: credentialID.base64URLEncoded(),
            type: "public-key",
            response: PasskeyResponse(
                clientDataJSON: rawClientDataJSON.base64URLEncoded(),
                authenticatorData: rawAuthenticatorData.base64URLEncoded(),
                signature: signature.base64URLEncoded(),
                attestationObject: nil,
                userHandle: userID.base64URLEncoded()
            )
        )
    }
}

// Note: base64URLEncoded() extension is defined in AuthService.swift
