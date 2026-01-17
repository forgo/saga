// Package helpers provides test utility functions for the Saga API.
//
// The helpers package contains common test utilities for assertions,
// pointer creation, and test data manipulation.
//
// # Pointer Helpers
//
// Create pointers to literal values:
//
//	name := helpers.StringPtr("test")
//	count := helpers.IntPtr(42)
//	flag := helpers.BoolPtr(true)
//
// # JWT Helpers
//
// Generate test JWT tokens:
//
//	token := helpers.GenerateTestToken(userID)
//	token := helpers.GenerateTestTokenWithClaims(claims)
//
// # Assertion Helpers
//
// Common test assertions:
//
//	helpers.AssertRecordExists(t, db, "guild:123")
//	helpers.AssertRecordNotExists(t, db, "guild:456")
//
// # Time Helpers
//
// Time manipulation for tests:
//
//	past := helpers.TimeAgo(24 * time.Hour)
//	future := helpers.TimeFromNow(1 * time.Hour)
package helpers
