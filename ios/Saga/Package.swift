// swift-tools-version: 6.0
import PackageDescription

let package = Package(
    name: "Saga",
    platforms: [
        .iOS(.v17)
    ],
    products: [
        .library(
            name: "Saga",
            targets: ["Saga"]
        ),
    ],
    dependencies: [
        .package(url: "https://github.com/kishikawakatsumi/KeychainAccess.git", from: "4.2.2"),
    ],
    targets: [
        .target(
            name: "Saga",
            dependencies: ["KeychainAccess"],
            path: "Sources"
        ),
        .testTarget(
            name: "SagaTests",
            dependencies: ["Saga"],
            path: "Tests"
        ),
    ]
)
