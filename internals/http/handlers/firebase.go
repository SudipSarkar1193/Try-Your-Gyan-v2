package handlers

import (
	"context"
	"encoding/base64"

	"fmt"
	"log"

	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

// Global Firebase Auth Client
var FirebaseAuthClient *auth.Client

// Initialize Firebase Admin SDK and Auth Client
func InitializeFirebaseApp() {
	encodedCreds := os.Getenv("FIREBASE_SERVICE_ACCOUNT")
	decodedCreds, err := base64.StdEncoding.DecodeString(encodedCreds)
	if err != nil {
		log.Fatalf("Failed to decode Firebase credentials: %v", err)
	}
	sa := option.WithCredentialsJSON(decodedCreds)

	app, err := firebase.NewApp(context.Background(), nil, sa)
	if err != nil {
		log.Fatalf("Failed to initialize Firebase App: %v", err)
	}

	// Initialize the Firebase Auth client
	FirebaseAuthClient, err = app.Auth(context.Background())
	if err != nil {
		log.Fatalf("Failed to initialize Firebase Auth client: %v", err)
	}

	log.Println("Firebase Auth initialized successfully",FirebaseAuthClient)
}

// Verify Firebase ID Token
func VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error) {
	
	if FirebaseAuthClient == nil {
		return nil, fmt.Errorf("firebase Auth client not initialized")
	}

	token, err := FirebaseAuthClient.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, err
	}

	return token, nil
}
