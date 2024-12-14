package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/database"
	"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/password"
	"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/response"
	"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/types"
	"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/utils/cloudinary"
	"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/utils/email"
	"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/utils/tokens"
)

func GenerateRandomString() string {
	rand.Seed(time.Now().UnixNano())                 // Seed to ensure randomness
	return fmt.Sprintf("%04d", 1000+rand.Intn(9000)) // Generate number [1000, 9999] and format to 4 digits
}

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

		existingUserWithEmail, _ := database.RetrieveUser(db, user.Email)

		existingUserWithUsername, _ := database.RetrieveUser(db, user.Username)

		if existingUserWithEmail != nil {
			http.Error(w, fmt.Sprintf("User with the email : %v already exists", existingUserWithEmail.Email), http.StatusBadRequest)
			return
		}
		if existingUserWithUsername != nil {
			http.Error(w, fmt.Sprintf("User with the username : %v already exists", existingUserWithUsername.Username), http.StatusBadRequest)
			return
		}

		var validate *validator.Validate

		validate = validator.New(validator.WithRequiredStructEnabled())

		if err := validate.Struct(&user); err != nil {
			if _, ok := err.(*validator.InvalidValidationError); ok {
				http.Error(w, "validate Struct error", http.StatusBadRequest)
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

		user_id, err := database.InsertNewUser(db, &user)
		if err != nil {
			http.Error(w, fmt.Sprintf("Database error : %v", err), http.StatusInternalServerError)
			return
		}
		token, err := tokens.GenerateVerifyToken(&user)

		if err != nil {
			http.Error(w, fmt.Sprintf("Could not generate tokens : %v", err.Error()), http.StatusInternalServerError)
			return
		}

		// Send tokens as JSON response
		tokenResponse := map[string]string{
			"verify_token": token,
		}

		//insert otp to database :
		otp := GenerateRandomString()

		database.InsertNewOTP(db, otp, user_id)

		if err := email.SendOTPEmail(user.Email, otp); err != nil {
			http.Error(w, fmt.Sprintf("Failed to send OTP %v", err), http.StatusInternalServerError)
			return
		}

		emptyResponse := response.CreateResponse(tokenResponse, http.StatusCreated, "User created Successfully", "<DeveloperMessage>", "<UserMessage>", false, "Err")

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
			http.Error(w, "user not found", http.StatusUnauthorized)
			return
		}

		// Check password
		isPasswordValid, err := password.CheckPassword(loginData.Password, user.Password)
		if err != nil || !isPasswordValid {
			http.Error(w, "Wrong password", http.StatusUnauthorized)
			return
		}

		// Generate tokens
		accessToken, refreshToken, err := tokens.GenerateTokens(user)
		if err != nil {
			http.Error(w, fmt.Sprintf("Could not generate tokens : %v", err.Error()), http.StatusInternalServerError)
			return
		}

		//Check if varified :
		if !user.IsVarified {
			verifyToken, err := tokens.GenerateVerifyToken(user)
			if err != nil {
				http.Error(w, fmt.Sprintf("Could not generate tokens : %v", err.Error()), http.StatusInternalServerError)
				return
			}
			tokenResponse := map[string]any{
				"access_token":  accessToken,
				"refresh_token": refreshToken,
				"verify_token":  verifyToken,
				"username":      user.Username,
				"isNotVarified": true,
			}

			response.WriteResponse(w, response.CreateResponse(tokenResponse, http.StatusOK, "Logged in successfully", "", "", false, ""))
			return
		}

		// Send tokens as JSON response
		tokenResponse := map[string]any{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"username":      user.Username,
			"isVarified":    user.IsVarified,
		}

		response.WriteResponse(w, response.CreateResponse(tokenResponse, http.StatusOK, "Logged in successfully", "", "", false, ""))
	}
}

func VerifyUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			http.Error(w, fmt.Sprintf("%v HTTP method is not allowed", r.Method), http.StatusBadRequest)
			return
		}

		userIDStr := r.Header.Get("userID")
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {

			http.Error(w, fmt.Sprintf("Invalid verification token : %v", err.Error()), http.StatusBadRequest)
			return

		}

		//Retrive the otp from the database:

		otp, err := database.RetrieveOTP(db, userID)

		if err != nil {
			http.Error(w, fmt.Sprintf("Error retrieving otp from Database, %v", err), http.StatusInternalServerError)
			return
		}

		var reqData struct {
			OTP string `json:"otp" validate:"required"`
		}

		if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		validate := validator.New()
		if err := validate.Struct(reqData); err != nil {
			response.ValidateResponse(w, err)
			return
		}

		if otp == reqData.OTP {

			if err := database.UpdateUserById(db, int64(userID), true); err != nil {
				http.Error(w, fmt.Sprintf("Error updating user from Database, %v", err), http.StatusInternalServerError)
				return
			}

			database.DeleteOTPbyUserId(db, userID)

			response.WriteResponse(w, response.CreateResponse(nil, http.StatusOK, "Verified successfully", "", "", false, ""))
			return

		} else {
			http.Error(w, "Wrong OTP", http.StatusBadRequest)
			return
		}

	}
}

