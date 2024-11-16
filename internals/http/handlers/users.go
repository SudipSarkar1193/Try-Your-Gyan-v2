package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/database"
	"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/password"
	"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/response"
	"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/types"
	"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/utils/tokens"
)

func New(db *sql.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			http.Error(w, fmt.Sprintf("%v HTTP method is not allowed", r.Method), http.StatusBadRequest)
			return
		}

		var user types.User

		err := json.NewDecoder(r.Body).Decode(&user)
		// ⭐⭐ Explaination :

		/*
			1. r.Body contains the body of the HTTP request, typically in JSON format, that is sent to the server.

			2. json.NewDecoder(r.Body) creates a new JSON decoder to read and parse the JSON data from the r.Body.

			3. .Decode(&student) attempts to decode (unmarshal) the JSON data into the student struct. The &student is a pointer to the student struct, which means the decoded data will be stored directly into this struct.

			So, essentially, this line reads the JSON payload from the request body and decodes it into the Go struct named student.

		*/
		if err != nil {
			if errors.Is(err, io.EOF) {
				//io.EOF is a sentinel error in Go that indicates the end of input (end of a file or stream), commonly returned by functions when there is no more data to read.

				http.Error(w, fmt.Sprintf("no data to read: %v", err.Error()), http.StatusBadRequest)
				return

			} else {
				// Handle other decoding errors

				http.Error(w, fmt.Sprintf("failed to decode JSON: %v", err.Error()), http.StatusInternalServerError)
				return
			}
		}

		// Validate that all fields are filled
		// if student.Id == 0 || student.Name == "" || student.Email == "" {
		// 	http.Error(w, "all fields are required !", http.StatusBadRequest)
		// 	return

		// }

		var validate *validator.Validate

		validate = validator.New(validator.WithRequiredStructEnabled())

		if err := validate.Struct(&user); err != nil {
			if _, ok := err.(*validator.InvalidValidationError); ok {
				fmt.Println(err)
				return
			}

			response.ValidateResponse(w, err)
			return
		}

		//Everything is fine till now

		hashpass, err := password.HashPassword(user.Password)

		if err != nil {
			http.Error(w, fmt.Sprintf("failed to encrypt the password : %v", err.Error()), http.StatusInternalServerError)
			return
		}

		user.Password = hashpass

		if err := database.InsertNewUser(db, &user); err != nil {
			http.Error(w, fmt.Sprintf("Database error : %v", err), http.StatusInternalServerError)
			return
		}

		emptyResponse := response.CreateResponse(user, http.StatusCreated, "User created Successfully", "<DeveloperMessage>", "<UserMessage>", false, "Err")

		response.WriteResponse(w, emptyResponse)

	}
}

func Login(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, fmt.Sprintf("%v HTTP method is not allowed", r.Method), http.StatusBadRequest)
			return
		}

		var loginData struct {
			Identifier string `json:"identifier" validate:"required"`
			Password   string `json:"password" validate:"required"`
		}
		
		err := json.NewDecoder(r.Body).Decode(&loginData)
		if err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		validate := validator.New()
		if err := validate.Struct(loginData); err != nil {
			response.ValidateResponse(w, err)
			return
		}

		// Retrieve user by email or username
		user, err := database.RetrieveUser(db, loginData.Identifier)
		if err != nil {
			http.Error(w, "user not found : Invalid credentials", http.StatusUnauthorized)
			return
		}

		// Check password
		isPasswordValid, err := password.CheckPassword(loginData.Password, user.Password)
		if err != nil || !isPasswordValid {
			http.Error(w, "user not found : Invalid credentials", http.StatusUnauthorized)
			return
		}

		// Generate tokens
		accessToken, refreshToken, err := tokens.GenerateTokens(user)
		if err != nil {
			http.Error(w, fmt.Sprintf("Could not generate tokens : %v", err.Error()), http.StatusInternalServerError)
			return
		}

		// Send tokens as JSON response
		tokenResponse := map[string]string{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"username":      user.Username,
		}
		response.WriteResponse(w, response.CreateResponse(tokenResponse, http.StatusOK, "Logged in successfully", "", "", false, ""))
	}
}
