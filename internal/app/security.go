package app

import (
	"fmt"
	"net/http"

	"github.com/puppe1990/cais/pkg/cais"
)

// securityHeadersAuctionHQ extends Cais security headers to allow design-system fonts
// (Inter, JetBrains Mono, Material Symbols) from Google Fonts.
func securityHeadersAuctionHQ(cfg cais.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			policy := cfg.PermissionsPolicy
			if policy == "" {
				policy = "camera=(), microphone=(), geolocation=()"
			}
			w.Header().Set("Permissions-Policy", policy)

			styleSrc := "'self' 'unsafe-inline' https://fonts.googleapis.com"
			if cfg.CSPStyleSrc != "" {
				styleSrc += " " + cfg.CSPStyleSrc
			}
			fontSrc := "'self' https://fonts.gstatic.com data:"
			connectSrc := "'self'"
			if cfg.CSPConnectSrc != "" {
				connectSrc += " " + cfg.CSPConnectSrc
			}
			mediaSrc := "'self'"
			if cfg.CSPMediaSrc != "" {
				mediaSrc += " " + cfg.CSPMediaSrc
			}
			imgSrc := "'self' data: https:"
			if cfg.CSPImgSrc != "" {
				imgSrc += " " + cfg.CSPImgSrc
			}

			w.Header().Set("Content-Security-Policy", fmt.Sprintf(
				"default-src 'self'; script-src 'self' 'unsafe-inline'; style-src %s; font-src %s; img-src %s; connect-src %s; media-src %s; frame-ancestors 'none'; base-uri 'self'; form-action 'self'",
				styleSrc, fontSrc, imgSrc, connectSrc, mediaSrc,
			))
			if cfg.Env == "production" {
				w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}
			next.ServeHTTP(w, r)
		})
	}
}