func VerifyEmail(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			http.Error(w, fmt.Sprintf("%v HTTP method is not allowed", r.Method), http.StatusBadRequest)
			return
		}

		userIDStr := r.Header.Get("userID")
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {

			http.Error(w, fmt.Sprintf("Invalid verification token : %v", err.Error()), http.StatusBadRequest)
			return

		}

		//Retrive the otp from the database:

		otp, err := database.RetrieveOTP(db, userID)

		if err != nil {
			http.Error(w, fmt.Sprintf("Error retrieving otp from Database, %v", err), http.StatusInternalServerError)
			return
		}

		var reqData struct {
			OTP      string `json:"otp" validate:"required"`
			NewEmail string `json:"newEmail" validate:"required"`
		}

		if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		validate := validator.New()
		if err := validate.Struct(reqData); err != nil {
			response.ValidateResponse(w, err)
			return
		}

		if otp == reqData.OTP {

			if err := database.UpdateUserEmail(db, userID, reqData.NewEmail); err != nil {
				http.Error(w, fmt.Sprintf("Error updating email , %v", err), http.StatusInternalServerError)
				return
			}

			database.DeleteOTPbyUserId(db, userID)

			response.WriteResponse(w, response.CreateResponse(nil, http.StatusOK, "Verified successfully", "", "", false, ""))
			return

		} else {
			http.Error(w, "Wrong OTP", http.StatusBadRequest)
			return
		}

	}
}

func RequestNewOTP(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, fmt.Sprintf("%v HTTP method is not allowed", r.Method), http.StatusBadRequest)
			return
		}
		userIDStr := r.Header.Get("userID")
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {

			http.Error(w, fmt.Sprintf("Invalid verification token : %v", err.Error()), http.StatusBadRequest)
			return

		}

		//Generate OTP and update :

		otp := GenerateRandomString()

		err = database.UpdateOtpForUser(db, userID, otp)
		if err != nil {
			fmt.Println(err)
			_, err = database.InsertNewOTP(db, otp, int64(userID))

			if err != nil {
				http.Error(w, fmt.Sprintf("Database Error : %v", err.Error()), http.StatusBadRequest)
				return
			}
		}

		user, err := database.RetrieveUser(db, userID)
		if err != nil {

			http.Error(w, fmt.Sprintf("Invalid verification token : %v", err.Error()), http.StatusBadRequest)
			return

		}

		if err := email.SendOTPEmail(user.Email, otp); err != nil {
			http.Error(w, fmt.Sprintf("Error sending email : %v", err.Error()), http.StatusBadRequest)
			return
		}

		response.WriteResponse(w, response.CreateResponse(otp, http.StatusOK, "New OTP has been sent to the registered email"))

	}
}

func RequestNewOTPToVerifyEmail(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, fmt.Sprintf("%v HTTP method is not allowed", r.Method), http.StatusBadRequest)
			return
		}
		userIDStr := r.Header.Get("userID")
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {

			http.Error(w, fmt.Sprintf("Invalid verification token : %v", err.Error()), http.StatusBadRequest)
			return

		}

		var reqData struct {
			NewEmail string `json:"newEmail" validate:"required"`
		}

		if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		validate := validator.New()
		if err := validate.Struct(reqData); err != nil {
			response.ValidateResponse(w, err)
			return
		}
		//Generate OTP and update :

		otp := GenerateRandomString()

		err = database.UpdateOtpForUser(db, userID, otp)
		if err != nil {
			fmt.Println(err)
			_, err = database.InsertNewOTP(db, otp, int64(userID))

			if err != nil {
				http.Error(w, fmt.Sprintf("Database Error : %v", err.Error()), http.StatusBadRequest)
				return
			}
		}

		if err != nil {

			http.Error(w, fmt.Sprintf("Invalid verification token : %v", err.Error()), http.StatusBadRequest)
			return

		}

		if err := email.SendOTPEmail(reqData.NewEmail, otp); err != nil {
			http.Error(w, fmt.Sprintf("Error sending email : %v", err.Error()), http.StatusBadRequest)
			return
		}

		response.WriteResponse(w, response.CreateResponse(otp, http.StatusOK, "New OTP has been sent to the registered email"))

	}
}

