package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"


	"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/database"
	"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/response"
	"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/utils/tokens"
	
)

var firebaseAuth *auth.Client

// Initialize Firebase Admin SDK
func initializeFirebaseApp() *firebase.App {
	// Get the service account JSON string from the environment variable
	serviceAccountJSON := os.Getenv("FIREBASE_SERVICE_ACCOUNT")
	if serviceAccountJSON == "" {
		log.Fatal("FIREBASE_SERVICE_ACCOUNT environment variable not set")
	}

	// Parse the JSON string to initialize Firebase
	sa := option.WithCredentialsJSON([]byte(serviceAccountJSON))
	app, err := firebase.NewApp(context.Background(), nil, sa)
	if err != nil {
		log.Fatalf("Failed to initialize Firebase App: %v", err)
	}

	log.Println("Firebase App initialized successfully")
	return app
}

// Verify Firebase ID Token
func verifyIDToken(ctx context.Context, idToken string) (*auth.Token, error) {
	token, err := firebaseAuth.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("invalid ID token: %v", err)
	}
	return token, nil
}

// Handle incoming requests from the frontend

func handleFirebaseAuth(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		idToken := r.Header.Get("Authorization")

		if idToken == "" {
			http.Error(w, "Missing Authorization token", http.StatusUnauthorized)
			return
		}

		// Validate Firebase token
		token, err := verifyIDToken(ctx, idToken)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Parse the incoming JSON data
		var requestData struct{
			Username string `json:username`
			Email string `json:email`
			FirebaseId string `json:firebaseId`
		}

		if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		fmt.Printf("Received data from user %s: %v\n", token.UID, requestData)

		user, err := database.RetrieveUser(db, requestData.Email)
		if err != nil {
			http.Error(w, "user not found : Invalid credentials", http.StatusUnauthorized)
			return
		}

		if(user==nil){
			
		}

		accessToken, refreshToken, err := tokens.GenerateTokens(user)
		if err != nil {
			http.Error(w, fmt.Sprintf("Could not generate tokens : %v", err.Error()), http.StatusInternalServerError)
			return
		}

		type responseDatastruct struct{
			AccessToken string `json:access_token`
			RefreshToken string `json:refresh_token`
			Username string `json:username`
		}

		responseData := responseDatastruct{
			AccessToken:accessToken,
			RefreshToken:refreshToken,
			Username:requestData.Username,

		}

		response.WriteResponse(w, response.CreateResponse(responseData, http.StatusOK, "Logged in successfully", "", "", false, ""))


	}
}
