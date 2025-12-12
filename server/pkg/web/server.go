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
	"github.com/Mahaveer86619/bookture/server/pkg/handlers"
	"github.com/Mahaveer86619/bookture/server/pkg/services"
)

type Server struct {
	cfg           config.Config
	router        *http.ServeMux
	healthHandler *handlers.HealthHandler
}

func NewServer(cfg config.Config) *Server {
	healthService := services.NewHealthService()
	healthHandler := handlers.NewHealthHandler(healthService)

	return &Server{
		cfg:           cfg,
		healthHandler: healthHandler,
		router:        http.NewServeMux(),
	}
}

func (s *Server) setupRoutes() {
	s.router.HandleFunc("/health", s.healthHandler.CheckHealth)
}

func (s *Server) Run() error {
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
		log.Printf("Server starting on port %s", s.cfg.PORT)
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
		log.Printf("Started %s %s", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)

		log.Printf("Completed %s %s in %v", r.Method, r.URL.Path, time.Since(start))
	})
}
