package model

import (
	"strings"
	"testing"
	"time"
)

// ============================================================================
// CreateVoteRequest Tests
// ============================================================================

func TestCreateVoteRequest_Validate_Valid(t *testing.T) {
	t.Parallel()

	scopeID := "guild:123"
	req := &CreateVoteRequest{
		ScopeType: "guild",
		ScopeID:   &scopeID,
		Title:     "Vote Title",
		VoteType:  "fptp",
		OpensAt:   "2025-01-01T00:00:00Z",
		ClosesAt:  "2025-01-02T00:00:00Z",
	}

	errors := req.Validate()
	if len(errors) > 0 {
		t.Errorf("expected no errors, got %v", errors)
	}
}

func TestCreateVoteRequest_Validate_MissingScopeType(t *testing.T) {
	t.Parallel()

	req := &CreateVoteRequest{
		Title:    "Vote Title",
		VoteType: "fptp",
		OpensAt:  "2025-01-01T00:00:00Z",
		ClosesAt: "2025-01-02T00:00:00Z",
	}

	errors := req.Validate()
	if len(errors) != 1 || errors[0].Field != "scope_type" {
		t.Errorf("expected scope_type error, got %v", errors)
	}
}

func TestCreateVoteRequest_Validate_InvalidScopeType(t *testing.T) {
	t.Parallel()

	req := &CreateVoteRequest{
		ScopeType: "invalid",
		Title:     "Vote Title",
		VoteType:  "fptp",
		OpensAt:   "2025-01-01T00:00:00Z",
		ClosesAt:  "2025-01-02T00:00:00Z",
	}

	errors := req.Validate()
	hasError := false
	for _, e := range errors {
		if e.Field == "scope_type" && strings.Contains(e.Message, "guild") {
			hasError = true
		}
	}
	if !hasError {
		t.Errorf("expected scope_type validation error, got %v", errors)
	}
}

func TestCreateVoteRequest_Validate_GuildScopeRequiresScopeID(t *testing.T) {
	t.Parallel()

	req := &CreateVoteRequest{
		ScopeType: "guild",
		Title:     "Vote Title",
		VoteType:  "fptp",
		OpensAt:   "2025-01-01T00:00:00Z",
		ClosesAt:  "2025-01-02T00:00:00Z",
	}

	errors := req.Validate()
	hasError := false
	for _, e := range errors {
		if e.Field == "scope_id" {
			hasError = true
		}
	}
	if !hasError {
		t.Errorf("expected scope_id error for guild scope, got %v", errors)
	}
}

func TestCreateVoteRequest_Validate_GlobalScopeNoScopeIDRequired(t *testing.T) {
	t.Parallel()

	req := &CreateVoteRequest{
		ScopeType: "global",
		Title:     "Vote Title",
		VoteType:  "fptp",
		OpensAt:   "2025-01-01T00:00:00Z",
		ClosesAt:  "2025-01-02T00:00:00Z",
	}

	errors := req.Validate()
	for _, e := range errors {
		if e.Field == "scope_id" {
			t.Errorf("unexpected scope_id error for global scope: %v", e)
		}
	}
}

func TestCreateVoteRequest_Validate_MissingTitle(t *testing.T) {
	t.Parallel()

	req := &CreateVoteRequest{
		ScopeType: "global",
		VoteType:  "fptp",
		OpensAt:   "2025-01-01T00:00:00Z",
		ClosesAt:  "2025-01-02T00:00:00Z",
	}

	errors := req.Validate()
	hasError := false
	for _, e := range errors {
		if e.Field == "title" {
			hasError = true
		}
	}
	if !hasError {
		t.Errorf("expected title error, got %v", errors)
	}
}

func TestCreateVoteRequest_Validate_TitleTooLong(t *testing.T) {
	t.Parallel()

	req := &CreateVoteRequest{
		ScopeType: "global",
		Title:     strings.Repeat("a", MaxVoteTitleLength+1),
		VoteType:  "fptp",
		OpensAt:   "2025-01-01T00:00:00Z",
		ClosesAt:  "2025-01-02T00:00:00Z",
	}

	errors := req.Validate()
	hasError := false
	for _, e := range errors {
		if e.Field == "title" && strings.Contains(e.Message, "200") {
			hasError = true
		}
	}
	if !hasError {
		t.Errorf("expected title length error, got %v", errors)
	}
}

func TestCreateVoteRequest_Validate_DescriptionTooLong(t *testing.T) {
	t.Parallel()

	longDesc := strings.Repeat("a", MaxVoteDescriptionLength+1)
	req := &CreateVoteRequest{
		ScopeType:   "global",
		Title:       "Vote",
		Description: &longDesc,
		VoteType:    "fptp",
		OpensAt:     "2025-01-01T00:00:00Z",
		ClosesAt:    "2025-01-02T00:00:00Z",
	}

	errors := req.Validate()
	hasError := false
	for _, e := range errors {
		if e.Field == "description" {
			hasError = true
		}
	}
	if !hasError {
		t.Errorf("expected description length error, got %v", errors)
	}
}

func TestCreateVoteRequest_Validate_InvalidVoteType(t *testing.T) {
	t.Parallel()

	req := &CreateVoteRequest{
		ScopeType: "global",
		Title:     "Vote",
		VoteType:  "invalid",
		OpensAt:   "2025-01-01T00:00:00Z",
		ClosesAt:  "2025-01-02T00:00:00Z",
	}

	errors := req.Validate()
	hasError := false
	for _, e := range errors {
		if e.Field == "vote_type" {
			hasError = true
		}
	}
	if !hasError {
		t.Errorf("expected vote_type error, got %v", errors)
	}
}

func TestCreateVoteRequest_Validate_AllVoteTypes(t *testing.T) {
	t.Parallel()

	validTypes := []string{"fptp", "ranked_choice", "approval", "multi_select"}
	for _, vt := range validTypes {
		req := &CreateVoteRequest{
			ScopeType: "global",
			Title:     "Vote",
			VoteType:  vt,
			OpensAt:   "2025-01-01T00:00:00Z",
			ClosesAt:  "2025-01-02T00:00:00Z",
		}

		errors := req.Validate()
		for _, e := range errors {
			if e.Field == "vote_type" {
				t.Errorf("unexpected vote_type error for %s: %v", vt, e)
			}
		}
	}
}

