import Foundation

extension Data {
    /// Encode data as Base64URL (RFC 4648 Section 5)
    func base64URLEncoded() -> String {
        base64EncodedString()
            .replacingOccurrences(of: "+", with: "-")
            .replacingOccurrences(of: "/", with: "_")
            .replacingOccurrences(of: "=", with: "")
    }

    /// Initialize from Base64URL encoded string
    init?(base64URLEncoded string: String) {
        // Convert Base64URL to standard Base64
        var base64 = string
            .replacingOccurrences(of: "-", with: "+")
            .replacingOccurrences(of: "_", with: "/")

        // Add padding if needed
        let remainder = base64.count % 4
        if remainder > 0 {
            base64 += String(repeating: "=", count: 4 - remainder)
        }

        self.init(base64Encoded: base64)
    }
}

extension String {
    /// Decode Base64URL string to Data
    func base64URLDecoded() -> Data? {
        Data(base64URLEncoded: self)
    }
}
