import Foundation

/// Centralized accessibility identifiers for UI testing
/// Pattern: {domain}_{element} or {domain}_{action}_{element}
enum AccessibilityID {

    // MARK: - Auth

    enum Auth {
        static let emailField = "auth_email_field"
        static let passwordField = "auth_password_field"
        static let submitButton = "auth_submit_button"
        static let errorMessage = "auth_error_message"
        static let modePicker = "auth_mode_picker"
        static let googleButton = "auth_google_button"
        static let appleButton = "auth_apple_button"
        static let passkeyButton = "auth_passkey_button"
        static let firstnameField = "auth_firstname_field"
        static let lastnameField = "auth_lastname_field"
    }

    // MARK: - Tab Bar

    enum TabBar {
        static let container = "tab_bar"
        static let guilds = "tab_guilds"
        static let events = "tab_events"
        static let discover = "tab_discover"
        static let profile = "tab_profile"
    }

    // MARK: - Guild

    enum Guild {
        static let list = "guild_list"
        static let createButton = "guild_create_button"
        static let emptyState = "guild_empty_state"
        static func row(id: String) -> String { "guild_row_\(id)" }

        // Create sheet
        static let nameField = "guild_name_field"
        static let descriptionField = "guild_description_field"
        static let iconPicker = "guild_icon_picker"
        static let colorPicker = "guild_color_picker"
        static let createConfirmButton = "guild_create_confirm"
        static let cancelButton = "guild_create_cancel"

        // Detail
        static let detailTitle = "guild_detail_title"
        static let detailDescription = "guild_detail_description"
        static let eventsLink = "guild_events_link"
        static let adventuresLink = "guild_adventures_link"
        static let poolsLink = "guild_pools_link"
        static let votesLink = "guild_votes_link"
        static let settingsLink = "guild_settings_link"
        static let addPersonButton = "guild_add_person"
        static let peopleSection = "guild_people_section"
        static let connectionStatus = "guild_connection_status"
    }

    // MARK: - Event

    enum Event {
        static let list = "event_list"
        static let createButton = "event_create_button"
        static let filterPicker = "event_filter_picker"
        static let emptyState = "event_empty_state"
        static func row(id: String) -> String { "event_row_\(id)" }

        // Detail
        static let title = "event_title"
        static let date = "event_date"
        static let location = "event_location"
        static let description = "event_description"
        static let rsvpGoing = "event_rsvp_going"
        static let rsvpMaybe = "event_rsvp_maybe"
        static let rsvpNotGoing = "event_rsvp_not_going"
        static let goingCount = "event_going_count"
        static let maybeCount = "event_maybe_count"
        static let rolesSection = "event_roles_section"
        static let addRoleButton = "event_add_role"
        static let editButton = "event_edit_button"
        static let cancelButton = "event_cancel_button"

        // Create sheet
        static let titleField = "event_title_field"
        static let descriptionField = "event_description_field"
        static let locationField = "event_location_field"
        static let startDatePicker = "event_start_date"
        static let endDatePicker = "event_end_date"
        static let capacityField = "event_capacity_field"
        static let visibilityPicker = "event_visibility_picker"
        static let createConfirmButton = "event_create_confirm"
        static let createCancelButton = "event_create_cancel"

        // Detail view
        static let detailView = "event_detail_view"
    }

    // MARK: - Discovery

    enum Discovery {
        static let tabPicker = "discover_tab_picker"
        static let peopleTab = "discover_people_tab"
        static let eventsTab = "discover_events_tab"
        static let interestsTab = "discover_interests_tab"
        static let compatibilitySlider = "discover_compatibility_slider"
        static let radiusSlider = "discover_radius_slider"
        static let searchButton = "discover_search_button"
        static let resultsList = "discover_results_list"
        static func resultRow(id: String) -> String { "discover_result_\(id)" }
    }

    // MARK: - Trust

    enum Trust {
        static let tabPicker = "trust_tab_picker"
        static let grantedTab = "trust_granted_tab"
        static let receivedTab = "trust_received_tab"
        static let irlTab = "trust_irl_tab"
        static let grantList = "trust_grant_list"
        static func grantRow(id: String) -> String { "trust_grant_\(id)" }
        static let confirmButton = "trust_confirm_button"
        static let declineButton = "trust_decline_button"
        static let grantTrustButton = "trust_grant_button"
    }

    // MARK: - Profile

    enum Profile {
        static let avatar = "profile_avatar"
        static let displayName = "profile_display_name"
        static let email = "profile_email"
        static let editButton = "profile_edit_button"
        static let logoutButton = "profile_logout_button"
        static let confirmLogoutButton = "profile_confirm_logout"
        static let availabilityLink = "profile_availability_link"
        static let nearbyLink = "profile_nearby_link"
        static let trustLink = "profile_trust_link"
        static let resonanceLink = "profile_resonance_link"
        static let settingsLink = "profile_settings_link"
    }

    // MARK: - Adventures

    enum Adventure {
        static let list = "adventure_list"
        static let createButton = "adventure_create_button"
        static let emptyState = "adventure_empty_state"
        static func row(id: String) -> String { "adventure_row_\(id)" }
        static let titleField = "adventure_title_field"
        static let descriptionField = "adventure_description_field"
        static let locationField = "adventure_location_field"
        static let statusBadge = "adventure_status_badge"
        static let createConfirmButton = "adventure_create_confirm"
        static let createCancelButton = "adventure_create_cancel"
    }

    // MARK: - Pools

    enum Pool {
        static let list = "pool_list"
        static let createButton = "pool_create_button"
        static let emptyState = "pool_empty_state"
        static func row(id: String) -> String { "pool_row_\(id)" }
        static let nameField = "pool_name_field"
        static let descriptionField = "pool_description_field"
        static let frequencyPicker = "pool_frequency_picker"
        static let joinButton = "pool_join_button"
        static let leaveButton = "pool_leave_button"
        static let createConfirmButton = "pool_create_confirm"
        static let createCancelButton = "pool_create_cancel"
    }

    // MARK: - Votes

    enum Vote {
        static let list = "vote_list"
        static let createButton = "vote_create_button"
        static let filterPicker = "vote_filter_picker"
        static let emptyState = "vote_empty_state"
        static func row(id: String) -> String { "vote_row_\(id)" }
        static let titleField = "vote_title_field"
        static let typePicker = "vote_type_picker"
        static let castButton = "vote_cast_button"
    }

    // MARK: - Common

    enum Common {
        static let loadingIndicator = "loading_indicator"
        static let emptyState = "empty_state"
        static let errorView = "error_view"
        static let retryButton = "retry_button"
        static let backButton = "back_button"
        static let doneButton = "done_button"
        static let cancelButton = "cancel_button"
        static let saveButton = "save_button"
    }
}
