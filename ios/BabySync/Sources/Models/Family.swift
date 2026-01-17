import Foundation

struct Family: Codable, Identifiable, Equatable, Sendable {
    let id: String
    var name: String
    let createdOn: Date
    let updatedOn: Date

    enum CodingKeys: String, CodingKey {
        case id, name
        case createdOn = "created_on"
        case updatedOn = "updated_on"
    }
}

struct Parent: Codable, Identifiable, Equatable, Sendable {
    let id: String
    let name: String
    let email: String
}

struct FamilyData: Codable, Sendable {
    var family: Family
    var parents: [Parent]
    var babies: [Baby]
    var activities: [Activity]
}
