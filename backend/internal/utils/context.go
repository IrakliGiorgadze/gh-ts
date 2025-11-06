package utils

import "context"

type CtxKey string

func GetString(ctx context.Context, key any) (string, bool) {
	v := ctx.Value(key)
	s, ok := v.(string)
	return s, ok
}