func TestCreateVoteRequest_Validate_InvalidResultsVisibility(t *testing.T) {
	t.Parallel()

	invalid := "invalid"
	req := &CreateVoteRequest{
		ScopeType:         "global",
		Title:             "Vote",
		VoteType:          "fptp",
		OpensAt:           "2025-01-01T00:00:00Z",
		ClosesAt:          "2025-01-02T00:00:00Z",
		ResultsVisibility: &invalid,
	}

	errors := req.Validate()
	hasError := false
	for _, e := range errors {
		if e.Field == "results_visibility" {
			hasError = true
		}
	}
	if !hasError {
		t.Errorf("expected results_visibility error, got %v", errors)
	}
}

func TestCreateVoteRequest_Validate_MultiSelectMaxOptionsTooSmall(t *testing.T) {
	t.Parallel()

	maxOpts := 0
	req := &CreateVoteRequest{
		ScopeType:            "global",
		Title:                "Vote",
		VoteType:             "multi_select",
		OpensAt:              "2025-01-01T00:00:00Z",
		ClosesAt:             "2025-01-02T00:00:00Z",
		MaxOptionsSelectable: &maxOpts,
	}

	errors := req.Validate()
	hasError := false
	for _, e := range errors {
		if e.Field == "max_options_selectable" {
			hasError = true
		}
	}
	if !hasError {
		t.Errorf("expected max_options_selectable error, got %v", errors)
	}
}

// ============================================================================
// UpdateVoteRequest Tests
// ============================================================================

func TestUpdateVoteRequest_Validate_EmptyTitle(t *testing.T) {
	t.Parallel()

	empty := ""
	req := &UpdateVoteRequest{Title: &empty}

	errors := req.Validate()
	if len(errors) != 1 || errors[0].Field != "title" {
		t.Errorf("expected title error, got %v", errors)
	}
}

func TestUpdateVoteRequest_Validate_TitleTooLong(t *testing.T) {
	t.Parallel()

	long := strings.Repeat("a", MaxVoteTitleLength+1)
	req := &UpdateVoteRequest{Title: &long}

	errors := req.Validate()
	if len(errors) != 1 || errors[0].Field != "title" {
		t.Errorf("expected title length error, got %v", errors)
	}
}

func TestUpdateVoteRequest_Validate_InvalidResultsVisibility(t *testing.T) {
	t.Parallel()

	invalid := "invalid"
	req := &UpdateVoteRequest{ResultsVisibility: &invalid}

	errors := req.Validate()
	if len(errors) != 1 || errors[0].Field != "results_visibility" {
		t.Errorf("expected results_visibility error, got %v", errors)
	}
}

func TestUpdateVoteRequest_Validate_Valid(t *testing.T) {
	t.Parallel()

	title := "New Title"
	visibility := "live"
	req := &UpdateVoteRequest{
		Title:             &title,
		ResultsVisibility: &visibility,
	}

	errors := req.Validate()
	if len(errors) > 0 {
		t.Errorf("expected no errors, got %v", errors)
	}
}

// ============================================================================
// CreateVoteOptionRequest Tests
// ============================================================================

func TestCreateVoteOptionRequest_Validate_MissingText(t *testing.T) {
	t.Parallel()

	req := &CreateVoteOptionRequest{}

	errors := req.Validate()
	if len(errors) != 1 || errors[0].Field != "option_text" {
		t.Errorf("expected option_text error, got %v", errors)
	}
}

func TestCreateVoteOptionRequest_Validate_TextTooLong(t *testing.T) {
	t.Parallel()

	req := &CreateVoteOptionRequest{
		OptionText: strings.Repeat("a", MaxOptionTextLength+1),
	}

	errors := req.Validate()
	if len(errors) != 1 || errors[0].Field != "option_text" {
		t.Errorf("expected option_text length error, got %v", errors)
	}
}

func TestCreateVoteOptionRequest_Validate_DescriptionTooLong(t *testing.T) {
	t.Parallel()

	longDesc := strings.Repeat("a", MaxOptionDescLength+1)
	req := &CreateVoteOptionRequest{
		OptionText:        "Valid",
		OptionDescription: &longDesc,
	}

	errors := req.Validate()
	if len(errors) != 1 || errors[0].Field != "option_description" {
		t.Errorf("expected option_description error, got %v", errors)
	}
}

func TestCreateVoteOptionRequest_Validate_Valid(t *testing.T) {
	t.Parallel()

	desc := "Description"
	req := &CreateVoteOptionRequest{
		OptionText:        "Option A",
		OptionDescription: &desc,
	}

	errors := req.Validate()
	if len(errors) > 0 {
		t.Errorf("expected no errors, got %v", errors)
	}
}

// ============================================================================
// UpdateVoteOptionRequest Tests
// ============================================================================

func TestUpdateVoteOptionRequest_Validate_EmptyText(t *testing.T) {
	t.Parallel()

	empty := ""
	req := &UpdateVoteOptionRequest{OptionText: &empty}

	errors := req.Validate()
	if len(errors) != 1 || errors[0].Field != "option_text" {
		t.Errorf("expected option_text error, got %v", errors)
	}
}

func TestUpdateVoteOptionRequest_Validate_TextTooLong(t *testing.T) {
	t.Parallel()

	long := strings.Repeat("a", MaxOptionTextLength+1)
	req := &UpdateVoteOptionRequest{OptionText: &long}

	errors := req.Validate()
	if len(errors) != 1 || errors[0].Field != "option_text" {
		t.Errorf("expected option_text length error, got %v", errors)
	}
}

// ============================================================================
// CastBallotRequest Tests
// ============================================================================

func TestCastBallotRequest_ValidateForVoteType_FPTP_Valid(t *testing.T) {
	t.Parallel()

	optionID := "option:1"
	req := &CastBallotRequest{OptionID: &optionID}

	errors := req.ValidateForVoteType(VoteTypeFPTP, nil, false)
	if len(errors) > 0 {
		t.Errorf("expected no errors, got %v", errors)
	}
}

func TestCastBallotRequest_ValidateForVoteType_FPTP_MissingOption(t *testing.T) {
	t.Parallel()

	req := &CastBallotRequest{}

	errors := req.ValidateForVoteType(VoteTypeFPTP, nil, false)
	if len(errors) != 1 || errors[0].Field != "option_id" {
		t.Errorf("expected option_id error, got %v", errors)
	}
}

