package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"

	"net/http"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/database"
	//"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/generateQuiz"
	"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/response"
	"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/types"
)

const pythonServerProduction = "https://try-your-gyan-quiz-generation.onrender.com/generate-quiz"
//const pythonServerDev = "http://localhost:8000/generate-quiz"

// const pythonServerProduction = "https://try-your-gyan-quiz-generation.onrender.com/generate-quiz"

func normalizeTopic(topic string) string {
	topic = strings.ToLower(strings.TrimSpace(topic))
	phrasesToRemove := []string{"generate me a quiz on", "create a quiz about", "quiz on", "a quiz on"}
	for _, phrase := range phrasesToRemove {
		topic = strings.ReplaceAll(topic, phrase, "")
	}
	return strings.TrimSpace(topic)
}

func GenerateQuiz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := log.New(log.Writer(), "[GenerateQuiz] ", log.LstdFlags)

		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is allowed", http.StatusBadRequest)
			return
		}

		userIDStr := r.Header.Get("userID")
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			logger.Printf("Invalid userID: %v", err)
			http.Error(w, "Invalid userID", http.StatusBadRequest)
			return
		}

		var quizRequest types.QuizRequest
		quizRequest.UserID = int64(userID)

		if err := json.NewDecoder(r.Body).Decode(&quizRequest); err != nil {
			logger.Printf("Failed to decode JSON: %v", err)
			http.Error(w, "Failed to decode request body", http.StatusInternalServerError)
			return
		}

		// Normalize topic and difficulty
		quizRequest.Topic = normalizeTopic(quizRequest.Topic)
		quizRequest.Difficulty = strings.ToLower(quizRequest.Difficulty) // Ensure "medium", not "Medium"

		// Marshal struct
		jsonData, err := json.Marshal(quizRequest)
		if err != nil {
			logger.Printf("Failed to marshal quizRequest: %v", err)
			http.Error(w, "Failed to process request", http.StatusInternalServerError)
			return
		}

		logger.Printf("Sending to FastAPI: %s", string(jsonData))

		// Call FastAPI
		resp, err := http.Post(pythonServerProduction, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			logger.Printf("Failed to call FastAPI: %v", err)
			http.Error(w, "Failed to generate quiz", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// Log status and body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Printf("Failed to read FastAPI response: %v", err)
			http.Error(w, "Failed to read response", http.StatusInternalServerError)
			return
		}

		logger.Printf("FastAPI Status: %d", resp.StatusCode)
		logger.Printf("FastAPI Response: %s", string(body))

		// Check status
		if resp.StatusCode != http.StatusOK {
			logger.Printf("FastAPI non-200 status: %d", resp.StatusCode)
			http.Error(w, fmt.Sprintf("Quiz service error: %s", string(body)), resp.StatusCode)
			return
		}

		// Decode response as object
		var responseData map[string]interface{}
		if err := json.Unmarshal(body, &responseData); err != nil {
			logger.Printf("Failed to decode FastAPI response: %v", err)
			http.Error(w, "Invalid response from quiz service", http.StatusInternalServerError)
			return
		}

		// Validate response
		ok, isOkBool := responseData["ok"].(bool)
		if !isOkBool {
			logger.Printf("Invalid or missing 'ok' field")
			http.Error(w, "Invalid quiz response", http.StatusInternalServerError)
			return
		}

		if !ok {
			errorMsg := "Quiz generation failed"
			if errors, ok := responseData["data"].([]interface{}); ok && len(errors) > 0 {
				if msg, ok := errors[0].(string); ok {
					errorMsg = msg
				}
			}
			logger.Printf("FastAPI error: %s", errorMsg)
			http.Error(w, errorMsg, http.StatusInternalServerError)
			return
		}

		questions, isQuestionsArray := responseData["data"].([]interface{})
		if !isQuestionsArray || len(questions) == 0 {
			logger.Printf("No quiz questions generated")
			http.Error(w, "No quiz questions available", http.StatusInternalServerError)
			return
		}

		logger.Printf("Generated %d quiz questions", len(questions))
		respData := response.CreateResponse(questions, 200, "Quiz generated successfully")
		response.WriteResponse(w, respData)
	}
}

