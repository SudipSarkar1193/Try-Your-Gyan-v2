package middlewares

import (
	"net/http"
)


// Middleware to set COOP and COEP headers
func CoopMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set COOP header to allow popups for Firebase Auth
		w.Header().Set("Cross-Origin-Opener-Policy", "same-origin-allow-popups")
		// Set COEP to be less restrictive
		w.Header().Set("Cross-Origin-Embedder-Policy", "unsafe-none")

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

func HandleOptionsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "https://try-your-gyan.vercel.app") // Update to specific origin if needed
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