func TestCastBallotRequest_ValidateForVoteType_RankedChoice_Valid(t *testing.T) {
	t.Parallel()

	req := &CastBallotRequest{
		Rankings: []string{"opt:1", "opt:2", "opt:3"},
	}

	errors := req.ValidateForVoteType(VoteTypeRankedChoice, nil, false)
	if len(errors) > 0 {
		t.Errorf("expected no errors, got %v", errors)
	}
}

func TestCastBallotRequest_ValidateForVoteType_RankedChoice_MissingRankings(t *testing.T) {
	t.Parallel()

	req := &CastBallotRequest{}

	errors := req.ValidateForVoteType(VoteTypeRankedChoice, nil, false)
	if len(errors) != 1 || errors[0].Field != "rankings" {
		t.Errorf("expected rankings error, got %v", errors)
	}
}

func TestCastBallotRequest_ValidateForVoteType_Approval_Valid(t *testing.T) {
	t.Parallel()

	req := &CastBallotRequest{
		SelectedOptions: []string{"opt:1", "opt:2"},
	}

	errors := req.ValidateForVoteType(VoteTypeApproval, nil, false)
	if len(errors) > 0 {
		t.Errorf("expected no errors, got %v", errors)
	}
}

func TestCastBallotRequest_ValidateForVoteType_Approval_MissingOptions(t *testing.T) {
	t.Parallel()

	req := &CastBallotRequest{}

	errors := req.ValidateForVoteType(VoteTypeApproval, nil, false)
	if len(errors) != 1 || errors[0].Field != "selected_options" {
		t.Errorf("expected selected_options error, got %v", errors)
	}
}

func TestCastBallotRequest_ValidateForVoteType_MultiSelect_TooManyOptions(t *testing.T) {
	t.Parallel()

	max := 2
	req := &CastBallotRequest{
		SelectedOptions: []string{"opt:1", "opt:2", "opt:3"},
	}

	errors := req.ValidateForVoteType(VoteTypeMultiSelect, &max, false)
	if len(errors) != 1 || errors[0].Field != "selected_options" {
		t.Errorf("expected selected_options error, got %v", errors)
	}
}

func TestCastBallotRequest_ValidateForVoteType_Abstain_Allowed(t *testing.T) {
	t.Parallel()

	req := &CastBallotRequest{IsAbstain: true}

	errors := req.ValidateForVoteType(VoteTypeFPTP, nil, true)
	if len(errors) > 0 {
		t.Errorf("expected no errors for allowed abstain, got %v", errors)
	}
}

func TestCastBallotRequest_ValidateForVoteType_Abstain_NotAllowed(t *testing.T) {
	t.Parallel()

	req := &CastBallotRequest{IsAbstain: true}

	errors := req.ValidateForVoteType(VoteTypeFPTP, nil, false)
	if len(errors) != 1 || errors[0].Field != "is_abstain" {
		t.Errorf("expected is_abstain error, got %v", errors)
	}
}

func TestCastBallotRequest_ToBallotData_FPTP(t *testing.T) {
	t.Parallel()

	optionID := "option:1"
	req := &CastBallotRequest{OptionID: &optionID}

	data := req.ToBallotData(VoteTypeFPTP)
	if data["option_id"] != "option:1" {
		t.Errorf("expected option_id in ballot data, got %v", data)
	}
}

func TestCastBallotRequest_ToBallotData_RankedChoice(t *testing.T) {
	t.Parallel()

	req := &CastBallotRequest{Rankings: []string{"a", "b", "c"}}

	data := req.ToBallotData(VoteTypeRankedChoice)
	rankings, ok := data["rankings"].([]string)
	if !ok || len(rankings) != 3 {
		t.Errorf("expected rankings in ballot data, got %v", data)
	}
}

func TestCastBallotRequest_ToBallotData_Abstain(t *testing.T) {
	t.Parallel()

	req := &CastBallotRequest{IsAbstain: true}

	data := req.ToBallotData(VoteTypeFPTP)
	if len(data) != 0 {
		t.Errorf("expected empty ballot data for abstain, got %v", data)
	}
}

// ============================================================================
// CreateAdventureRequest Tests
// ============================================================================

func TestCreateAdventureRequest_Validate_Valid(t *testing.T) {
	t.Parallel()

	guildID := "guild:123"
	req := &CreateAdventureRequest{
		GuildID:   &guildID,
		Title:     "Adventure",
		StartDate: "2025-01-01T00:00:00Z",
		EndDate:   "2025-01-07T00:00:00Z",
	}

	errors := req.Validate()
	if len(errors) > 0 {
		t.Errorf("expected no errors, got %v", errors)
	}
}

func TestCreateAdventureRequest_Validate_MissingTitle(t *testing.T) {
	t.Parallel()

	req := &CreateAdventureRequest{
		StartDate: "2025-01-01T00:00:00Z",
		EndDate:   "2025-01-07T00:00:00Z",
	}

	errors := req.Validate()
	hasError := false
	for _, e := range errors {
		if e.Field == "title" {
			hasError = true
		}
	}
	if !hasError {
		t.Errorf("expected title error, got %v", errors)
	}
}

func TestCreateAdventureRequest_Validate_TitleTooLong(t *testing.T) {
	t.Parallel()

	req := &CreateAdventureRequest{
		Title:     strings.Repeat("a", MaxAdventureTitleLength+1),
		StartDate: "2025-01-01T00:00:00Z",
		EndDate:   "2025-01-07T00:00:00Z",
	}

	errors := req.Validate()
	hasError := false
	for _, e := range errors {
		if e.Field == "title" && strings.Contains(e.Message, "100") {
			hasError = true
		}
	}
	if !hasError {
		t.Errorf("expected title length error, got %v", errors)
	}
}

func TestCreateAdventureRequest_Validate_DescriptionTooLong(t *testing.T) {
	t.Parallel()

	longDesc := strings.Repeat("a", MaxAdventureDescLength+1)
	guildID := "guild:123"
	req := &CreateAdventureRequest{
		GuildID:     &guildID,
		Title:       "Adventure",
		Description: &longDesc,
		StartDate:   "2025-01-01T00:00:00Z",
		EndDate:     "2025-01-07T00:00:00Z",
	}

	errors := req.Validate()
	hasError := false
	for _, e := range errors {
		if e.Field == "description" {
			hasError = true
		}
	}
	if !hasError {
		t.Errorf("expected description error, got %v", errors)
	}
}

