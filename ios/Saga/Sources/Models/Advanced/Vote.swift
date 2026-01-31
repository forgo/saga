import Foundation

// MARK: - Vote

/// A vote/poll within a guild
struct Vote: Codable, Identifiable, Hashable, Sendable {
    let id: String
    let guildId: String
    let creatorId: String
    let title: String
    let description: String?
    let voteType: VoteType
    let options: [VoteOption]
    let settings: VoteSettings
    let status: VoteStatus
    let startTime: Date
    let endTime: Date?
    let totalVoters: Int
    let createdOn: Date

    enum CodingKeys: String, CodingKey {
        case id
        case guildId = "guild_id"
        case creatorId = "creator_id"
        case title
        case description
        case voteType = "vote_type"
        case options
        case settings
        case status
        case startTime = "start_time"
        case endTime = "end_time"
        case totalVoters = "total_voters"
        case createdOn = "created_on"
    }

    /// Whether the vote is currently active
    var isActive: Bool {
        guard status == .active else { return false }
        if let endTime = endTime, endTime < Date() { return false }
        return true
    }

    /// Time remaining text
    var timeRemainingText: String? {
        guard let endTime = endTime else { return nil }
        let remaining = endTime.timeIntervalSince(Date())
        if remaining <= 0 { return "Ended" }

        let hours = Int(remaining / 3600)
        let minutes = Int((remaining.truncatingRemainder(dividingBy: 3600)) / 60)

        if hours > 24 {
            let days = hours / 24
            return "\(days) day\(days == 1 ? "" : "s") left"
        } else if hours > 0 {
            return "\(hours)h \(minutes)m left"
        } else {
            return "\(minutes)m left"
        }
    }
}

// MARK: - Vote Type

enum VoteType: String, Codable, Sendable, CaseIterable {
    case fptp = "fptp"
    case ranked = "ranked"
    case approval
    case multiSelect = "multi_select"

    var displayName: String {
        switch self {
        case .fptp: return "Single Choice"
        case .ranked: return "Ranked Choice"
        case .approval: return "Approval"
        case .multiSelect: return "Multi-Select"
        }
    }

    var description: String {
        switch self {
        case .fptp: return "Pick one option (First Past the Post)"
        case .ranked: return "Rank options in order of preference"
        case .approval: return "Approve or disapprove each option"
        case .multiSelect: return "Select multiple options"
        }
    }

    var iconName: String {
        switch self {
        case .fptp: return "1.circle"
        case .ranked: return "list.number"
        case .approval: return "hand.thumbsup"
        case .multiSelect: return "checklist"
        }
    }
}

// MARK: - Vote Option

/// An option in a vote
struct VoteOption: Codable, Identifiable, Hashable, Sendable {
    let id: String
    let voteId: String
    let text: String
    let description: String?
    let sortOrder: Int

    enum CodingKeys: String, CodingKey {
        case id
        case voteId = "vote_id"
        case text
        case description
        case sortOrder = "sort_order"
    }
}

// MARK: - Vote Settings

/// Settings for a vote
struct VoteSettings: Codable, Hashable, Sendable {
    let allowAbstain: Bool
    let showResultsBeforeEnd: Bool
    let maxSelections: Int?
    let anonymousVoting: Bool
    let requireAllRanked: Bool

    enum CodingKeys: String, CodingKey {
        case allowAbstain = "allow_abstain"
        case showResultsBeforeEnd = "show_results_before_end"
        case maxSelections = "max_selections"
        case anonymousVoting = "anonymous_voting"
        case requireAllRanked = "require_all_ranked"
    }
}

// MARK: - Vote Status

enum VoteStatus: String, Codable, Sendable, CaseIterable {
    case draft
    case active
    case ended
    case cancelled

    var displayName: String {
        switch self {
        case .draft: return "Draft"
        case .active: return "Active"
        case .ended: return "Ended"
        case .cancelled: return "Cancelled"
        }
    }
}

// MARK: - Ballot

/// A user's ballot for a vote
struct Ballot: Codable, Identifiable, Hashable, Sendable {
    let id: String
    let voteId: String
    let userId: String
    let selections: [BallotSelection]
    let abstained: Bool
    let submittedOn: Date

    enum CodingKeys: String, CodingKey {
        case id
        case voteId = "vote_id"
        case userId = "user_id"
        case selections
        case abstained
        case submittedOn = "submitted_on"
    }
}

// MARK: - Ballot Selection

/// A selection on a ballot
struct BallotSelection: Codable, Hashable, Sendable {
    let optionId: String
    let rank: Int?
    let approved: Bool?

    enum CodingKeys: String, CodingKey {
        case optionId = "option_id"
        case rank
        case approved
    }
}

// MARK: - Vote Results

/// Results of a vote
struct VoteResults: Codable, Sendable {
    let voteId: String
    let totalVoters: Int
    let totalAbstained: Int
    let optionResults: [OptionResult]
    let winner: VoteOption?
    let calculatedOn: Date

    enum CodingKeys: String, CodingKey {
        case voteId = "vote_id"
        case totalVoters = "total_voters"
        case totalAbstained = "total_abstained"
        case optionResults = "option_results"
        case winner
        case calculatedOn = "calculated_on"
    }
}

// MARK: - Option Result

/// Result for a single option
struct OptionResult: Codable, Sendable {
    let optionId: String
    let optionText: String
    let votes: Int
    let percentage: Double
    let approvals: Int?
    let disapprovals: Int?
    let averageRank: Double?

    enum CodingKeys: String, CodingKey {
        case optionId = "option_id"
        case optionText = "option_text"
        case votes
        case percentage
        case approvals
        case disapprovals
        case averageRank = "average_rank"
    }

    var formattedPercentage: String {
        String(format: "%.1f%%", percentage * 100)
    }
}

// MARK: - Vote With Details

/// Vote with user's ballot and results
struct VoteWithDetails: Codable, Sendable {
    let vote: Vote
    let myBallot: Ballot?
    let results: VoteResults?

    enum CodingKeys: String, CodingKey {
        case vote
        case myBallot = "my_ballot"
        case results
    }
}

// MARK: - Request Types

struct CreateVoteRequest: Codable, Sendable {
    let guildId: String
    let title: String
    let description: String?
    let voteType: VoteType
    let options: [CreateVoteOptionRequest]
    let settings: VoteSettings
    let startTime: Date
    let endTime: Date?

    enum CodingKeys: String, CodingKey {
        case guildId = "guild_id"
        case title
        case description
        case voteType = "vote_type"
        case options
        case settings
        case startTime = "start_time"
        case endTime = "end_time"
    }
}

struct CreateVoteOptionRequest: Codable, Sendable {
    let text: String
    let description: String?
}

struct UpdateVoteRequest: Codable, Sendable {
    var title: String?
    var description: String?
    var endTime: Date?
    var status: VoteStatus?

    enum CodingKeys: String, CodingKey {
        case title
        case description
        case endTime = "end_time"
        case status
    }
}

struct CastBallotRequest: Codable, Sendable {
    let selections: [BallotSelection]
    let abstain: Bool
}
