import Foundation

// MARK: - Person

/// A person (contact) tracked within a guild
struct Person: Codable, Sendable, Identifiable, Hashable {
    let id: String
    let name: String
    let nickname: String?
    let birthday: String?  // ISO date string (YYYY-MM-DD)
    let notes: String?
    let createdOn: Date?
    let updatedOn: Date?

    enum CodingKeys: String, CodingKey {
        case id, name, nickname, birthday, notes
        case createdOn = "created_on"
        case updatedOn = "updated_on"
    }

    /// Display name (nickname if available, otherwise name)
    var displayName: String {
        nickname ?? name
    }

    /// Initials for avatar
    var initials: String {
        let components = name.split(separator: " ")
        if components.count >= 2 {
            return "\(components[0].prefix(1))\(components[1].prefix(1))".uppercased()
        }
        return String(name.prefix(2)).uppercased()
    }

    /// Birthday as Date
    var birthdayDate: Date? {
        guard let birthday else { return nil }
        let formatter = DateFormatter()
        formatter.dateFormat = "yyyy-MM-dd"
        return formatter.date(from: birthday)
    }

    /// Days until next birthday (nil if no birthday set)
    var daysUntilBirthday: Int? {
        guard let birthDate = birthdayDate else { return nil }

        let calendar = Calendar.current
        let today = calendar.startOfDay(for: Date())

        // Get this year's birthday
        var birthdayComponents = calendar.dateComponents([.month, .day], from: birthDate)
        birthdayComponents.year = calendar.component(.year, from: today)

        guard let thisYearBirthday = calendar.date(from: birthdayComponents) else { return nil }

        // If birthday already passed this year, use next year's
        let targetBirthday: Date
        if thisYearBirthday < today {
            birthdayComponents.year! += 1
            targetBirthday = calendar.date(from: birthdayComponents) ?? thisYearBirthday
        } else {
            targetBirthday = thisYearBirthday
        }

        return calendar.dateComponents([.day], from: today, to: targetBirthday).day
    }
}

// MARK: - Person with Timers

/// Person details with their associated timers
struct PersonWithTimers: Codable, Sendable {
    let person: Person
    let timers: [TimerWithActivity]
}

// MARK: - Person Requests

struct CreatePersonRequest: Codable, Sendable {
    let name: String
    let nickname: String?
    let birthday: String?
    let notes: String?

    init(name: String, nickname: String? = nil, birthday: String? = nil, notes: String? = nil) {
        self.name = name
        self.nickname = nickname
        self.birthday = birthday
        self.notes = notes
    }
}

struct UpdatePersonRequest: Codable, Sendable {
    let name: String?
    let nickname: String?
    let birthday: String?
    let notes: String?

    init(name: String? = nil, nickname: String? = nil, birthday: String? = nil, notes: String? = nil) {
        self.name = name
        self.nickname = nickname
        self.birthday = birthday
        self.notes = notes
    }
}
