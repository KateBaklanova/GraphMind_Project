package utils

import (
	"regexp"
	"strings"
	"unicode"
)

func ValidateEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	match, _ := regexp.MatchString(pattern, email)
	return match
}

func ValidatePassword(password string) (bool, string) {
	if len(password) < 8 {
		return false, "password must be at least 8 characters"
	}
	if len(password) > 72 {
		return false, "password must be less than 72 characters"
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return false, "password must contain at least one uppercase letter"
	}
	if !hasLower {
		return false, "password must contain at least one lowercase letter"
	}
	if !hasNumber {
		return false, "password must contain at least one number"
	}
	if !hasSpecial {
		return false, "password must contain at least one special character"
	}
	return true, ""
}

func SanitizeInput(input string) string {
	dangerous := []string{"<", ">", "\"", "'", "&", ";", "`", "|", "*", "?", "~", "^", "(", ")", "[", "]", "{", "}", "$", "!", "\\"}
	result := input
	for _, char := range dangerous {
		result = strings.ReplaceAll(result, char, "")
	}
	return strings.TrimSpace(result)
}

func ValidateGraphName(name string) (bool, string) {
	if len(name) < 1 {
		return false, "graph name cannot be empty"
	}
	if len(name) > 100 {
		return false, "graph name must be less than 100 characters"
	}
	return true, ""
}
