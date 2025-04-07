package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/database"
	"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/response"
	"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/types"
	"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/utils/tokens"
)

// HandleFirebaseAuth handles Firebase authentication for users
func HandleFirebaseAuth(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Println("[FirebaseAuth] Request received")

		ctx := context.Background()
		idToken := r.Header.Get("Authorization")

		if idToken == "" || !strings.HasPrefix(idToken, "Bearer ") {
			log.Println("[FirebaseAuth] Missing or invalid Authorization token")
			http.Error(w, "Missing or invalid Authorization token", http.StatusUnauthorized)
			return
		}

		// Validate Firebase token
		idToken = strings.TrimPrefix(idToken, "Bearer ")

		token, err := VerifyIDToken(ctx, idToken)
		if err != nil {
			log.Printf("[FirebaseAuth] Invalid ID token: %v", err)
			http.Error(w, fmt.Sprintf("invalid ID token: %v", err), http.StatusUnauthorized)
			return
		}

		log.Printf("[FirebaseAuth] Authenticated Firebase user: %s", token.UID)

		// Parse incoming JSON data
		var requestData struct {
			Username   string `json:"username"`
			Email      string `json:"email"`
			FirebaseId string `json:"firebaseId"`
			IsNewUser  bool   `json:"isNewUser"`
			ProfileImg string `json:"profileImg"`
		}

		if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
			log.Printf("[FirebaseAuth] Invalid JSON: %v", err)
			http.Error(w, "Invalid JSON ", http.StatusBadRequest)
			return
		}

		var user *types.User
		var isCondHit1 bool
		var isCondHit2 bool

		if requestData.IsNewUser {
			log.Println("[FirebaseAuth] New user registration")
			cleanedUsername := strings.ReplaceAll(requestData.Username, " ", "")
			var uniqueUsername string
			for {
				randomNumber := rand.Intn(900) + 100
				generatedUsername := fmt.Sprintf("%s%d", cleanedUsername, randomNumber)
				exists, err := database.UsernameExists(db, generatedUsername)
				if err != nil {
					log.Printf("[FirebaseAuth] DB error on checking username existence: %v", err)
					http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
					return
				}
				if !exists {
					uniqueUsername = generatedUsername
					break
				}
			}

			// Register new user
			user = &types.User{
				Username:   uniqueUsername,
				Email:      requestData.Email,
				Password:   "",
				IsVarified: true,
				ProfileImg: requestData.ProfileImg,
			}

			id, err := database.InsertNewUser(db, user)
			if err != nil {
				if strings.Contains(err.Error(), `duplicate key value violates unique constraint "users_email_key"`) {

					log.Println("[FirebaseAuth] Duplicate email, fetching existing user")

					// Retrieve that existing user
					user, err = database.RetrieveUser(db, requestData.Email)
					if err != nil {
						log.Printf("[FirebaseAuth] Error retrieving user: %v", err)
						http.Error(w, fmt.Sprintf("Error retrieving user: %v", err), http.StatusInternalServerError)
						return
					}

					if user.ProfileImg == "" || user.ProfileImg == "https://res.cloudinary.com/dvsutdpx2/image/upload/v1732181213/ryi6ouf4e0mwcgz1tcxx.png" {
						database.UserFindByEmailAndUpdateProfileImg(db, user.Email, requestData.ProfileImg)
					}

					if !user.IsVarified {
						isCondHit1 = true
						database.UpdateUserById(db, user.Id, true)
					}
					isCondHit2 = true
					goto POINT01
				}
				log.Printf("[FirebaseAuth] Error inserting new user: %v", err)
				http.Error(w, fmt.Sprintf("Error inserting new user: %v", err), http.StatusInternalServerError)
				return
			}
			user.Id = id
		} else {
			log.Println("[FirebaseAuth] Returning user login")
			user, err = database.RetrieveUser(db, requestData.Email)
			if err != nil {
				log.Printf("[FirebaseAuth] Error retrieving user: %v", err)
				http.Error(w, fmt.Sprintf("Error retrieving user: %v", err), http.StatusInternalServerError)
				return
			}
			if user == nil {
				log.Println("No user found")
				http.Error(w, "No user found", http.StatusInternalServerError)
				return
			}

			if user.ProfileImg == "" || user.ProfileImg == "https://res.cloudinary.com/dvsutdpx2/image/upload/v1732181213/ryi6ouf4e0mwcgz1tcxx.png" {
				database.UserFindByEmailAndUpdateProfileImg(db, requestData.Email, requestData.ProfileImg)
			}
			if !user.IsVarified {
				database.UpdateUserById(db, user.Id, true)
			}

		}

	POINT01:
		accessToken, refreshToken, err := tokens.GenerateTokens(user)
		if err != nil {
			log.Printf("[FirebaseAuth] Failed to generate tokens: %v", err)
			http.Error(w, fmt.Sprintf("Could not generate tokens: %v", err), http.StatusInternalServerError)
			return
		}

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
		msg := "Logged in successfully with Google"
		if isCondHit2 {
			if isCondHit1 {
				msg = "This email is now varified"
			} else {
				msg = "This email is already varified using OTP"
			}
		}

		log.Printf("[FirebaseAuth] Completed in %v", time.Since(start))
		response.WriteResponse(w, response.CreateResponse(responseData, http.StatusOK, msg, "", "", false, ""))
	}
}
