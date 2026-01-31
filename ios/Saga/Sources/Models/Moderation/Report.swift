import Foundation

// MARK: - Report

/// A report submitted against a user or content
struct Report: Codable, Identifiable, Hashable, Sendable {
    let id: String
    let reporterId: String
    let targetType: ReportTargetType
    let targetId: String
    let targetUserId: String?
    let reason: ReportReason
    let details: String?
    let status: ReportStatus
    let resolution: ReportResolution?
    let resolutionNotes: String?
    let createdOn: Date
    let resolvedOn: Date?

    enum CodingKeys: String, CodingKey {
        case id
        case reporterId = "reporter_id"
        case targetType = "target_type"
        case targetId = "target_id"
        case targetUserId = "target_user_id"
        case reason
        case details
        case status
        case resolution
        case resolutionNotes = "resolution_notes"
        case createdOn = "created_on"
        case resolvedOn = "resolved_on"
    }
}

// MARK: - Report Target Type

enum ReportTargetType: String, Codable, Sendable, CaseIterable {
    case user
    case event
    case adventure
    case message
    case review
    case profile

    var displayName: String {
        switch self {
        case .user: return "User"
        case .event: return "Event"
        case .adventure: return "Adventure"
        case .message: return "Message"
        case .review: return "Review"
        case .profile: return "Profile"
        }
    }

    var iconName: String {
        switch self {
        case .user: return "person.fill"
        case .event: return "calendar"
        case .adventure: return "figure.hiking"
        case .message: return "message.fill"
        case .review: return "star.fill"
        case .profile: return "person.crop.circle"
        }
    }
}

// MARK: - Report Reason

enum ReportReason: String, Codable, Sendable, CaseIterable {
    case harassment
    case spam
    case inappropriate
    case threatening
    case impersonation
    case privacy
    case scam
    case other

    var displayName: String {
        switch self {
        case .harassment: return "Harassment"
        case .spam: return "Spam"
        case .inappropriate: return "Inappropriate Content"
        case .threatening: return "Threatening Behavior"
        case .impersonation: return "Impersonation"
        case .privacy: return "Privacy Violation"
        case .scam: return "Scam or Fraud"
        case .other: return "Other"
        }
    }

    var description: String {
        switch self {
        case .harassment: return "Bullying, intimidation, or targeted abuse"
        case .spam: return "Unsolicited messages or promotional content"
        case .inappropriate: return "Content that violates community guidelines"
        case .threatening: return "Threats of violence or harm"
        case .impersonation: return "Pretending to be someone else"
        case .privacy: return "Sharing private information without consent"
        case .scam: return "Deceptive practices or fraud attempts"
        case .other: return "Other violation not listed above"
        }
    }

    var iconName: String {
        switch self {
        case .harassment: return "exclamationmark.bubble"
        case .spam: return "envelope.badge"
        case .inappropriate: return "eye.slash"
        case .threatening: return "exclamationmark.triangle"
        case .impersonation: return "person.crop.circle.badge.questionmark"
        case .privacy: return "lock.open"
        case .scam: return "dollarsign.circle"
        case .other: return "questionmark.circle"
        }
    }
}

// MARK: - Report Status

enum ReportStatus: String, Codable, Sendable, CaseIterable {
    case pending
    case reviewing
    case resolved
    case dismissed

    var displayName: String {
        switch self {
        case .pending: return "Pending"
        case .reviewing: return "Under Review"
        case .resolved: return "Resolved"
        case .dismissed: return "Dismissed"
        }
    }
}

// MARK: - Report Resolution

enum ReportResolution: String, Codable, Sendable, CaseIterable {
    case warning
    case contentRemoved = "content_removed"
    case accountSuspended = "account_suspended"
    case accountBanned = "account_banned"
    case noAction = "no_action"

    var displayName: String {
        switch self {
        case .warning: return "Warning Issued"
        case .contentRemoved: return "Content Removed"
        case .accountSuspended: return "Account Suspended"
        case .accountBanned: return "Account Banned"
        case .noAction: return "No Action Taken"
        }
    }
}

// MARK: - Request Types

struct CreateReportRequest: Codable, Sendable {
    let targetType: ReportTargetType
    let targetId: String
    let reason: ReportReason
    let details: String?

    enum CodingKeys: String, CodingKey {
        case targetType = "target_type"
        case targetId = "target_id"
        case reason
        case details
    }
}

// MARK: - My Reports Response

struct MyReportsResponse: Codable, Sendable {
    let submitted: [Report]
    let against: [Report]
}