func UpdateProfilePic(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPut {
			http.Error(w, fmt.Sprintf("%v HTTP method is not allowed", r.Method), http.StatusBadRequest)
			return
		}
		userIDStr := r.Header.Get("userID")
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {

			http.Error(w, fmt.Sprintf("Invalid verification token : %v", err.Error()), http.StatusBadRequest)
			return

		}

		var reqData struct {
			Url string `json:"profileImgUrl" validate:"required"`
		}

		if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		validate := validator.New()
		if err := validate.Struct(reqData); err != nil {
			response.ValidateResponse(w, err)
			return
		}

		cld, ctx, err := cloudinary.Credentials()
		if err != nil {
			fmt.Println("Error getting credentials", err)

			http.Error(w, fmt.Sprintf("Error getting credentials: %v", err.Error()), http.StatusInternalServerError)
		}

		url, err := cloudinary.UploadImage(cld, ctx, reqData.Url)
		if err != nil {
			fmt.Println("Error updating profile image to cloudinary", err)

			http.Error(w, fmt.Sprintf("Error updating profile image to cloudinary : %v", err.Error()), http.StatusInternalServerError)
			return
		}

		//Deleting prev url :

		user, err := database.RetrieveUser(db, userID)

		if err != nil {
			fmt.Println("Error retrieving user from database", err)

			http.Error(w, fmt.Sprintf("Error retrieving user from database : %v", err.Error()), http.StatusInternalServerError)
			return
		}

		prevImgUrl := user.ProfileImg

		if err := cloudinary.DeleteImage(cld, ctx, prevImgUrl); err != nil {
			fmt.Println("Error Deleting previous profile image", err)

			http.Error(w, fmt.Sprintf("Error Deleting previous profile image : %v", err.Error()), http.StatusInternalServerError)
			return
		}

		if err := database.UpdateUserProfilePic(db, userID, url); err != nil {

			fmt.Println("Error updating profile image", err)

			http.Error(w, fmt.Sprintf("Error updating profile image : %v", err.Error()), http.StatusInternalServerError)
			return
		}

		response.WriteResponse(w, response.CreateResponse(nil, http.StatusOK, "Profile picture updated"))
	}
}

func UpdateUserDetails(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPut {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		userIDStr := r.Header.Get("userID")
		userId, err := strconv.Atoi(userIDStr)
		if err != nil {

			http.Error(w, fmt.Sprintf("Invalid verification token : %v", err.Error()), http.StatusBadRequest)
			return

		}

		var request types.UserUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}

		if request.IsEmailChanged {
			u, _ := database.RetrieveUser(db, request.Email)
			if u != nil {
				http.Error(w, fmt.Sprintf("%v has already an account", u.Email), http.StatusInternalServerError)
				return
			}
			otp := GenerateRandomString()

			if _, err := database.InsertNewOTP(db, otp, int64(userId)); err != nil {
				http.Error(w, "Failed to insert OTP", http.StatusInternalServerError)
				return
			}

			if err := email.SendOTPEmail(request.Email, otp); err != nil {
				http.Error(w, fmt.Sprintf("Failed to send OTP %v", err), http.StatusInternalServerError)
				return
			}

		}

		if request.IsPasswordChanged {

			user, err := database.RetrieveUser(db, userId)
			if err != nil {
				http.Error(w, "Failed to retrieve user", http.StatusInternalServerError)
				return
			}

			correct, err := password.CheckPassword(request.CurrentPassword, user.Password)
			if err != nil {
				http.Error(w, "Failed to Check the Password", http.StatusInternalServerError)
				return
			}

			if correct {
				hashPassword, err := password.HashPassword(request.NewPassword)
				if err != nil {
					http.Error(w, "Failed to hash the password", http.StatusInternalServerError)
					return
				}

				if err := database.UpdatePassword(db, userId, hashPassword); err != nil {
					http.Error(w, "Failed to update password", http.StatusInternalServerError)
					return
				}
			}
		}

		if request.IsBioChanged {
			if err := database.UpdateBio(db, userId, request.Bio); err != nil {
				http.Error(w, "Failed to update bio", http.StatusInternalServerError)
				return
			}
		}

		if request.IsUsernameChanged {
			fmt.Println("DEBUG: request.IsUsernameChanged")
			u, _ := database.RetrieveUser(db, request.Username)
			if u != nil {
				fmt.Println("DEBUG:  is already taken . Try another ", u.Username)

				http.Error(w, fmt.Sprintf("%v is already taken . Try another ", u.Username), http.StatusInternalServerError)
				return
			}
			if err := database.UpdateUsername(db, userId, request.Username); err != nil {
				http.Error(w, fmt.Sprintf("Failed to update username : %v", err), http.StatusInternalServerError)
				return
			}
		}

		response.WriteResponse(w, response.CreateResponse(nil, http.StatusOK, "Profile details updated"))
	}
}