func TestCreateAdventureRequest_Validate_InvalidOrganizerType(t *testing.T) {
	t.Parallel()

	invalid := "invalid"
	req := &CreateAdventureRequest{
		OrganizerType: &invalid,
		Title:         "Adventure",
		StartDate:     "2025-01-01T00:00:00Z",
		EndDate:       "2025-01-07T00:00:00Z",
	}

	errors := req.Validate()
	hasError := false
	for _, e := range errors {
		if e.Field == "organizer_type" {
			hasError = true
		}
	}
	if !hasError {
		t.Errorf("expected organizer_type error, got %v", errors)
	}
}

func TestCreateAdventureRequest_Validate_GuildOrganizerRequiresGuildID(t *testing.T) {
	t.Parallel()

	orgType := "guild"
	req := &CreateAdventureRequest{
		OrganizerType: &orgType,
		Title:         "Adventure",
		StartDate:     "2025-01-01T00:00:00Z",
		EndDate:       "2025-01-07T00:00:00Z",
	}

	errors := req.Validate()
	hasError := false
	for _, e := range errors {
		if e.Field == "guild_id" {
			hasError = true
		}
	}
	if !hasError {
		t.Errorf("expected guild_id error, got %v", errors)
	}
}

func TestCreateAdventureRequest_GetOrganizerType_DefaultsToGuildWithGuildID(t *testing.T) {
	t.Parallel()

	guildID := "guild:123"
	req := &CreateAdventureRequest{GuildID: &guildID}

	if req.GetOrganizerType() != AdventureOrganizerGuild {
		t.Errorf("expected guild organizer type, got %v", req.GetOrganizerType())
	}
}

func TestCreateAdventureRequest_GetOrganizerType_DefaultsToUserWithoutGuildID(t *testing.T) {
	t.Parallel()

	req := &CreateAdventureRequest{}

	if req.GetOrganizerType() != AdventureOrganizerUser {
		t.Errorf("expected user organizer type, got %v", req.GetOrganizerType())
	}
}

func TestCreateAdventureRequest_GetOrganizerType_ExplicitType(t *testing.T) {
	t.Parallel()

	orgType := "user"
	guildID := "guild:123"
	req := &CreateAdventureRequest{
		OrganizerType: &orgType,
		GuildID:       &guildID,
	}

	if req.GetOrganizerType() != AdventureOrganizerUser {
		t.Errorf("expected explicit user type, got %v", req.GetOrganizerType())
	}
}

// ============================================================================
// RespondToAdmissionRequest Tests
// ============================================================================

func TestRespondToAdmissionRequest_Validate_AdmitNoReason(t *testing.T) {
	t.Parallel()

	req := &RespondToAdmissionRequest{Admit: true}

	errors := req.Validate()
	if len(errors) > 0 {
		t.Errorf("expected no errors for admit, got %v", errors)
	}
}

func TestRespondToAdmissionRequest_Validate_RejectRequiresReason(t *testing.T) {
	t.Parallel()

	req := &RespondToAdmissionRequest{Admit: false}

	errors := req.Validate()
	if len(errors) != 1 || errors[0].Field != "rejection_reason" {
		t.Errorf("expected rejection_reason error, got %v", errors)
	}
}

func TestRespondToAdmissionRequest_Validate_RejectWithReason(t *testing.T) {
	t.Parallel()

	reason := "Not enough experience"
	req := &RespondToAdmissionRequest{
		Admit:           false,
		RejectionReason: &reason,
	}

	errors := req.Validate()
	if len(errors) > 0 {
		t.Errorf("expected no errors, got %v", errors)
	}
}

func TestRespondToAdmissionRequest_Validate_ReasonTooLong(t *testing.T) {
	t.Parallel()

	reason := strings.Repeat("a", MaxRejectionReasonLength+1)
	req := &RespondToAdmissionRequest{
		Admit:           false,
		RejectionReason: &reason,
	}

	errors := req.Validate()
	hasError := false
	for _, e := range errors {
		if e.Field == "rejection_reason" && strings.Contains(e.Message, "500") {
			hasError = true
		}
	}
	if !hasError {
		t.Errorf("expected rejection_reason length error, got %v", errors)
	}
}

// ============================================================================
// InviteToAdventureRequest Tests
// ============================================================================

func TestInviteToAdventureRequest_Validate_MissingUserID(t *testing.T) {
	t.Parallel()

	req := &InviteToAdventureRequest{}

	errors := req.Validate()
	if len(errors) != 1 || errors[0].Field != "user_id" {
		t.Errorf("expected user_id error, got %v", errors)
	}
}

func TestInviteToAdventureRequest_Validate_Valid(t *testing.T) {
	t.Parallel()

	req := &InviteToAdventureRequest{UserID: "user:123"}

	errors := req.Validate()
	if len(errors) > 0 {
		t.Errorf("expected no errors, got %v", errors)
	}
}

// ============================================================================
// TransferAdventureRequest Tests
// ============================================================================

func TestTransferAdventureRequest_Validate_MissingNewOrganizer(t *testing.T) {
	t.Parallel()

	req := &TransferAdventureRequest{}

	errors := req.Validate()
	if len(errors) != 1 || errors[0].Field != "new_organizer_user_id" {
		t.Errorf("expected new_organizer_user_id error, got %v", errors)
	}
}

func TestTransferAdventureRequest_Validate_Valid(t *testing.T) {
	t.Parallel()

	req := &TransferAdventureRequest{NewOrganizerUserID: "user:123"}

	errors := req.Validate()
	if len(errors) > 0 {
		t.Errorf("expected no errors, got %v", errors)
	}
}

// ============================================================================
// UnfreezeAdventureRequest Tests
// ============================================================================

func TestUnfreezeAdventureRequest_Validate_MissingNewOrganizer(t *testing.T) {
	t.Parallel()

	req := &UnfreezeAdventureRequest{}

	errors := req.Validate()
	if len(errors) != 1 || errors[0].Field != "new_organizer_user_id" {
		t.Errorf("expected new_organizer_user_id error, got %v", errors)
	}
}

