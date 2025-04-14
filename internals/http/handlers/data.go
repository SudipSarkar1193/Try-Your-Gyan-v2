package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"net/http"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/database"
	//"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/generateQuiz"
	"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/response"
	"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/types"
)

const pythonServerDev = "http://localhost:8000/generate-quiz"
const pythonServerProduction = "https://try-your-gyan-quiz-generation.onrender.com/generate-quiz"

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

		if r.Method != http.MethodPost {
			http.Error(w, fmt.Sprintf("%v HTTP method is not allowed", r.Method), http.StatusBadRequest)
			return
		}
		fmt.Println("INSIDE")
		userIDStr := r.Header.Get("userID")
		userID, err := strconv.Atoi(userIDStr)

		if err != nil {
			fmt.Println("Error converting userID:", err)
			http.Error(w, fmt.Sprintf("Invalid verification token : %v", err.Error()), http.StatusBadRequest)
			return

		}

		var quizRequest types.QuizRequest

		quizRequest.UserID = int64(userID)

		if err := json.NewDecoder(r.Body).Decode(&quizRequest); err != nil {
			http.Error(w, fmt.Sprintf("failed to decode JSON: %v", err.Error()), http.StatusInternalServerError)
			return
		} 

		// Normalize topic (example implementation)
		normalizedTopic := normalizeTopic(quizRequest.Topic)
		quizRequest.Topic = normalizedTopic

		// Call Python FastAPI service
		jsonData, err := json.Marshal(quizRequest)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to Marshal 'quizRequest': %v", err.Error()), http.StatusInternalServerError)
			return
		}

		resp, err := http.Post(pythonServerProduction, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to call quiz service: %v", err.Error()), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		var data interface{}
		json.NewDecoder(resp.Body).Decode(&data)

		// Check for error structure
		if dataMap, ok := data.(map[string]interface{}); ok {
			if detail, exists := dataMap["detail"]; exists {
				if detailStr, ok := detail.(string); ok && detailStr == "Internal server error" {
					http.Error(w, "Quiz generation failed", http.StatusInternalServerError)
					return
				}
			}
		}

		respData := response.CreateResponse(data, 200, "Quiz generated successfully")
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
