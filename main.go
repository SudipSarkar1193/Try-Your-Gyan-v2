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

	if err := config.LoadEnvFile(".env"); err != nil {
		log.Println("Error loading Env file", err)
	}

	// Initialize Firebase Auth client
	handlers.InitializeFirebaseApp()

	// Configure CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "https://try-your-gyan.vercel.app"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		Debug:            true, // Enable for debugging CORS issues
	})

	// Initialize router
	router := http.NewServeMux()

	// Set up routes
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("DEBUG: Fallback route hit, Path: %s, Method: %s\n", r.URL.Path, r.Method)
		http.NotFound(w, r)
	})
	router.HandleFunc("/api/users/new", handlers.New(db))
	router.HandleFunc("/api/users/login", handlers.Login(db))
	router.HandleFunc("/api/users/auth/google", handlers.HandleFirebaseAuth(db))
	router.HandleFunc("/api/users/auth/verify", middlewares.VerifyUserMiddleware(handlers.VerifyUser(db)))
	router.HandleFunc("/api/users/auth/newotp", middlewares.VerifyUserMiddleware(handlers.RequestNewOTP(db)))
	// router.HandleFunc("/api/users/update-profile-pic", middlewares.AuthMiddleware(handlers.UpdateProfilePic(db)))

	router.HandleFunc("/api/users/update-profile-pic", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("DEBUG: Reached /api/users/update-profile-pic, Method: %s\n", r.Method)
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
			w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.WriteHeader(http.StatusOK)
			log.Println("DEBUG: Preflight request handled")
			return
		}
		log.Printf("DEBUG: Actual request handled, Method: %s\n", r.Method)
		middlewares.AuthMiddleware(handlers.UpdateProfilePic(db)).ServeHTTP(w, r)
	})

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

	//DEBUG :

	router.HandleFunc("/api/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("CORS is working!"))
	})
	// // Combine middlewares: CORS first, then COOP
	// handler := c.Handler(router)
	// handler = middlewares.CoopMiddleware(handler)

	///

	// handler := middlewares.DebugOriginMiddleware(c.Handler(router))
	// handler = middlewares.CoopMiddleware(handler)
	// handler = middlewares.DebugMiddleware(handler)

	handler := middlewares.DebugOriginMiddleware(c.Handler(router))
	handler = middlewares.HandleOptionsMiddleware(handler)
	handler = middlewares.CoopMiddleware(handler)
	handler = middlewares.DebugMiddleware(handler)

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