// ============================================================================
// CreateTrustRatingRequest Tests
// ============================================================================

func TestCreateTrustRatingRequest_Validate_Valid(t *testing.T) {
	t.Parallel()

	req := &CreateTrustRatingRequest{
		RateeID:     "user:123",
		AnchorType:  "event",
		AnchorID:    "event:456",
		TrustLevel:  "trust",
		TrustReview: "Great person!",
	}

	errors := req.Validate()
	if len(errors) > 0 {
		t.Errorf("expected no errors, got %v", errors)
	}
}

func TestCreateTrustRatingRequest_Validate_MissingRateeID(t *testing.T) {
	t.Parallel()

	req := &CreateTrustRatingRequest{
		AnchorType:  "event",
		AnchorID:    "event:456",
		TrustLevel:  "trust",
		TrustReview: "Great person!",
	}

	errors := req.Validate()
	hasError := false
	for _, e := range errors {
		if e.Field == "ratee_id" {
			hasError = true
		}
	}
	if !hasError {
		t.Errorf("expected ratee_id error, got %v", errors)
	}
}

func TestCreateTrustRatingRequest_Validate_InvalidAnchorType(t *testing.T) {
	t.Parallel()

	req := &CreateTrustRatingRequest{
		RateeID:     "user:123",
		AnchorType:  "invalid",
		AnchorID:    "event:456",
		TrustLevel:  "trust",
		TrustReview: "Great!",
	}

	errors := req.Validate()
	hasError := false
	for _, e := range errors {
		if e.Field == "anchor_type" {
			hasError = true
		}
	}
	if !hasError {
		t.Errorf("expected anchor_type error, got %v", errors)
	}
}

func TestCreateTrustRatingRequest_Validate_InvalidTrustLevel(t *testing.T) {
	t.Parallel()

	req := &CreateTrustRatingRequest{
		RateeID:     "user:123",
		AnchorType:  "event",
		AnchorID:    "event:456",
		TrustLevel:  "neutral",
		TrustReview: "OK person",
	}

	errors := req.Validate()
	hasError := false
	for _, e := range errors {
		if e.Field == "trust_level" {
			hasError = true
		}
	}
	if !hasError {
		t.Errorf("expected trust_level error, got %v", errors)
	}
}

func TestCreateTrustRatingRequest_Validate_MissingReview(t *testing.T) {
	t.Parallel()

	req := &CreateTrustRatingRequest{
		RateeID:    "user:123",
		AnchorType: "event",
		AnchorID:   "event:456",
		TrustLevel: "trust",
	}

	errors := req.Validate()
	hasError := false
	for _, e := range errors {
		if e.Field == "trust_review" {
			hasError = true
		}
	}
	if !hasError {
		t.Errorf("expected trust_review error, got %v", errors)
	}
}

func TestCreateTrustRatingRequest_Validate_ReviewTooLong(t *testing.T) {
	t.Parallel()

	req := &CreateTrustRatingRequest{
		RateeID:     "user:123",
		AnchorType:  "event",
		AnchorID:    "event:456",
		TrustLevel:  "trust",
		TrustReview: strings.Repeat("a", MaxTrustReviewLength+1),
	}

	errors := req.Validate()
	hasError := false
	for _, e := range errors {
		if e.Field == "trust_review" && strings.Contains(e.Message, "240") {
			hasError = true
		}
	}
	if !hasError {
		t.Errorf("expected trust_review length error, got %v", errors)
	}
}

// ============================================================================
// UpdateTrustRatingRequest Tests
// ============================================================================

func TestUpdateTrustRatingRequest_Validate_InvalidTrustLevel(t *testing.T) {
	t.Parallel()

	invalid := "neutral"
	req := &UpdateTrustRatingRequest{TrustLevel: &invalid}

	errors := req.Validate()
	if len(errors) != 1 || errors[0].Field != "trust_level" {
		t.Errorf("expected trust_level error, got %v", errors)
	}
}

func TestUpdateTrustRatingRequest_Validate_EmptyReview(t *testing.T) {
	t.Parallel()

	empty := ""
	req := &UpdateTrustRatingRequest{TrustReview: &empty}

	errors := req.Validate()
	if len(errors) != 1 || errors[0].Field != "trust_review" {
		t.Errorf("expected trust_review error, got %v", errors)
	}
}

func TestUpdateTrustRatingRequest_Validate_ReviewTooLong(t *testing.T) {
	t.Parallel()

	long := strings.Repeat("a", MaxTrustReviewLength+1)
	req := &UpdateTrustRatingRequest{TrustReview: &long}

	errors := req.Validate()
	if len(errors) != 1 || errors[0].Field != "trust_review" {
		t.Errorf("expected trust_review length error, got %v", errors)
	}
}

func TestUpdateTrustRatingRequest_Validate_Valid(t *testing.T) {
	t.Parallel()

	level := "distrust"
	review := "Changed my mind"
	req := &UpdateTrustRatingRequest{
		TrustLevel:  &level,
		TrustReview: &review,
	}

	errors := req.Validate()
	if len(errors) > 0 {
		t.Errorf("expected no errors, got %v", errors)
	}
}

// ============================================================================
// CreateEndorsementRequest Tests
// ============================================================================

func TestCreateEndorsementRequest_Validate_MissingType(t *testing.T) {
	t.Parallel()

	req := &CreateEndorsementRequest{}

	errors := req.Validate()
	if len(errors) != 1 || errors[0].Field != "endorsement_type" {
		t.Errorf("expected endorsement_type error, got %v", errors)
	}
}

func TestCreateEndorsementRequest_Validate_InvalidType(t *testing.T) {
	t.Parallel()

	req := &CreateEndorsementRequest{EndorsementType: "neutral"}

	errors := req.Validate()
	if len(errors) != 1 || errors[0].Field != "endorsement_type" {
		t.Errorf("expected endorsement_type error, got %v", errors)
	}
}

func TestCreateEndorsementRequest_Validate_NoteTooLong(t *testing.T) {
	t.Parallel()

	long := strings.Repeat("a", MaxEndorsementNoteLength+1)
	req := &CreateEndorsementRequest{
		EndorsementType: "agree",
		Note:            &long,
	}

	errors := req.Validate()
	if len(errors) != 1 || errors[0].Field != "note" {
		t.Errorf("expected note length error, got %v", errors)
	}
}

