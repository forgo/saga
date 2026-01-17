# Saga iOS

Modern SwiftUI app for social coordination and meaningful connections.

## Requirements

- iOS 17.0+
- macOS 14.0+ (for development)
- Xcode 16.0+
- Swift 6.0

## Architecture

The app follows a modern SwiftUI architecture with:

- **@Observable** for reactive state management
- **async/await** for all networking
- **Swift 6 Concurrency** with actors for thread safety
- **Server-Sent Events (SSE)** for real-time updates

## Project Structure

```
ios/Saga/
├── Sources/
│   ├── App/
│   │   ├── SagaApp.swift              # App entry point
│   │   ├── AppState.swift             # Global state coordinator
│   │   └── Environment.swift          # API URL configuration
│   ├── Models/
│   │   ├── Core/                      # Guild, Person, User, Activity, Timer
│   │   ├── Events/                    # Event, EventRole
│   │   ├── Social/                    # Profile, Trust, Review, Availability
│   │   ├── Discovery/                 # Interest, Questionnaire
│   │   ├── Advanced/                  # Adventure, Pool, Vote, RoleCatalog
│   │   ├── Gamification/              # Resonance
│   │   ├── Moderation/                # Report, Block
│   │   └── API/                       # APIResponse, APIError, AuthRequests
│   ├── Networking/
│   │   ├── APIClient.swift            # REST API client (actor-based)
│   │   ├── APIClient+*.swift          # API extensions by domain
│   │   └── SSEClient.swift            # Server-Sent Events client
│   ├── Services/
│   │   ├── AuthService.swift          # Authentication & token management
│   │   ├── PasskeyService.swift       # WebAuthn passkey support
│   │   ├── GuildService.swift         # Guild data & real-time sync
│   │   ├── EventService.swift         # Event management
│   │   ├── ProfileService.swift       # User profiles
│   │   └── DiscoveryService.swift     # Discovery & matching
│   └── Views/
│       ├── App/                       # ContentView, MainTabView
│       ├── Auth/                      # Login, Register, Passkey setup
│       ├── Guild/                     # Guild list, detail, management
│       ├── People/                    # Person detail, timers
│       ├── Events/                    # Event list, detail, RSVP
│       ├── Social/                    # Profile, Trust, Availability
│       ├── Discovery/                 # Nearby, Interests, Questionnaires
│       ├── Advanced/                  # Adventures, Pools, Votes
│       ├── Moderation/                # Reports, Blocks, Resonance
│       ├── Settings/                  # Account, Notifications, Privacy
│       └── Components/                # Reusable UI components
└── Tests/
    └── SagaTests.swift
```

## Features

### Authentication
- Email/password registration and login
- Sign in with Apple
- Sign in with Google (PKCE)
- Passkeys (WebAuthn) for biometric login
- Token refresh with secure keychain storage

### Guilds
- Create and join guilds (communities)
- View guild members
- Manage people and relationships
- Activity timers with thresholds

### Events
- Create and discover events
- RSVP management (going/maybe/not going)
- Event roles and assignments
- Check-in and feedback

### Discovery
- Interest-based matching (teach/learn)
- Questionnaire compatibility scoring
- Nearby availability
- Event recommendations

### Trust Network
- Trust grants (basic/elevated/full)
- IRL confirmations
- Trust ratings with anchors
- Endorsements and reviews

### Advanced Features
- Adventures with admission control
- Matching pools (donut-style)
- Voting (FPTP, ranked choice, approval, multi-select)
- Role catalogs

### Moderation & Safety
- Report submission
- Block/unblock users
- Moderation status

### Gamification
- Resonance scoring
- Level progression
- Score breakdown

### Real-Time Sync
- SSE connection for live updates
- Automatic reconnection with exponential backoff

## Setup

1. Open `ios/Saga/Saga.xcodeproj` in Xcode
2. Wait for Swift Package Manager to resolve dependencies
3. Select a simulator or device
4. Build and run (⌘R)

Or use Make:
```bash
make dev-ios
```

## Configuration

The API environment is configured in `Environment.swift`:

```swift
enum APIEnvironment {
    case development
    case staging
    case production
}
```

## Dependencies

- **KeychainAccess** - Secure token storage

## Testing

### Unit Tests

Run unit tests with:
```bash
make test-ios
# Or directly:
cd ios/Saga && swift test
```

Or use Xcode's Test Navigator (⌘6).

### UI Tests (E2E)

UI tests require the API running locally with seed data:

```bash
# Terminal 1: Start API with seed data
make dev

# Terminal 2: Run UI tests
cd ios/Saga
xcodebuild test -scheme SagaUITests -destination 'platform=iOS Simulator,name=iPhone 16 Pro'
```

### Test Modes

The app supports launch arguments for testing:

| Argument | Description |
|----------|-------------|
| `--uitesting` | Enables UI test mode, uses test environment |
| `--demo` | Auto-login with demo user (for quick feature testing) |

These can be set in Xcode's scheme editor (Product > Scheme > Edit Scheme > Run > Arguments) or passed via `XCUIApplication.launchArguments` in UI tests.

### Demo Credentials

For manual testing with the local API:

| Email | Password |
|-------|----------|
| `demo@forgo.software` | `password123` |
| `second@forgo.software` | `password123` |

### Accessibility Identifiers

Key views include accessibility identifiers for reliable UI testing:

- `login_email_field` - Email text field
- `login_password_field` - Password secure field
- `login_submit_button` - Sign in button
- `login_error_message` - Error message label
- `guild_list` - Guild list view
- `create_guild_button` - Create guild button
- `guild_row_{id}` - Individual guild row
