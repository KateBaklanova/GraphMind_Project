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
		return false, "Пароль должен быть длинне 8 символов"
	}
	if len(password) > 72 {
		return false, "Пароль не может быть длинее 72 символов"
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
		return false, "В пароле должна быть как минимум 1 заглавная буква"
	}
	if !hasLower {
		return false, "В пароле должна быть как минимум одна прописная буква"
	}
	if !hasNumber {
		return false, "В пароле должна быть как минимум одна цифра"
	}
	if !hasSpecial {
		return false, "В пароле должен быть как минимум однин спец. символ"
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
		return false, "Имя графа не может быть пустым"
	}
	if len(name) > 100 {
		return false, "Имя графа не может превышать 100 символов"
	}
	return true, ""
}
