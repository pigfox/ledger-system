package api

import (
	"context"
	"github.com/gorilla/mux"
	"ledger-system/internal/config"
	"ledger-system/recoverx"
	"net/http"
	"time"
)

func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer recoverx.RecoverAndLog("HTTP Middleware")
		next.ServeHTTP(w, r)
	})
}

func CheckAPIKeyMiddleware(next http.Handler) http.Handler {
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

func ContextTimeoutMiddleware(timeout time.Duration) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			// Replace the request context with the timeout-enabled one
			r = r.WithContext(ctx)

			// Run the next handler
			next.ServeHTTP(w, r)
		})
	}
}
