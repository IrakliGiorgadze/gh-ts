package middleware

import (
	"net/http"

	"gh-ts/internal/utils"
)

// RequireAuth blocks when no user is present in context (set by WithAuth).
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid, ok := utils.GetString(r.Context(), CtxUserID)
		if !ok || uid == "" {
			utils.Error(w, http.StatusUnauthorized, "authentication required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireRoles allows request only if the current role is in the allowed list.
func RequireRoles(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, _ := utils.GetString(r.Context(), CtxRole)
			if _, ok := allowed[role]; !ok {
				utils.Error(w, http.StatusForbidden, "forbidden")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