func TestCreateEndorsementRequest_Validate_Valid(t *testing.T) {
	t.Parallel()

	note := "I was there too"
	req := &CreateEndorsementRequest{
		EndorsementType: "agree",
		Note:            &note,
	}

	errors := req.Validate()
	if len(errors) > 0 {
		t.Errorf("expected no errors, got %v", errors)
	}
}

// ============================================================================
// CreateRSVPRequest Tests
// ============================================================================

func TestCreateRSVPRequest_Validate_MissingTargetType(t *testing.T) {
	t.Parallel()

	req := &CreateRSVPRequest{TargetID: "event:123"}

	errors := req.Validate()
	hasError := false
	for _, e := range errors {
		if e.Field == "target_type" {
			hasError = true
		}
	}
	if !hasError {
		t.Errorf("expected target_type error, got %v", errors)
	}
}

func TestCreateRSVPRequest_Validate_InvalidTargetType(t *testing.T) {
	t.Parallel()

	req := &CreateRSVPRequest{
		TargetType: "invalid",
		TargetID:   "thing:123",
	}

	errors := req.Validate()
	hasError := false
	for _, e := range errors {
		if e.Field == "target_type" && strings.Contains(e.Message, "invalid") {
			hasError = true
		}
	}
	if !hasError {
		t.Errorf("expected target_type validation error, got %v", errors)
	}
}

func TestCreateRSVPRequest_Validate_MissingTargetID(t *testing.T) {
	t.Parallel()

	req := &CreateRSVPRequest{TargetType: "event"}

	errors := req.Validate()
	hasError := false
	for _, e := range errors {
		if e.Field == "target_id" {
			hasError = true
		}
	}
	if !hasError {
		t.Errorf("expected target_id error, got %v", errors)
	}
}

func TestCreateRSVPRequest_Validate_TooManyPlusOnes(t *testing.T) {
	t.Parallel()

	plusOnes := MaxPlusOnes + 1
	req := &CreateRSVPRequest{
		TargetType: "event",
		TargetID:   "event:123",
		PlusOnes:   &plusOnes,
	}

	errors := req.Validate()
	hasError := false
	for _, e := range errors {
		if e.Field == "plus_ones" {
			hasError = true
		}
	}
	if !hasError {
		t.Errorf("expected plus_ones error, got %v", errors)
	}
}

func TestCreateRSVPRequest_Validate_NoteTooLong(t *testing.T) {
	t.Parallel()

	long := strings.Repeat("a", MaxRSVPNote+1)
	req := &CreateRSVPRequest{
		TargetType: "event",
		TargetID:   "event:123",
		Note:       &long,
	}

	errors := req.Validate()
	hasError := false
	for _, e := range errors {
		if e.Field == "note" {
			hasError = true
		}
	}
	if !hasError {
		t.Errorf("expected note error, got %v", errors)
	}
}

func TestCreateRSVPRequest_Validate_Valid(t *testing.T) {
	t.Parallel()

	req := &CreateRSVPRequest{
		TargetType: "event",
		TargetID:   "event:123",
	}

	errors := req.Validate()
	if len(errors) > 0 {
		t.Errorf("expected no errors, got %v", errors)
	}
}

// ============================================================================
// RSVPFeedbackRequest Tests
// ============================================================================

func TestRSVPFeedbackRequest_Validate_InvalidRating(t *testing.T) {
	t.Parallel()

	req := &RSVPFeedbackRequest{HelpfulnessRating: "invalid"}

	errors := req.Validate()
	hasError := false
	for _, e := range errors {
		if e.Field == "helpfulness_rating" {
			hasError = true
		}
	}
	if !hasError {
		t.Errorf("expected helpfulness_rating error, got %v", errors)
	}
}

func TestRSVPFeedbackRequest_Validate_TooManyTags(t *testing.T) {
	t.Parallel()

	req := &RSVPFeedbackRequest{
		HelpfulnessRating: "YES",
		Tags:              make([]string, MaxHelpfulTags+1),
	}

	errors := req.Validate()
	hasError := false
	for _, e := range errors {
		if e.Field == "tags" && strings.Contains(e.Message, "5") {
			hasError = true
		}
	}
	if !hasError {
		t.Errorf("expected tags count error, got %v", errors)
	}
}

func TestRSVPFeedbackRequest_Validate_TagTooLong(t *testing.T) {
	t.Parallel()

	req := &RSVPFeedbackRequest{
		HelpfulnessRating: "YES",
		Tags:              []string{strings.Repeat("a", MaxTagLength+1)},
	}

	errors := req.Validate()
	hasError := false
	for _, e := range errors {
		if e.Field == "tags" && strings.Contains(e.Message, "too long") {
			hasError = true
		}
	}
	if !hasError {
		t.Errorf("expected tag length error, got %v", errors)
	}
}

func TestRSVPFeedbackRequest_Validate_ValidRatings(t *testing.T) {
	t.Parallel()

	validRatings := []string{"YES", "SOMEWHAT", "NOT_REALLY", "SKIP"}
	for _, rating := range validRatings {
		req := &RSVPFeedbackRequest{HelpfulnessRating: rating}

		errors := req.Validate()
		for _, e := range errors {
			if e.Field == "helpfulness_rating" {
				t.Errorf("unexpected helpfulness_rating error for %s: %v", rating, e)
			}
		}
	}
}

// ============================================================================
// CreateRoleCatalogRequest Tests
// ============================================================================

func TestCreateRoleCatalogRequest_Validate_MissingRoleType(t *testing.T) {
	t.Parallel()

	req := &CreateRoleCatalogRequest{Name: "DJ"}

	errors := req.Validate()
	hasError := false
	for _, e := range errors {
		if e.Field == "role_type" {
			hasError = true
		}
	}
	if !hasError {
		t.Errorf("expected role_type error, got %v", errors)
	}
}

func TestCreateRoleCatalogRequest_Validate_InvalidRoleType(t *testing.T) {
	t.Parallel()

	req := &CreateRoleCatalogRequest{
		RoleType: "invalid",
		Name:     "DJ",
	}

	errors := req.Validate()
	hasError := false
	for _, e := range errors {
		if e.Field == "role_type" {
			hasError = true
		}
	}
	if !hasError {
		t.Errorf("expected role_type error, got %v", errors)
	}
}

