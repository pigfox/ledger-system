package api

import (
	"ledger-system/internal/config"
	"ledger-system/recoverx"
	"net/http"
)

func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer recoverx.RecoverAndLog("HTTP Middleware")
		next.ServeHTTP(w, r)
	})
}

func APIKeyMiddleware(next http.Handler) http.Handler {
	requiredAPIKey := config.Cfg.APIKEY

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" || apiKey != requiredAPIKey {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
