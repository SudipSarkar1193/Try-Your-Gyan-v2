package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log/slog"

	"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/config"
	"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/database"
	"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/http/handlers"
	"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/middlewares"
	"github.com/rs/cors"
)

func main() {
	// Load configuration and connect to the database
	cfg := config.MustLoad()
	db := database.ConnectToDatabase(cfg.PsqlInfo)

	// if err := config.LoadEnvFile(".env"); err != nil {
	// 	log.Println("Error loading Env file", err)
	// }

	// Initialize Firebase Auth client
	handlers.InitializeFirebaseApp()

	// Configure CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "https://try-your-gyan.vercel.app"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "Access-Control-Allow-Origin"},
		AllowCredentials: true,
		Debug:            true,
	})

	// Initialize router
	router := http.NewServeMux()

	// Set up routes
	router.HandleFunc("/api/users/new", handlers.New(db))
	router.HandleFunc("/api/users/login", handlers.Login(db))
	router.HandleFunc("/api/users/auth/google", handlers.HandleFirebaseAuth(db))
	router.HandleFunc("/api/users/auth/verify", middlewares.VerifyUserMiddleware(handlers.VerifyUser(db)))
	router.HandleFunc("/api/users/auth/newotp", middlewares.VerifyUserMiddleware(handlers.RequestNewOTP(db)))
	router.HandleFunc("/api/users/update-profile-pic", middlewares.AuthMiddleware(handlers.UpdateProfilePic(db)))
	router.HandleFunc("/api/quiz/generate", middlewares.AuthMiddleware(handlers.GenerateQuiz()))
	router.HandleFunc("/api/quiz/new", middlewares.AuthMiddleware(handlers.CreateQuizInDatabase(db)))
	router.HandleFunc("/api/quiz/questions/new", middlewares.AuthMiddleware(handlers.InsertQuestions(db)))
	router.HandleFunc("/api/quiz/quizzes", middlewares.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetUserQuizzesHandler(db)(w, r)
		case http.MethodDelete:
			handlers.DeleteQuiz(db)(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	router.HandleFunc("/api/quiz/questions", middlewares.AuthMiddleware(handlers.GetQuizQuestionsHandler(db)))
	router.HandleFunc("/api/auth/me", middlewares.AuthMiddleware(middlewares.GetUserDetails(db)))

	// Wrap with middlewares
	handler := c.Handler(router)
	handler = middlewares.HandleOptionsMiddleware(handler)
	handler = middlewares.CoopMiddleware(handler)

	// Setup HTTP server
	server := http.Server{
		Addr:    cfg.Addr,
		Handler: handler,
	}

	slog.Info("Server started at", slog.String("PORT", cfg.Addr))

	// Graceful shutdown setup
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		}
	}()

	<-done
	slog.Info("Shutting down the server...")

	// Gracefully shutdown the server with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Failed to shut down server", slog.String("Error", err.Error()))
	} else {
		slog.Info("Server shut down successfully")
	}
}
