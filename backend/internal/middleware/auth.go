package middleware

import (
	"context"
	"net/http"
	"strings"

	"gh-ts/internal/config"
	"gh-ts/internal/utils"

	"github.com/rs/zerolog"
)

type ctxKey string

const (
	CtxUserID ctxKey = "uid"
	CtxRole   ctxKey = "role"
)

func WithAuth(log zerolog.Logger, cfg config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Read JWT from cookie "session" or Authorization: Bearer
			var tok string
			if c, err := r.Cookie("session"); err == nil {
				tok = c.Value
			} else if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
				tok = strings.TrimPrefix(h, "Bearer ")
			}

			if tok == "" {
				next.ServeHTTP(w, r) // unauthenticated; handlers can decide
				return
			}

			claims, err := utils.ParseJWT(cfg.SessionSecret, tok)
			if err != nil {
				// IMPORTANT: clear broken/expired cookie so it stops being sent
				http.SetCookie(w, &http.Cookie{
					Name:     "session",
					Value:    "",
					Path:     "/",
					HttpOnly: true,
					MaxAge:   -1,
				})
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), CtxUserID, claims.UserID)
			ctx = context.WithValue(ctx, CtxRole, claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
