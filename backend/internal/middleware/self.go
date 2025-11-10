package middleware

import (
	"net/http"

	"gh-ts/internal/utils"

	"github.com/go-chi/chi/v5"
)

// RequireSelfOrRoles allows if {id} == ctx user id OR user has any of the given roles.
func RequireSelfOrRoles(roles ...string) func(http.Handler) http.Handler {
	roleSet := map[string]struct{}{}
	for _, r := range roles {
		roleSet[r] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctxUID, _ := utils.GetString(r.Context(), CtxUserID)
			ctxRole, _ := utils.GetString(r.Context(), CtxRole)
			pathID := chi.URLParam(r, "id")

			// allow admins (or other roles you include)
			if _, ok := roleSet[ctxRole]; ok {
				next.ServeHTTP(w, r)
				return
			}
			// otherwise only self
			if ctxUID != "" && pathID == ctxUID {
				next.ServeHTTP(w, r)
				return
			}
			http.Error(w, "forbidden", http.StatusForbidden)
		})
	}
}
