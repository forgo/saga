package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/forgo/saga/api/pkg/jwt"
)

func main() {
	// Flags for customization
	privateKeyPath := flag.String("key", "./keys/private.pem", "Path to JWT private key")
	userID := flag.String("user", "admin-dev-user", "User ID for the token")
	email := flag.String("email", "admin@saga.dev", "Email for the token")
	issuer := flag.String("issuer", "saga", "JWT issuer")
	expMins := flag.Int("exp", 60*24*7, "Token expiration in minutes (default: 7 days)")
	outputJSON := flag.Bool("json", false, "Output as JSON")

	flag.Parse()

	// Create JWT service with just the private key
	jwtService, err := jwt.NewService(jwt.Config{
		PrivateKeyPath: *privateKeyPath,
		Issuer:         *issuer,
		ExpirationMins: *expMins,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating JWT service: %v\n", err)
		fmt.Fprintf(os.Stderr, "\nMake sure you have generated keys with: make keys-generate\n")
		os.Exit(1)
	}

	// Create admin claims
	claims := jwt.Claims{
		UserID:   *userID,
		Email:    *email,
		Username: "Admin",
		Role:     "admin",
	}

	// Sign token
	token, err := jwtService.Sign(claims)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error signing token: %v\n", err)
		os.Exit(1)
	}

	if *outputJSON {
		output := map[string]any{
			"access_token": token,
			"token_type":   "Bearer",
			"expires_in":   *expMins * 60,
			"user_id":      *userID,
			"email":        *email,
			"role":         "admin",
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(output)
	} else {
		expTime := time.Now().Add(time.Duration(*expMins) * time.Minute)
		fmt.Println("Admin Token Generated")
		fmt.Println("=====================")
		fmt.Printf("User ID:  %s\n", *userID)
		fmt.Printf("Email:    %s\n", *email)
		fmt.Printf("Role:     admin\n")
		fmt.Printf("Expires:  %s\n", expTime.Format(time.RFC3339))
		fmt.Println()
		fmt.Println("Token:")
		fmt.Println(token)
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Printf("  curl -H 'Authorization: Bearer %s' http://localhost:8080/v1/admin/seed/scenarios\n", token[:50]+"...")
	}
}
