import SwiftUI

struct FamilyListView: View {
    @Environment(FamilyService.self) private var familyService

    @State private var showCreateFamily = false
    @State private var showJoinFamily = false
    @State private var newFamilyName = ""
    @State private var joinFamilyId = ""

    var body: some View {
        NavigationStack {
            Group {
                if familyService.isLoading && familyService.families.isEmpty {
                    ProgressView("Loading families...")
                } else if familyService.families.isEmpty {
                    ContentUnavailableView {
                        Label("No Families", systemImage: "house")
                    } description: {
                        Text("Create a family or join an existing one")
                    } actions: {
                        Button("Create Family") {
                            showCreateFamily = true
                        }
                        .buttonStyle(.borderedProminent)

                        Button("Join Family") {
                            showJoinFamily = true
                        }
                    }
                } else {
                    List(familyService.families) { family in
                        NavigationLink {
                            FamilyDetailView(familyId: family.id)
                        } label: {
                            FamilyRow(family: family)
                        }
                    }
                    .refreshable {
                        try? await familyService.loadFamilies()
                    }
                }
            }
            .navigationTitle("Families")
            .toolbar {
                ToolbarItem(placement: .primaryAction) {
                    Menu {
                        Button {
                            showCreateFamily = true
                        } label: {
                            Label("Create Family", systemImage: "plus")
                        }

                        Button {
                            showJoinFamily = true
                        } label: {
                            Label("Join Family", systemImage: "person.badge.plus")
                        }
                    } label: {
                        Image(systemName: "plus")
                    }
                }
            }
            .sheet(isPresented: $showCreateFamily) {
                CreateFamilySheet(familyName: $newFamilyName) {
                    Task {
                        try? await familyService.createFamily(name: newFamilyName.isEmpty ? nil : newFamilyName)
                        newFamilyName = ""
                        showCreateFamily = false
                    }
                }
            }
            .sheet(isPresented: $showJoinFamily) {
                JoinFamilySheet(familyId: $joinFamilyId) {
                    Task {
                        try? await familyService.joinFamily(id: joinFamilyId)
                        joinFamilyId = ""
                        showJoinFamily = false
                    }
                }
            }
            .task {
                try? await familyService.loadFamilies()
            }
        }
    }
}

struct FamilyRow: View {
    let family: Family

    var body: some View {
        VStack(alignment: .leading, spacing: 4) {
            Text(family.name)
                .font(.headline)

            Text("Created \(family.createdOn.formatted(date: .abbreviated, time: .omitted))")
                .font(.caption)
                .foregroundStyle(.secondary)
        }
        .padding(.vertical, 4)
    }
}

struct CreateFamilySheet: View {
    @Binding var familyName: String
    let onCreate: () -> Void

    @Environment(\.dismiss) private var dismiss

    var body: some View {
        NavigationStack {
            Form {
                Section {
                    TextField("Family Name", text: $familyName)
                } footer: {
                    Text("Leave blank to use \"My Family\"")
                }
            }
            .navigationTitle("Create Family")
            #if os(iOS)
            .navigationBarTitleDisplayMode(.inline)
            #endif
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Create", action: onCreate)
                }
            }
        }
        .presentationDetents([.medium])
    }
}

struct JoinFamilySheet: View {
    @Binding var familyId: String
    let onJoin: () -> Void

    @Environment(\.dismiss) private var dismiss

    var body: some View {
        NavigationStack {
            Form {
                Section {
                    TextField("Family ID", text: $familyId)
                } footer: {
                    Text("Enter the family ID shared with you")
                }
            }
            .navigationTitle("Join Family")
            #if os(iOS)
            .navigationBarTitleDisplayMode(.inline)
            #endif
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Join", action: onJoin)
                        .disabled(familyId.isEmpty)
                }
            }
        }
        .presentationDetents([.medium])
    }
}

#Preview {
    FamilyListView()
        .environment(FamilyService.shared)
}
