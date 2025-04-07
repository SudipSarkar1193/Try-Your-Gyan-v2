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

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		start := time.Now()
		log.Println("[FirebaseAuth] Request received")

		ctx := r.Context()
		idToken := r.Header.Get("Authorization")

		if idToken == "" || !strings.HasPrefix(idToken, "Bearer ") {
			log.Println("[FirebaseAuth] Missing or invalid Authorization token")
			http.Error(w, "Missing or invalid Authorization token", http.StatusUnauthorized)
			return
		}

		// Validate Firebase token
		tokenStart := time.Now()
		log.Println("[FirebaseAuth] Verifying ID token")

		idToken = strings.TrimPrefix(idToken, "Bearer ")

		token, err := VerifyIDToken(ctx, idToken)
		if err != nil {
			log.Printf("[FirebaseAuth] Invalid ID token: %v", err)
			http.Error(w, fmt.Sprintf("invalid ID token: %v", err), http.StatusUnauthorized)
			return
		}

		log.Printf("[FirebaseAuth] Token verified in %v", time.Since(tokenStart))

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
		var shouldUpdateProfileImg bool
		var shouldVerifyUser bool

		if requestData.IsNewUser {
			IsNewUserStart := time.Now()

			log.Println("[FirebaseAuth] New user registration starts")

			log.Println("[FirebaseAuth] Searching for new user unique username -> started")

			usernameStart := time.Now()
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

			log.Println("[FirebaseAuth] Searching for new user unique username -> Ended . Time taken :", time.Since(usernameStart))

			// Register new user
			user = &types.User{
				Username:   uniqueUsername,
				Email:      requestData.Email,
				Password:   "",
				IsVarified: true,
				ProfileImg: requestData.ProfileImg,
			}

			log.Println("[FirebaseAuth] Inserting new user in DB -> started")
			insertTime := time.Now()
			id, err := database.InsertNewUser(db, user)
			log.Println("[FirebaseAuth] Inserting new user in DB -> Ended . Time taken :", time.Since(insertTime))
			if err != nil {
				if strings.Contains(err.Error(), `duplicate key value violates unique constraint "users_email_key"`) {

					duplicateEmailTime := time.Now()

					log.Println("[FirebaseAuth] Duplicate email, fetching existing user -> started")

					// Retrieve that existing user
					user, err = database.RetrieveUser(db, requestData.Email)
					log.Println("[FirebaseAuth] Duplicate email, fetching existing user -> ended. Time.taken : ", time.Since(duplicateEmailTime))

					if err != nil {
						log.Printf("[FirebaseAuth] Error retrieving user: %v", err)
						http.Error(w, fmt.Sprintf("Error retrieving user: %v", err), http.StatusInternalServerError)
						return
					}

					attributeTime := time.Now()
					log.Println("[FirebaseAuth] Duplicate email, Existing user attribute flagging -> started")

					shouldUpdateProfileImg = user.ProfileImg == "" ||
						user.ProfileImg == "https://res.cloudinary.com/dvsutdpx2/image/upload/v1732181213/ryi6ouf4e0mwcgz1tcxx.png"

					shouldVerifyUser = !user.IsVarified

					if shouldUpdateProfileImg {
						isCondHit2 = true
					}
					if shouldVerifyUser {
						if err := database.UpdateUserById(ctx, db, user.Id, true); err != nil {
							log.Printf("[FirebaseAuth] Failed to verify user: %v", err)
							http.Error(w, "Verification failed", http.StatusInternalServerError)
							return
						}
						isCondHit1 = true
						isCondHit2 = true
					}

					log.Println("[FirebaseAuth] Duplicate email, Existing user attribute flagging -> ended. Time taken:", time.Since(attributeTime))

					log.Println("[FirebaseAuth] Duplicate email, Going to [ goto POINT01 / respondWithTokens ]")
					// goto POINT01
					respondWithTokens(w, user, isCondHit1, isCondHit2, start)

				}

				log.Printf("[FirebaseAuth] Error inserting new user: %v", err)
				http.Error(w, fmt.Sprintf("Error inserting new user: %v", err), http.StatusInternalServerError)
				return
			}
			user.Id = id

			log.Println("[FirebaseAuth] New user registration done . time taken :", time.Since(IsNewUserStart))

		} else {
			isNotNewUserStart := time.Now()

			log.Println("[FirebaseAuth] Existing user login -> started")

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

			shouldUpdateProfileImg = user.ProfileImg == "" ||
				user.ProfileImg == "https://res.cloudinary.com/dvsutdpx2/image/upload/v1732181213/ryi6ouf4e0mwcgz1tcxx.png"

			shouldVerifyUser = !user.IsVarified

			if shouldVerifyUser {
				if err := database.UpdateUserById(ctx, db, user.Id, true); err != nil {
					log.Printf("[FirebaseAuth] Failed to verify user: %v", err)
					http.Error(w, "Verification failed", http.StatusInternalServerError)
					return
				}
				isCondHit1 = true
				isCondHit2 = true
			}

			log.Println("[FirebaseAuth] Existing user login -> ended. Time taken : ", time.Since(isNotNewUserStart))

		}

		//POINT01:

		respondWithTokens(w, user, isCondHit1, isCondHit2, start)

		go func() {
			
			ctx := context.Background()
			goRoutineTime := time.Now()
			if shouldUpdateProfileImg {
				log.Println("Go routine for updating user ->started ")
			
				if err := database.UserFindByEmailAndUpdateProfileImg(ctx, db, user.Email, requestData.ProfileImg); err != nil {
					log.Printf("[Post-Response Task] Failed to update profile image: %v", err)
				}
			}

			log.Println("Go routine for updating user ->ended . Time taken :", time.Since(goRoutineTime))
		}()

	}
}

func respondWithTokens(w http.ResponseWriter, user *types.User, isCondHit1, isCondHit2 bool, start time.Time) {
	log.Println("[FirebaseAuth] TOKEN generation -> started")
	tokenGenTime := time.Now()
	accessToken, refreshToken, err := tokens.GenerateTokens(user)
	if err != nil {
		log.Printf("[FirebaseAuth] Failed to generate tokens: %v", err)
		http.Error(w, fmt.Sprintf("Could not generate tokens: %v", err), http.StatusInternalServerError)
		return
	}

	log.Println("[FirebaseAuth] TOKEN generation -> ended. Time taken :", time.Since(tokenGenTime))

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
