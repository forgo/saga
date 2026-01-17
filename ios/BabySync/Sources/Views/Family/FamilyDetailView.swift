import SwiftUI

struct FamilyDetailView: View {
    let familyId: String

    @Environment(FamilyService.self) private var familyService
    @State private var showAddBaby = false
    @State private var newBabyName = ""

    var body: some View {
        Group {
            if familyService.isLoading && familyService.currentFamily == nil {
                ProgressView("Loading...")
            } else if let familyData = familyService.currentFamily {
                ScrollView {
                    LazyVStack(spacing: 24) {
                        // Family info section
                        FamilyInfoSection(family: familyData.family, parents: familyData.parents)

                        // Babies section
                        BabiesSection(
                            babies: familyData.babies,
                            activities: familyData.activities,
                            familyId: familyId,
                            onAddBaby: { showAddBaby = true }
                        )
                    }
                    .padding()
                }
            } else {
                ContentUnavailableView("Family Not Found", systemImage: "house.slash")
            }
        }
        .navigationTitle(familyService.currentFamily?.family.name ?? "Family")
        #if os(iOS)
        .navigationBarTitleDisplayMode(.large)
        #endif
        .sheet(isPresented: $showAddBaby) {
            AddBabySheet(babyName: $newBabyName) {
                Task {
                    try? await familyService.createBaby(name: newBabyName)
                    newBabyName = ""
                    showAddBaby = false
                }
            }
        }
        .task {
            try? await familyService.selectFamily(id: familyId)
        }
    }
}

struct FamilyInfoSection: View {
    let family: Family
    let parents: [Parent]

    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            Text("Family Members")
                .font(.headline)

            ForEach(parents) { parent in
                HStack {
                    Image(systemName: "person.circle.fill")
                        .font(.title2)
                        .foregroundStyle(.blue)

                    VStack(alignment: .leading) {
                        Text(parent.name)
                            .font(.subheadline)
                        Text(parent.email)
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                }
            }
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding()
        #if os(iOS)
        .background(Color(.systemGray6))
        #else
        .background(Color.gray.opacity(0.1))
        #endif
        .cornerRadius(12)
    }
}

struct BabiesSection: View {
    let babies: [Baby]
    let activities: [Activity]
    let familyId: String
    let onAddBaby: () -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            HStack {
                Text("Babies")
                    .font(.headline)
                Spacer()
                Button(action: onAddBaby) {
                    Image(systemName: "plus.circle.fill")
                        .font(.title2)
                }
            }

            if babies.isEmpty {
                Text("No babies yet. Add your first baby!")
                    .foregroundStyle(.secondary)
                    .frame(maxWidth: .infinity)
                    .padding()
            } else {
                ForEach(babies) { baby in
                    NavigationLink {
                        BabyDetailView(baby: baby, activities: activities, familyId: familyId)
                    } label: {
                        BabyCard(baby: baby)
                    }
                    .buttonStyle(.plain)
                }
            }
        }
        .frame(maxWidth: .infinity, alignment: .leading)
    }
}

struct BabyCard: View {
    let baby: Baby

    var body: some View {
        HStack {
            Image(systemName: "figure.child")
                .font(.largeTitle)
                .foregroundStyle(.pink)
                .frame(width: 50)

            VStack(alignment: .leading) {
                Text(baby.name)
                    .font(.headline)
                Text("Added \(baby.createdOn.formatted(date: .abbreviated, time: .omitted))")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }

            Spacer()

            Image(systemName: "chevron.right")
                .foregroundStyle(.secondary)
        }
        .padding()
        #if os(iOS)
        .background(Color(.systemGray6))
        #else
        .background(Color.gray.opacity(0.1))
        #endif
        .cornerRadius(12)
    }
}

struct AddBabySheet: View {
    @Binding var babyName: String
    let onAdd: () -> Void

    @Environment(\.dismiss) private var dismiss

    var body: some View {
        NavigationStack {
            Form {
                Section {
                    TextField("Baby's Name", text: $babyName)
                }
            }
            .navigationTitle("Add Baby")
            #if os(iOS)
            .navigationBarTitleDisplayMode(.inline)
            #endif
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Add", action: onAdd)
                        .disabled(babyName.isEmpty)
                }
            }
        }
        .presentationDetents([.medium])
    }
}

#Preview {
    NavigationStack {
        FamilyDetailView(familyId: "family:123")
    }
    .environment(FamilyService.shared)
}
