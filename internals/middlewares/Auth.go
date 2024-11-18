package middlewares

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	// "github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/config"
	"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/response"
	"github.com/golang-jwt/jwt/v5"
)

// Replace with your actual secret key

// AuthMiddleware verifies the access token from the request header
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// if err := config.LoadEnvFile(".env"); err != nil {
		// 	fmt.Println("LoadEnvFile error :", err ," statusCode:",http.StatusInternalServerError)

		// }
		var jwtSecret = []byte(os.Getenv("JWT_SECRET_KEY"))
		authHeader := r.Header.Get("Authorization")
		
		

		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Authorization header is required !", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		
		// Parse and validate the token
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid { // !token.Valid: If the token itself is invalid (e.g., incorrect signature or expired)
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Extract user information from token claims
		claims, ok := token.Claims.(jwt.MapClaims)

		if ok && token.Valid {
			// Attach user ID or other details to request context
			r.Header.Set("userID", fmt.Sprintf("%v", claims["sub"]))
			r.Header.Set("username", fmt.Sprintf("%v", claims["name"]))
		} else {
			http.Error(w, "Could not parse token claims", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func GetUserDetails() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := r.Header.Get("userId")
		username := r.Header.Get("username")

		userDetails := struct {
			Id       string `json:"id"`
			Username string `json:"username"`
		}{
			Id:       userId,
			Username: username,
		}

		resp := response.CreateResponse(userDetails, 200, "User-details retrieved succesfully")

		response.WriteResponse(w, resp)
	}
}

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
