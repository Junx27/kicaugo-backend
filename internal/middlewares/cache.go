package middlewares

import "github.com/danielgtaylor/huma/v2"

const publicCacheControl = "public, max-age=3600"

// CachePublic menetapkan header cache untuk data referensi yang boleh di-cache publik.
func CachePublic() func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		ctx.SetHeader("Cache-Control", publicCacheControl)
		next(ctx)
	}
}
