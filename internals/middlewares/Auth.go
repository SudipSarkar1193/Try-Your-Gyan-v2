package middlewares

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	// "github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/config"

	"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/database"
	"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/response"
	"github.com/golang-jwt/jwt/v5"
)

// Replace with your actual secret key

// AuthMiddleware verifies the access token from the request header
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		log.Printf("DEBUG: AuthMiddleware hit, Path: %s, Method: %s\n", r.URL.Path, r.Method)
		log.Println()
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

func GetUserDetails(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDStr := r.Header.Get("userID")
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		user, err := database.RetrieveUser(db, userID)

		if user.Password != nil {
			user.Password = true
		} else {
			user.Password = false
		}

		if err != nil {
			http.Error(w, fmt.Sprintf("Couldn't retrieve the user : %v", err), http.StatusInternalServerError)
		}

		resp := response.CreateResponse(user, 200, "User-details retrieved succesfully")

		response.WriteResponse(w, resp)
	}
}

func VerifyUserMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

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
			// Attach user ID to request context
			r.Header.Set("userID", fmt.Sprintf("%v", claims["sub"]))
		} else {
			http.Error(w, "Could not parse token claims", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}
}
