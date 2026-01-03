package main

import (
	"fmt"
	"log"

	"github.com/Mahaveer86619/bookture/pkg/config"
	"github.com/Mahaveer86619/bookture/pkg/db"
	"github.com/Mahaveer86619/bookture/pkg/errz"
	"github.com/Mahaveer86619/bookture/pkg/handlers"
	"github.com/Mahaveer86619/bookture/pkg/middleware"
	"github.com/Mahaveer86619/bookture/pkg/services"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

func main() {
	// 1. Config & DB
	config.LoadConfig()
	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to init DB: %v", err)
	}

	// 2. Echo Instance
	e := echo.New()
	e.Use(echoMiddleware.RequestLogger())
	e.Use(echoMiddleware.Recover())
	e.Use(echoMiddleware.CORS())

	e.HTTPErrorHandler = errz.HandleErrors

	// 3. Services & Handlers
	healthService := services.NewHealthService()
	authService := services.NewAuthService()
	userService := services.NewUserService()

	healthHandler := handlers.NewHealthHandler(healthService)
	authHandler := handlers.NewAuthHandler(*authService)
	userHandler := handlers.NewUserHandler(userService)

	// 4. Routes
	api := e.Group("/api/v1")

	// --- Health (Public) ---
	api.GET("/health", healthHandler.Check)

	// --- Auth (Public) ---
	api.POST("/auth/register", authHandler.Register)
	api.POST("/auth/login", authHandler.Login)
	api.POST("/auth/refresh", authHandler.RefreshToken)

	// --- User (Public) ---
	api.GET("/users/:id", userHandler.GetProfile)
	api.GET("/users/:id/followers", userHandler.GetFollowers)
	api.GET("/users/:id/following", userHandler.GetFollowing)

	// --- User (Protected) ---
	// Create a group that requires JWT
	protected := api.Group("")
	protected.Use(middleware.Middleware)

	// Profile Management
	protected.GET("/users/me", userHandler.GetMe)
	protected.PUT("/users/me", userHandler.UpdateMe)

	// Actions
	protected.POST("/users/:id/follow", userHandler.Follow)
	protected.DELETE("/users/:id/follow", userHandler.Unfollow)

	// 5. Start Server
	addr := fmt.Sprintf(":%d", config.AppConfig.Port)
	e.Logger.Fatal(e.Start(addr))
}
