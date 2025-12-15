package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Mahaveer86619/bookture/server/pkg/config"
	"github.com/Mahaveer86619/bookture/server/pkg/db"
	"github.com/Mahaveer86619/bookture/server/pkg/handlers"
	"github.com/Mahaveer86619/bookture/server/pkg/middleware"
	"github.com/Mahaveer86619/bookture/server/pkg/services"
	"github.com/Mahaveer86619/bookture/server/pkg/services/storage"
)

type Server struct {
	cfg    config.Config
	router *http.ServeMux
}

func NewServer() *Server {
	return &Server{
		cfg:    config.AppConfig,
		router: http.NewServeMux(),
	}
}

func (s *Server) initSystem() {
	db.InitBookture()
}

func (s *Server) setupRoutes() {
	storageService := storage.NewStorageService()
	if err := storageService.Init(); err != nil {
		log.Fatalf("Fatal: Failed to initialize storage: %v", err)
	}

	healthService := services.NewHealthService(storageService)
	healthHandler := handlers.NewHealthHandler(healthService)

	userService := services.NewUserService()
	userHandler := handlers.NewUserHandler(userService)

	libraryService := services.NewLibraryService()
	libraryHandler := handlers.NewLibraryHander(libraryService)

	// Health
	s.router.HandleFunc("GET /health", healthHandler.CheckHealth)

	// Auth
	s.router.HandleFunc("POST /register", userHandler.Register)
	s.router.HandleFunc("POST /login", userHandler.Login)
	s.router.HandleFunc("POST /refresh", userHandler.RefreshToken)

	// Protected routes
	// User auth
	s.router.HandleFunc("GET /me", middleware.Middleware(userHandler.Me))
	s.router.HandleFunc("PUT /user", middleware.Middleware(userHandler.UpdateUser))
	s.router.HandleFunc("DELETE /user", middleware.Middleware(userHandler.DeleteHandler))

	// Library crud
	s.router.HandleFunc("POST /library", middleware.Middleware(libraryHandler.CreateNewLibrary))
	s.router.HandleFunc("GET /library", middleware.Middleware(libraryHandler.GetLibraries))
	s.router.HandleFunc("GET /library/get", middleware.Middleware(libraryHandler.GetLibrary))
	s.router.HandleFunc("PUT /library", middleware.Middleware(libraryHandler.UpdateLibrary))
	s.router.HandleFunc("DELETE /library", middleware.Middleware(libraryHandler.DeleteLibrary))

	// Book upload
}

func (s *Server) Run() error {
	s.initSystem()
	s.setupRoutes()

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", s.cfg.PORT),
		Handler:      s.loggingMiddleware(s.router),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Minute,
	}

	serverErrors := make(chan error, 1)

	go func() {
		fmt.Println(GetStartMessage(s.cfg.PORT))
		serverErrors <- srv.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Printf("Shutdown signal received: %v", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			return srv.Close()
		}
	}

	return nil
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Request -> %s %s", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)

		log.Printf("Response -> %s %s in %v", r.Method, r.URL.Path, time.Since(start))
	})
}
