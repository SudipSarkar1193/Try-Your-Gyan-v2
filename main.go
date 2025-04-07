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
		{"/api/users/new", "POST", handlers.New(db), false},
		{"/api/users/login", "POST", handlers.Login(db), false},
		{"/api/users/auth/google", "POST", handlers.HandleFirebaseAuth(db), false},
		{"/api/users/auth/verify", "POST", handlers.VerifyUser(db), true},
		{"/api/users/auth/newotp", "POST", handlers.RequestNewOTP(db), true},
		{"/api/users/newotp", "POST", handlers.RequestNewOTPToVerifyEmail(db), true},
		{"/api/users/update-profile-pic", "PUT", handlers.UpdateProfilePic(db), true},
		{"/api/users/verify-email", "POST", handlers.VerifyEmailToUpdate(db, client), true},
		{"/api/users/update-profile", "PUT", handlers.UpdateUserDetails(db), true},
		{"/api/quiz/generate", "POST", handlers.GenerateQuiz(), false},
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
		log.Printf("Registered route: %s [%s, OPTIONS]", route.Path, route.Method)
	}
}

func main() {
	cfg := config.MustLoad()
	db := database.ConnectToDatabase(cfg.PsqlInfo)
	if db == nil {
		log.Fatal("Database connection failed")
	}

	if err := config.LoadEnvFile(".env"); err != nil {
		slog.Warn("Error loading .env file", slog.String("error", err.Error()))
	}

	client := handlers.InitializeFirebaseApp()
	if client == nil {
		log.Fatal("Firebase initialization failed")
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"https://try-your-gyan.vercel.app", "http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		Debug:            true, // Enable CORS debug logs
	})

	router := mux.NewRouter()
	registerRoutes(router, db, client)

	router.Methods(http.MethodOptions).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
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
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		} else {
			slog.Info("Server started at", slog.String("PORT", cfg.Addr))
		}
	}()

	<-done
	slog.Info("Shutting down the server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Failed to shut down server", slog.String("Error", err.Error()))
	} else {
		slog.Info("Server shut down successfully")
	}
}