/*---------------------------------------*/
func CreateQuizInDatabase(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, fmt.Sprintf("%v HTTP method is not allowed", r.Method), http.StatusBadRequest)
			return
		}

		var quiz types.Quiz
		err := json.NewDecoder(r.Body).Decode(&quiz)
		if err != nil {
			if errors.Is(err, io.EOF) {
				http.Error(w, "no data to read", http.StatusBadRequest)
				return
			} else {
				http.Error(w, fmt.Sprintf("failed to decode JSON: %v", err), http.StatusInternalServerError)
				return
			}
		}

		validate := validator.New()
		if err := validate.Struct(&quiz); err != nil {
			response.ValidateResponse(w, err)
			return
		}

		if err := database.InsertNewQuiz(db, &quiz); err != nil {
			http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
			return
		}

		quizResponse := response.CreateResponse(quiz, http.StatusCreated, "Quiz created successfully", "<DeveloperMessage>", "<UserMessage>", false, "Err")
		response.WriteResponse(w, quizResponse)
	}
}

/*----------------------------------------------------------------------*/

func InsertQuestions(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, fmt.Sprintf("%v HTTP method is not allowed", r.Method), http.StatusMethodNotAllowed)
			return
		}

		var questions []types.Question

		// Decode JSON body into questions slice
		err := json.NewDecoder(r.Body).Decode(&questions)
		if err != nil {
			if errors.Is(err, io.EOF) {
				http.Error(w, "No data provided", http.StatusBadRequest)
			} else {
				http.Error(w, fmt.Sprintf("Failed to decode JSON: %v", err), http.StatusInternalServerError)
			}
			return
		}

		// Validate each question
		validate := validator.New()
		for _, question := range questions {

			if err := validate.Struct(question); err != nil {
				response.ValidateResponse(w, err)
				return
			}
		}

		tx, err := db.Begin()
		if err != nil {
			http.Error(w, fmt.Sprintf("Database transaction error: %v", err), http.StatusInternalServerError)
			return
		}

		// Insert questions in a single call
		if err := database.InsertNewQuestions(tx, questions); err != nil {
			tx.Rollback()
			http.Error(w, fmt.Sprintf("Failed to insert questions: %v", err), http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(); err != nil {
			http.Error(w, fmt.Sprintf("Transaction commit error: %v", err), http.StatusInternalServerError)
			return
		}

		successResponse := response.CreateResponse(nil, http.StatusCreated, "Questions inserted successfully", "<DeveloperMessage>", "<UserMessage>", false, "Err")
		response.WriteResponse(w, successResponse)
	}
}

func GetUserQuizzesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			http.Error(w, fmt.Sprintf("%v HTTP method is not allowed", r.Method), http.StatusMethodNotAllowed)
			return
		}

		userIDStr := r.Header.Get("userID")
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		quizzes, err := database.FetchQuizzesByUser(db, userID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error fetching quizzes %v", err.Error()), http.StatusInternalServerError)
			return
		}

		res := response.CreateResponse(quizzes, http.StatusOK, "Quizzes retrived successfully")

		response.WriteResponse(w, res)
	}
}

// GetQuizQuestionsHandler handles the endpoint to get questions for a quiz
func GetQuizQuestionsHandler(db *sql.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			http.Error(w, fmt.Sprintf("%v HTTP method is not allowed", r.Method), http.StatusMethodNotAllowed)
			return
		}

		quizIDStr := r.URL.Query().Get("quizID")

		quizID, err := strconv.Atoi(quizIDStr)

		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid Quiz Id : %v", err.Error()), http.StatusBadRequest)
			return
		}

		questions, err := database.FetchQuestionsByQuiz(db, quizID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error fetching questions : %v", err.Error()), http.StatusInternalServerError)
			return
		}

		quiz, err := database.FetchQuizzesByQuizId(db, quizID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error fetching quiz : %v", err.Error()), http.StatusInternalServerError)
			return
		}

		data := map[string]interface{}{
			"questions": questions,
			"quiz":      quiz,
		}
		res := response.CreateResponse(data, http.StatusOK, "Questions retrived successfully")

		response.WriteResponse(w, res)
	}
}

func DeleteQuiz(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, fmt.Sprintf("%v HTTP method is not allowed", r.Method), http.StatusMethodNotAllowed)
			return
		}

		quizIDStr := r.URL.Query().Get("quizID")

		quizID, err := strconv.Atoi(quizIDStr)

		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid Quiz Id : %v", err.Error()), http.StatusBadRequest)
			return
		}

		err = database.DeleteQuizById(db, quizID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error deleting Quiz : %v", err.Error()), http.StatusInternalServerError)
			return
		}

		res := response.CreateResponse(nil, http.StatusOK, "Quiz deleted successfully")

		response.WriteResponse(w, res)
	}
}
