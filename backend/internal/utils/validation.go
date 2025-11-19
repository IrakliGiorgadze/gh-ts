package utils

import (
	"errors"
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

const (
	MaxEmailLength       = 254
	MaxNameLength        = 100
	MaxTitleLength       = 200
	MaxDescriptionLength = 5000
	MaxCommentLength     = 2000
	MaxSearchLength      = 100
	MaxDepartmentLength  = 100
)

// ValidateEmail checks if email format is valid
func ValidateEmail(email string) bool {
	email = strings.TrimSpace(email)
	if len(email) < 3 || len(email) > MaxEmailLength {
		return false
	}
	return emailRegex.MatchString(email)
}

// ValidateLength checks if string is within max length
func ValidateLength(s string, max int) bool {
	return len(s) <= max
}

// TruncateString safely truncates string to max length
func TruncateString(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}

// ValidatePasswordStrength checks password meets requirements
// - At least 8 characters
// - Contains at least one uppercase letter
// - Contains at least one lowercase letter
// - Contains at least one number
func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	var (
		hasUpper bool
		hasLower bool
		hasNumber bool
	)

	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case 'a' <= char && char <= 'z':
			hasLower = true
		case '0' <= char && char <= '9':
			hasNumber = true
		}
	}

	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}
	if !hasNumber {
		return errors.New("password must contain at least one number")
	}

	return nil
}

