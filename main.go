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
	cfg := config.MustLoad()
	db := database.ConnectToDatabase(cfg.PsqlInfo)

	// Initialize database tables
	database.CreateUserTable(db)
	database.CreateQuizzesTable(db)
	database.CreateQuestionsTable(db)
	database.DisplayData(db)

	// Configure CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"}, // Replace with your allowed origins
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	// Setup router
	router := http.NewServeMux()
	router.HandleFunc("/api/users/new", handlers.New(db))
	router.HandleFunc("/api/users/login", handlers.Login(db))
	router.HandleFunc("/api/quiz/generate", middlewares.AuthMiddleware(handlers.GenerateQuiz()))
	router.HandleFunc("/api/quiz/new", middlewares.AuthMiddleware(handlers.CreateQuizInDatabase(db)))
	router.HandleFunc("/api/quiz/questions/new", middlewares.AuthMiddleware(handlers.InsertQuestions(db)))
	router.HandleFunc("/api/quiz/quizzes", middlewares.AuthMiddleware(handlers.GetUserQuizzesHandler(db)))
	router.HandleFunc("/api/quiz/questions", middlewares.AuthMiddleware(handlers.GetQuizQuestionsHandler(db)))
	router.HandleFunc("/api/auth/me", middlewares.AuthMiddleware(middlewares.GetUserDetails()))

	// Wrap router with CORS middleware
	handler := c.Handler(router)

	// Setup server
	server := http.Server{
		Addr:    cfg.Addr,
		Handler: handler,
	}

	slog.Info("Server started at", slog.String("PORT", cfg.Addr))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		}
	}()

	<-done
	slog.Info("Shutting down the server...")

	// Logic to gracefully shut down the server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Failed to shut down server", slog.String("Error", err.Error()))
	} else {
		slog.Info("Server shut down successfully")
	}

	//database.DisplayData(db)
}
