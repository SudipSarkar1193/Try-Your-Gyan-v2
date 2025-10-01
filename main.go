package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"encoding/json"

	"fmt"
	"log/slog"

	"firebase.google.com/go/v4/auth"
	"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/config"
	"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/database"
	"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/http/handlers"
	"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/middlewares"
	"github.com/rs/cors"

	"github.com/gorilla/mux"
)

// Route struct to store API endpoints
type Route struct {
	Path    string
	Method  string
	Handler http.HandlerFunc
	Auth    bool // Apply authentication middleware if true
}

// Function to return all API routes
func getRoutes(db *sql.DB, client *auth.Client) []Route {
	return []Route{
		{"/", "GET", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				http.Error(w, fmt.Sprintf("%v HTTP method is not allowed", r.Method), http.StatusBadRequest)
				return
			}
		
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
		
			resp := map[string]string{
				"status": "ok",
				"message":"Welcome to try-your-gyan from '/'",
			}
			json.NewEncoder(w).Encode(resp)
		}, false},
		{"/health", "GET", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				http.Error(w, fmt.Sprintf("%v HTTP method is not allowed", r.Method), http.StatusBadRequest)
				return
			}
		
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
		
			resp := map[string]string{
				"status": "ok",
				"message":"Welcome to try-your-gyan from '/health'",
			}
			json.NewEncoder(w).Encode(resp)
		}, false},
		{"/api/users/new", "POST", handlers.New(db), false},
		{"/api/users/login", "POST", handlers.Login(db), false},
		{"/api/users/auth/google", "POST", handlers.HandleFirebaseAuth(db), false},
		{"/api/users/auth/verify", "POST", handlers.VerifyUser(db), true},
		{"/api/users/auth/newotp", "POST", handlers.RequestNewOTP(db), true},
		{"/api/users/newotp", "POST", handlers.RequestNewOTPToVerifyEmail(db), true},
		{"/api/users/update-profile-pic", "PUT", handlers.UpdateProfilePic(db), true},
		{"/api/users/verify-email", "POST", handlers.VerifyEmailToUpdate(db, client), true},
		{"/api/users/update-profile", "PUT", handlers.UpdateUserDetails(db), true},
		{"/api/quiz/generate", "POST", handlers.GenerateQuiz(), true},
		{"/api/quiz/new", "POST", handlers.CreateQuizInDatabase(db), true},
		{"/api/quiz/questions/new", "POST", handlers.InsertQuestions(db), true},
		{"/api/quiz/quizzes", "GET", handlers.GetUserQuizzesHandler(db), true},
		{"/api/quiz/quizzes", "DELETE", handlers.DeleteQuiz(db), true},
		{"/api/quiz/questions", "GET", handlers.GetQuizQuestionsHandler(db), false},
		{"/api/auth/me", "GET", middlewares.GetUserDetails(db), true},
	}
}

// Register routes dynamically using Gorilla Mux
func registerRoutes(router *mux.Router, db *sql.DB, client *auth.Client) {
	for _, route := range getRoutes(db, client) {
		handler := route.Handler
		if route.Auth {
			handler = middlewares.AuthMiddleware(handler)
		}
		router.HandleFunc(route.Path, handler).Methods(route.Method, "OPTIONS")
		
		log.Printf("Registered route: %s [%s]", route.Path, route.Method)
	}
}

func main() {
	log.Println("Welcome to GO backend")
	cfg := config.MustLoad()
	db := database.ConnectToDatabase(cfg.PsqlInfo)
	if db == nil {
		log.Fatal("Database connection failed")
	}

	client := handlers.InitializeFirebaseApp()
	if client == nil {
		log.Fatal("Firebase initialization failed")
	}

	origins := []string{"https://try-your-gyan.vercel.app", "http://localhost:5173"}
	if localOrigin := os.Getenv("CORS_LOCAL_ORIGIN"); localOrigin != "" {
		origins = append(origins, localOrigin)
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "userID"},
		AllowCredentials: true,
		Debug:            false, // Back to false
	})

	router := mux.NewRouter()
	registerRoutes(router, db, client)

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Database unavailable"))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	handler := c.Handler(router)
	handler = middlewares.CoopMiddleware(handler)

	server := http.Server{
		Addr:    cfg.Addr,
		Handler: handler,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("Server starting at", slog.String("PORT", cfg.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		}
	}()

	<-done
	slog.Info("Shutting down the server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Failed to shut down server", slog.String("error", err.Error()))
	} else {
		slog.Info("Server shut down successfully")
	}
}