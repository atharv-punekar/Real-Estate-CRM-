package utils

import "regexp"

// Strict lowercase-only email regex
var strictEmailRegex = regexp.MustCompile(`^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$`)

func IsValidEmail(email string) bool {
	return strictEmailRegex.MatchString(email)
}
