import Foundation

struct Baby: Codable, Identifiable, Equatable, Sendable {
    let id: String
    let familyId: String
    var name: String
    let createdOn: Date
    let updatedOn: Date

    enum CodingKeys: String, CodingKey {
        case id, name
        case familyId = "family_id"
        case createdOn = "created_on"
        case updatedOn = "updated_on"
    }
}

struct CreateBabyRequest: Codable, Sendable {
    let name: String
}

struct UpdateBabyRequest: Codable, Sendable {
    var name: String?
}
