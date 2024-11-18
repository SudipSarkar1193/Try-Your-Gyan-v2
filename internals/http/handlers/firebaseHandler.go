package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/database"
	"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/response"
	"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/types"
	"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/utils/tokens"
)

// HandleFirebaseAuth handles Firebase authentication for users
func HandleFirebaseAuth(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		fmt.Println()
		fmt.Println("HIT HandleFirebaseAuth")
		fmt.Println()

		ctx := context.Background()
		idToken := r.Header.Get("Authorization")
		
		if idToken == "" || !strings.HasPrefix(idToken, "Bearer ") {
			http.Error(w, "Missing or invalid Authorization token", http.StatusUnauthorized)
			return
		}
		idToken = strings.TrimPrefix(idToken, "Bearer ")
		

		// Validate Firebase token
		token, err := VerifyIDToken(ctx, idToken)
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid ID token: %v", err), http.StatusUnauthorized)
			return
		}

		log.Printf("Authenticated Firebase user: %s", token.UID)

		// Parse incoming JSON data
		var requestData struct {
			Username   string `json:"username"`
			Email      string `json:"email"`
			FirebaseId string `json:"firebaseId"`
			IsNewUser  bool   `json:"isNewUser"`
		}

		if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		var user *types.User
		if requestData.IsNewUser {
			// Register new user
			user = &types.User{
				Username: requestData.Username,
				Email:    requestData.Email,
				Password: requestData.FirebaseId,
			}

			id, err := database.InsertNewUser(db, user)
			if err != nil {
				http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
				return
			}
			user.Id = id

		} else {
			// Retrieve existing user
			user, err = database.RetrieveUser(db, requestData.Email)
			if err != nil {
				log.Printf("Error retrieving user: %v", err)
				http.Error(w, "Invalid credentials", http.StatusUnauthorized)
				return
			}

			if user == nil {
				log.Println("No user found")
				http.Error(w, "No user found", http.StatusUnauthorized)
				return
			}
		}

		// Generate tokens
		accessToken, refreshToken, err := tokens.GenerateTokens(user)
		if err != nil {
			http.Error(w, fmt.Sprintf("Could not generate tokens: %v", err), http.StatusInternalServerError)
			return
		}

		// Prepare response data
		responseData := struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
			Username     string `json:"username"`
		}{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			Username:     user.Username,
		}

		// Write response
		response.WriteResponse(w, response.CreateResponse(responseData, http.StatusOK, "Logged in successfully", "", "", false, ""))
	}
}