func TestCreateRoleCatalogRequest_Validate_MissingName(t *testing.T) {
	t.Parallel()

	req := &CreateRoleCatalogRequest{RoleType: "event"}

	errors := req.Validate()
	hasError := false
	for _, e := range errors {
		if e.Field == "name" {
			hasError = true
		}
	}
	if !hasError {
		t.Errorf("expected name error, got %v", errors)
	}
}

func TestCreateRoleCatalogRequest_Validate_NameTooLong(t *testing.T) {
	t.Parallel()

	req := &CreateRoleCatalogRequest{
		RoleType: "event",
		Name:     strings.Repeat("a", MaxRoleCatalogNameLength+1),
	}

	errors := req.Validate()
	hasError := false
	for _, e := range errors {
		if e.Field == "name" && strings.Contains(e.Message, "50") {
			hasError = true
		}
	}
	if !hasError {
		t.Errorf("expected name length error, got %v", errors)
	}
}

func TestCreateRoleCatalogRequest_Validate_DescriptionTooLong(t *testing.T) {
	t.Parallel()

	long := strings.Repeat("a", MaxRoleCatalogDescLength+1)
	req := &CreateRoleCatalogRequest{
		RoleType:    "event",
		Name:        "DJ",
		Description: &long,
	}

	errors := req.Validate()
	hasError := false
	for _, e := range errors {
		if e.Field == "description" {
			hasError = true
		}
	}
	if !hasError {
		t.Errorf("expected description error, got %v", errors)
	}
}

func TestCreateRoleCatalogRequest_Validate_Valid(t *testing.T) {
	t.Parallel()

	desc := "Plays music"
	req := &CreateRoleCatalogRequest{
		RoleType:    "event",
		Name:        "DJ",
		Description: &desc,
	}

	errors := req.Validate()
	if len(errors) > 0 {
		t.Errorf("expected no errors, got %v", errors)
	}
}

// ============================================================================
// UpdateRoleCatalogRequest Tests
// ============================================================================

func TestUpdateRoleCatalogRequest_Validate_EmptyName(t *testing.T) {
	t.Parallel()

	empty := ""
	req := &UpdateRoleCatalogRequest{Name: &empty}

	errors := req.Validate()
	if len(errors) != 1 || errors[0].Field != "name" {
		t.Errorf("expected name error, got %v", errors)
	}
}

func TestUpdateRoleCatalogRequest_Validate_NameTooLong(t *testing.T) {
	t.Parallel()

	long := strings.Repeat("a", MaxRoleCatalogNameLength+1)
	req := &UpdateRoleCatalogRequest{Name: &long}

	errors := req.Validate()
	if len(errors) != 1 || errors[0].Field != "name" {
		t.Errorf("expected name length error, got %v", errors)
	}
}

// ============================================================================
// CreateRoleFromCatalogRequest Tests
// ============================================================================

func TestCreateRoleFromCatalogRequest_Validate_MissingCatalogID(t *testing.T) {
	t.Parallel()

	req := &CreateRoleFromCatalogRequest{}

	errors := req.Validate()
	if len(errors) != 1 || errors[0].Field != "catalog_role_id" {
		t.Errorf("expected catalog_role_id error, got %v", errors)
	}
}

func TestCreateRoleFromCatalogRequest_Validate_NegativeMaxSlots(t *testing.T) {
	t.Parallel()

	negative := -1
	req := &CreateRoleFromCatalogRequest{
		CatalogRoleID: "role:123",
		MaxSlots:      &negative,
	}

	errors := req.Validate()
	if len(errors) != 1 || errors[0].Field != "max_slots" {
		t.Errorf("expected max_slots error, got %v", errors)
	}
}

// ============================================================================
// CreateRideshareRoleRequest Tests
// ============================================================================

func TestCreateRideshareRoleRequest_Validate_MissingName(t *testing.T) {
	t.Parallel()

	req := &CreateRideshareRoleRequest{}

	errors := req.Validate()
	hasError := false
	for _, e := range errors {
		if e.Field == "name" {
			hasError = true
		}
	}
	if !hasError {
		t.Errorf("expected name error, got %v", errors)
	}
}

func TestCreateRideshareRoleRequest_Validate_NegativeMaxSlots(t *testing.T) {
	t.Parallel()

	req := &CreateRideshareRoleRequest{
		Name:     "Driver",
		MaxSlots: -1,
	}

	errors := req.Validate()
	hasError := false
	for _, e := range errors {
		if e.Field == "max_slots" {
			hasError = true
		}
	}
	if !hasError {
		t.Errorf("expected max_slots error, got %v", errors)
	}
}

// ============================================================================
// UpdateRideshareRoleRequest Tests
// ============================================================================

func TestUpdateRideshareRoleRequest_Validate_EmptyName(t *testing.T) {
	t.Parallel()

	empty := ""
	req := &UpdateRideshareRoleRequest{Name: &empty}

	errors := req.Validate()
	if len(errors) != 1 || errors[0].Field != "name" {
		t.Errorf("expected name error, got %v", errors)
	}
}

func TestUpdateRideshareRoleRequest_Validate_NegativeMaxSlots(t *testing.T) {
	t.Parallel()

	negative := -1
	req := &UpdateRideshareRoleRequest{MaxSlots: &negative}

	errors := req.Validate()
	if len(errors) != 1 || errors[0].Field != "max_slots" {
		t.Errorf("expected max_slots error, got %v", errors)
	}
}

// ============================================================================
// AssignRideshareRoleRequest Tests
// ============================================================================

func TestAssignRideshareRoleRequest_Validate_MissingRoleID(t *testing.T) {
	t.Parallel()

	req := &AssignRideshareRoleRequest{}

	errors := req.Validate()
	if len(errors) != 1 || errors[0].Field != "role_id" {
		t.Errorf("expected role_id error, got %v", errors)
	}
}

func TestAssignRideshareRoleRequest_Validate_NoteTooLong(t *testing.T) {
	t.Parallel()

	long := strings.Repeat("a", MaxRideshareRoleNoteLength+1)
	req := &AssignRideshareRoleRequest{
		RoleID: "role:123",
		Note:   &long,
	}

	errors := req.Validate()
	if len(errors) != 1 || errors[0].Field != "note" {
		t.Errorf("expected note length error, got %v", errors)
	}
}

