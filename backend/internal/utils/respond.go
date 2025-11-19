package utils

import (
	"encoding/json"
	"net/http"
	"strings"
)

func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func Error(w http.ResponseWriter, status int, msg string) {
	JSON(w, status, map[string]string{"error": msg})
}

// SafeError returns a safe error message based on environment
// In production, returns generic messages to avoid information disclosure
// In development, returns the actual error message for debugging
func SafeError(env string, err error) string {
	if err == nil {
		if env == "prod" {
			return "an error occurred"
		}
		return "unknown error"
	}
	
	errStr := err.Error()
	
	// In production, map common errors to safe messages
	if env == "prod" {
		lowerErr := strings.ToLower(errStr)
		
		// Database errors
		if strings.Contains(lowerErr, "duplicate") || strings.Contains(lowerErr, "unique") {
			return "resource already exists"
		}
		if strings.Contains(lowerErr, "foreign key") {
			return "invalid reference"
		}
		if strings.Contains(lowerErr, "not found") || strings.Contains(lowerErr, "no rows") {
			return "resource not found"
		}
		if strings.Contains(lowerErr, "connection") || strings.Contains(lowerErr, "timeout") {
			return "service temporarily unavailable"
		}
		if strings.Contains(lowerErr, "permission") || strings.Contains(lowerErr, "access") {
			return "access denied"
		}
		
		// Generic fallback for production
		return "an error occurred"
	}
	
	// In development, return actual error
	return errStr
}
