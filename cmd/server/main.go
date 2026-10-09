package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"lofi-student-match/internal/config"
	"lofi-student-match/internal/handlers"
	"lofi-student-match/internal/services"
	realtime "lofi-student-match/internal/websocket"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberws "github.com/gofiber/websocket/v2"
)

func main() {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatal("MONGO_URI environment variable is required")
	}

	startupCtx, cancelStartup := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := config.ConnectMongo(startupCtx, mongoURI)
	cancelStartup()
	if err != nil {
		log.Fatalf("MongoDB startup failed: %v", err)
	}

	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	defer func() {
		disconnectCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := client.Disconnect(disconnectCtx); err != nil {
			log.Printf("MongoDB disconnect failed: %v", err)
		}
	}()

	app := fiber.New(fiber.Config{
		AppName:      "Student Matching API",
		ErrorHandler: apiErrorHandler,
	})
	frontendOrigins := strings.TrimSpace(os.Getenv("FRONTEND_ORIGINS"))
	if frontendOrigins == "" {
		frontendOrigins = "http://localhost:3001,http://127.0.0.1:3001"
	}
	allowedOrigins := strings.Split(frontendOrigins, ",")
	app.Use(cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool {
			return originAllowed(origin, allowedOrigins)
		},
		AllowMethods: "GET,POST,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept",
	}))
	database := client.Database(databaseName())
	studentHandler := handlers.NewStudentHandler(database)
	matchService, err := services.NewMatchService(database)
	if err != nil {
		log.Fatalf("initialize matchmaking service: %v", err)
	}
	hub := realtime.NewHub()
	matchHandler := handlers.NewMatchHandler(matchService, hub)
	chatHandler := handlers.NewChatHandler(hub, database)

	api := app.Group("/api")
	students := api.Group("/students")
	students.Post("/register", studentHandler.Register)
	students.Post("/login", studentHandler.Login)
	students.Get("/:id", studentHandler.GetByID)
	matches := api.Group("/match")
	matches.Post("/find", matchHandler.Find)
	matches.Get("/history/:student_id", matchHandler.History)
	matches.Get("/:match_id/messages", chatHandler.History)
	app.Get("/ws/chat", func(c *fiber.Ctx) error {
		if !originAllowed(c.Get(fiber.HeaderOrigin), allowedOrigins) {
			return fiber.ErrForbidden
		}
		return c.Next()
	}, fiberws.New(chatHandler.Handle, fiberws.Config{
		Origins: []string{"*"},
	}))

	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "3000"
	}

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- app.Listen(":" + port)
	}()

	select {
	case err := <-serverErrors:
		if err != nil && shutdownCtx.Err() == nil {
			log.Printf("HTTP server failed: %v", err)
		}
	case <-shutdownCtx.Done():
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := app.ShutdownWithContext(ctx); err != nil {
			log.Printf("HTTP server shutdown failed: %v", err)
		}
	}
}

func databaseName() string {
	name := strings.TrimSpace(os.Getenv("MONGO_DATABASE"))
	if name == "" {
		return "student_matching"
	}
	return name
}

func apiErrorHandler(c *fiber.Ctx, err error) error {
	status := fiber.StatusInternalServerError
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		status = fiberErr.Code
	}

	return c.Status(status).JSON(fiber.Map{"error": err.Error()})
}
