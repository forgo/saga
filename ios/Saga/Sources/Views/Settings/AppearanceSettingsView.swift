import SwiftUI

/// Appearance and theme settings
struct AppearanceSettingsView: View {
    @AppStorage("appearance_mode") private var appearanceMode: AppearanceMode = .system
    @AppStorage("appearance_accent") private var accentColorName: String = "blue"
    @AppStorage("appearance_compact") private var useCompactMode = false

    var body: some View {
        List {
            // Theme
            Section {
                Picker("Theme", selection: $appearanceMode) {
                    ForEach(AppearanceMode.allCases, id: \.self) { mode in
                        Label(mode.displayName, systemImage: mode.iconName)
                            .tag(mode)
                    }
                }
                .pickerStyle(.inline)
                .labelsHidden()
            } header: {
                Text("Theme")
            }

            // Accent color
            Section {
                LazyVGrid(columns: Array(repeating: GridItem(.flexible()), count: 6), spacing: 16) {
                    ForEach(AccentColor.allCases, id: \.self) { color in
                        Button {
                            accentColorName = color.rawValue
                        } label: {
                            ZStack {
                                Circle()
                                    .fill(color.color)
                                    .frame(width: 44, height: 44)

                                if accentColorName == color.rawValue {
                                    Image(systemName: "checkmark")
                                        .font(.headline.bold())
                                        .foregroundStyle(.white)
                                }
                            }
                        }
                        .buttonStyle(.plain)
                    }
                }
                .padding(.vertical, 8)
            } header: {
                Text("Accent Color")
            }

            // Display
            Section {
                Toggle("Compact Mode", isOn: $useCompactMode)
            } header: {
                Text("Display")
            } footer: {
                Text("Show more content with smaller text and tighter spacing")
            }
        }
        .navigationTitle("Appearance")
        .navigationBarTitleDisplayMode(.inline)
    }
}

// MARK: - Appearance Mode

enum AppearanceMode: String, CaseIterable {
    case system
    case light
    case dark

    var displayName: String {
        switch self {
        case .system: return "System"
        case .light: return "Light"
        case .dark: return "Dark"
        }
    }

    var iconName: String {
        switch self {
        case .system: return "circle.lefthalf.filled"
        case .light: return "sun.max.fill"
        case .dark: return "moon.fill"
        }
    }

    var colorScheme: ColorScheme? {
        switch self {
        case .system: return nil
        case .light: return .light
        case .dark: return .dark
        }
    }
}

// MARK: - Accent Color

enum AccentColor: String, CaseIterable {
    case blue
    case purple
    case pink
    case red
    case orange
    case yellow
    case green
    case teal
    case indigo

    var color: Color {
        switch self {
        case .blue: return .blue
        case .purple: return .purple
        case .pink: return .pink
        case .red: return .red
        case .orange: return .orange
        case .yellow: return .yellow
        case .green: return .green
        case .teal: return .teal
        case .indigo: return .indigo
        }
    }
}

#Preview {
    NavigationStack {
        AppearanceSettingsView()
    }
}
