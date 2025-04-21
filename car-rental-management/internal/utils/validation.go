package utils

import "regexp"

// Consider adding more complex validation logic if needed (e.g., stronger password checks)

var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)

// IsValidEmail checks if the email format is valid.
func IsValidEmail(email string) bool {
	// Basic check for presence and standard format
	if email == "" || len(email) > 254 { // RFC 5321 limit might be relevant
		return false
	}
	return emailRegex.MatchString(email)
}
