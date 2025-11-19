package utils

import "context"

type CtxKey string

func GetString(ctx context.Context, key any) (string, bool) {
	v := ctx.Value(key)
	s, ok := v.(string)
	return s, ok
}

// GetEnv returns the environment from context (defaults to "dev")
func GetEnv(ctx context.Context) string {
	env, _ := GetString(ctx, "env")
	if env == "" {
		return "dev"
	}
	return env
}
