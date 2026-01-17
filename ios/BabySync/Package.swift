// swift-tools-version: 6.0
import PackageDescription

let package = Package(
    name: "BabySync",
    platforms: [
        .iOS(.v17)
    ],
    products: [
        .library(
            name: "BabySync",
            targets: ["BabySync"]
        ),
    ],
    dependencies: [
        .package(url: "https://github.com/kishikawakatsumi/KeychainAccess.git", from: "4.2.2"),
    ],
    targets: [
        .target(
            name: "BabySync",
            dependencies: ["KeychainAccess"],
            path: "Sources"
        ),
        .testTarget(
            name: "BabySyncTests",
            dependencies: ["BabySync"],
            path: "Tests"
        ),
    ]
)
