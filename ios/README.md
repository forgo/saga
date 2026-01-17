# BabySync iOS

Modern SwiftUI app for tracking baby care activities across families.

## Requirements

- iOS 17.0+
- macOS 14.0+ (for development)
- Xcode 16.0+
- Swift 6.0

## Architecture

The app follows a modern SwiftUI architecture with:

- **@Observable** for reactive state management
- **async/await** for all networking
- **Swift Concurrency** with actors for thread safety
- **Server-Sent Events (SSE)** for real-time updates
- **MVVM** pattern for view organization

## Project Structure

```
Sources/
├── App/
│   └── BabySyncApp.swift          # App entry point
├── Models/
│   ├── User.swift                 # User, Identity, Passkey models
│   ├── Family.swift               # Family, Parent models
│   ├── Baby.swift                 # Baby model
│   ├── Activity.swift             # Activity model
│   ├── Timer.swift                # BabyTimer model
│   ├── APIResponse.swift          # Response wrappers, errors
│   └── AuthRequests.swift         # Auth request DTOs
├── Networking/
│   ├── APIClient.swift            # REST API client
│   └── SSEClient.swift            # Server-Sent Events client
├── Services/
│   ├── AuthService.swift          # Authentication management
│   ├── PasskeyService.swift       # WebAuthn passkey support
│   └── FamilyService.swift        # Family data management
├── Views/
│   ├── Auth/
│   │   └── AuthView.swift         # Login/Register screen
│   ├── Family/
│   │   ├── FamilyListView.swift   # Family list
│   │   └── FamilyDetailView.swift # Family details with babies
│   ├── Baby/
│   │   └── BabyDetailView.swift   # Baby with timers
│   ├── Timer/
│   │   └── TimerCard.swift        # Timer display component
│   └── Settings/
│       └── SettingsView.swift     # Settings & account
└── Tests/
    └── BabySyncTests.swift        # Unit tests
```

## Features

### Authentication
- Email/password registration and login
- Sign in with Apple
- Sign in with Google (PKCE)
- Passkeys (WebAuthn) for biometric login

### Family Management
- Create and join families
- View family members
- Leave or merge families

### Baby Tracking
- Add babies to families
- Create activity timers for each baby
- Real-time timer synchronization across devices

### Timers
- Visual timer cards with elapsed time
- Warning and critical threshold indicators
- One-tap reset
- Push notification support

### Real-Time Sync
- SSE connection for live updates
- Automatic reconnection with exponential backoff
- Offline-capable timer display (client-side calculation)

## Setup

1. Open `Package.swift` in Xcode
2. Wait for Swift Package Manager to resolve dependencies
3. Select a simulator or device
4. Build and run

## Configuration

The API base URL can be configured in `APIClient.swift`:

```swift
init(
    baseURL: URL = URL(string: "http://localhost:8080/v1")!,
    ...
)
```

For production, update this to your deployed API URL.

## Dependencies

- **KeychainAccess** - Secure token storage

## Testing

Run tests with:
```bash
swift test
```

Or use Xcode's Test Navigator (⌘+6).