func TestAssignRideshareRoleRequest_Validate_Valid(t *testing.T) {
	t.Parallel()

	note := "Will bring snacks"
	req := &AssignRideshareRoleRequest{
		RoleID: "role:123",
		Note:   &note,
	}

	errors := req.Validate()
	if len(errors) > 0 {
		t.Errorf("expected no errors, got %v", errors)
	}
}

// ============================================================================
// Profile Helper Tests
// ============================================================================

func TestGetActivityStatus_Nil(t *testing.T) {
	t.Parallel()

	status := GetActivityStatus(nil)
	if status != ActivityStatusAway {
		t.Errorf("expected away for nil, got %v", status)
	}
}

func TestGetActivityStatus_Now(t *testing.T) {
	t.Parallel()

	now := time.Now()
	status := GetActivityStatus(&now)
	if status != ActivityStatusNow {
		t.Errorf("expected active_now, got %v", status)
	}
}

func TestGetActivityStatus_Recently(t *testing.T) {
	t.Parallel()

	recent := time.Now().Add(-15 * time.Minute)
	status := GetActivityStatus(&recent)
	if status != ActivityStatusRecently {
		t.Errorf("expected active_recently, got %v", status)
	}
}

func TestGetActivityStatus_ThisHour(t *testing.T) {
	t.Parallel()

	thisHour := time.Now().Add(-45 * time.Minute)
	status := GetActivityStatus(&thisHour)
	if status != ActivityStatusThisHour {
		t.Errorf("expected active_this_hour, got %v", status)
	}
}

func TestGetActivityStatus_Today(t *testing.T) {
	t.Parallel()

	today := time.Now().Add(-12 * time.Hour)
	status := GetActivityStatus(&today)
	if status != ActivityStatusToday {
		t.Errorf("expected active_today, got %v", status)
	}
}

func TestGetActivityStatus_Yesterday(t *testing.T) {
	t.Parallel()

	yesterday := time.Now().Add(-36 * time.Hour)
	status := GetActivityStatus(&yesterday)
	if status != ActivityStatusYesterday {
		t.Errorf("expected active_yesterday, got %v", status)
	}
}

func TestGetActivityStatus_ThisWeek(t *testing.T) {
	t.Parallel()

	thisWeek := time.Now().Add(-4 * 24 * time.Hour)
	status := GetActivityStatus(&thisWeek)
	if status != ActivityStatusThisWeek {
		t.Errorf("expected active_this_week, got %v", status)
	}
}

func TestGetActivityStatus_Away(t *testing.T) {
	t.Parallel()

	away := time.Now().Add(-10 * 24 * time.Hour)
	status := GetActivityStatus(&away)
	if status != ActivityStatusAway {
		t.Errorf("expected away, got %v", status)
	}
}

func TestGetDistanceBucket_Nearby(t *testing.T) {
	t.Parallel()

	bucket := GetDistanceBucket(0.5)
	if bucket != DistanceNearby {
		t.Errorf("expected nearby, got %v", bucket)
	}
}

func TestGetDistanceBucket_2km(t *testing.T) {
	t.Parallel()

	bucket := GetDistanceBucket(1.5)
	if bucket != Distance2km {
		t.Errorf("expected ~2km, got %v", bucket)
	}
}

func TestGetDistanceBucket_5km(t *testing.T) {
	t.Parallel()

	bucket := GetDistanceBucket(3.5)
	if bucket != Distance5km {
		t.Errorf("expected ~5km, got %v", bucket)
	}
}

func TestGetDistanceBucket_10km(t *testing.T) {
	t.Parallel()

	bucket := GetDistanceBucket(7.0)
	if bucket != Distance10km {
		t.Errorf("expected ~10km, got %v", bucket)
	}
}

func TestGetDistanceBucket_20kmPlus(t *testing.T) {
	t.Parallel()

	bucket := GetDistanceBucket(25.0)
	if bucket != Distance20kmPlus {
		t.Errorf("expected >20km, got %v", bucket)
	}
}

// ============================================================================
// UserProfile Tests
// ============================================================================

func TestIsEligibleForDiscovery_NotEnoughQuestions(t *testing.T) {
	t.Parallel()

	profile := &UserProfile{
		QuestionCount:       2,
		CategoriesCompleted: RequiredQuestionCategories,
	}

	if profile.IsEligibleForDiscovery() {
		t.Error("expected ineligible with too few questions")
	}
}

func TestIsEligibleForDiscovery_MissingCategory(t *testing.T) {
	t.Parallel()

	profile := &UserProfile{
		QuestionCount:       10,
		CategoriesCompleted: []string{"values", "social"}, // missing lifestyle, communication
	}

	if profile.IsEligibleForDiscovery() {
		t.Error("expected ineligible with missing categories")
	}
}

func TestIsEligibleForDiscovery_Eligible(t *testing.T) {
	t.Parallel()

	profile := &UserProfile{
		QuestionCount:       10,
		CategoriesCompleted: RequiredQuestionCategories,
	}

	if !profile.IsEligibleForDiscovery() {
		t.Error("expected eligible")
	}
}

// ============================================================================
// Adventure Helper Tests
// ============================================================================

func TestAdventure_IsFrozen(t *testing.T) {
	t.Parallel()

	adventure := &Adventure{Status: AdventureStatusFrozen}
	if !adventure.IsFrozen() {
		t.Error("expected frozen")
	}

	adventure.Status = AdventureStatusActive
	if adventure.IsFrozen() {
		t.Error("expected not frozen")
	}
}

func TestAdventure_IsGuildOrganized(t *testing.T) {
	t.Parallel()

	adventure := &Adventure{OrganizerType: AdventureOrganizerGuild}
	if !adventure.IsGuildOrganized() {
		t.Error("expected guild organized")
	}

	adventure.OrganizerType = AdventureOrganizerUser
	if adventure.IsGuildOrganized() {
		t.Error("expected not guild organized")
	}
}

func TestAdventure_IsUserOrganized(t *testing.T) {
	t.Parallel()

	adventure := &Adventure{OrganizerType: AdventureOrganizerUser}
	if !adventure.IsUserOrganized() {
		t.Error("expected user organized")
	}

	adventure.OrganizerType = AdventureOrganizerGuild
	if adventure.IsUserOrganized() {
		t.Error("expected not user organized")
	}
}
