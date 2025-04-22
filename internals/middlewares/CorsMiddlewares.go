package middlewares

import (
	"log"
	"net/http"
)

// Middleware to set COOP and COEP headers
func CoopMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//DEBUG :

		log.Println()
		log.Println("DEBUG : Inside CoopMiddleware")
		log.Println()

		// Set COOP header to allow popups for Firebase Auth
		w.Header().Set("Cross-Origin-Opener-Policy", "same-origin-allow-popups")
		// Set COEP to be less restrictive
		w.Header().Set("Cross-Origin-Embedder-Policy", "unsafe-none")

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

func DebugOriginMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		log.Println()
		log.Println("Inside DebugOriginMiddleware")
		log.Println()

		origin := r.Header.Get("Origin")
		log.Println()
		if origin == "" {
			log.Println("Debug: Missing Origin header")
		} else {
			log.Printf("Debug: Origin header present: %s\n", origin)
		}
		log.Println()
		next.ServeHTTP(w, r)
	})
}

func DebugMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("DEBUG: DebugMiddleware hit")
		log.Println()
		log.Printf("Request Headers: %+v\n", r.Header)
		log.Println()
		log.Printf("Request Method: %s\n", r.Method)
		log.Println()

		log.Printf("Response Headers: %+v\n", w.Header())
		log.Println()

		log.Printf("GETTING OUT NOW FROM DebugMiddleware\n, Path: %s, Method: %s\n", r.URL.Path, r.Method)
		next.ServeHTTP(w, r)
	})
}

func HandleOptionsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method == http.MethodOptions {
			// Add CORS headers for preflight requests
			w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Access-Control-Allow-Origin")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			// log.Println("Handled 'r.Method == http.MethodOptions' block ")
			w.WriteHeader(http.StatusOK)

		}
		// log.Println("Reached out of if 'r.Method == http.MethodOptions' block")

		next.ServeHTTP(w, r)
	})

}
