import Foundation
import CommonCrypto

/// PKCE (Proof Key for Code Exchange) helper for OAuth flows
struct PKCEHelper {
    let codeVerifier: String
    let codeChallenge: String

    init() {
        // Generate random 32-byte code verifier
        var buffer = [UInt8](repeating: 0, count: 32)
        _ = SecRandomCopyBytes(kSecRandomDefault, buffer.count, &buffer)
        codeVerifier = Data(buffer).base64URLEncoded()

        // SHA256 hash and base64url encode for challenge
        let data = Data(codeVerifier.utf8)
        var hash = [UInt8](repeating: 0, count: 32)
        data.withUnsafeBytes { buffer in
            _ = CC_SHA256(buffer.baseAddress, CC_LONG(buffer.count), &hash)
        }
        codeChallenge = Data(hash).base64URLEncoded()
    }

    /// The code challenge method (always S256)
    static let challengeMethod = "S256"
}

// MARK: - OAuth URL Helpers

extension PKCEHelper {
    /// Build Google OAuth authorization URL
    func googleAuthURL(clientID: String, redirectURI: String, scopes: [String] = ["openid", "email", "profile"]) -> URL? {
        var components = URLComponents(string: "https://accounts.google.com/o/oauth2/v2/auth")
        components?.queryItems = [
            URLQueryItem(name: "client_id", value: clientID),
            URLQueryItem(name: "redirect_uri", value: redirectURI),
            URLQueryItem(name: "response_type", value: "code"),
            URLQueryItem(name: "scope", value: scopes.joined(separator: " ")),
            URLQueryItem(name: "code_challenge", value: codeChallenge),
            URLQueryItem(name: "code_challenge_method", value: Self.challengeMethod),
            URLQueryItem(name: "access_type", value: "offline"),
            URLQueryItem(name: "prompt", value: "consent")
        ]
        return components?.url
    }
}
