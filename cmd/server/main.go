package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/slack-go/slack"

	"github.com/aberyotaro/drink-tracker/internal/handlers"
	"github.com/aberyotaro/drink-tracker/internal/services"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./drink_tracker.db"
	}

	dbService, err := services.NewDatabaseService(dbPath)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer dbService.Close()

	userService := services.NewUserService(dbService.DB)
	drinkService := services.NewDrinkService(dbService.DB)

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	slackClient := slack.New(os.Getenv("SLACK_BOT_TOKEN"))
	signingSecret := os.Getenv("SLACK_SIGNING_SECRET")

	slackHandler := handlers.NewSlackHandler(slackClient, signingSecret, userService, drinkService)

	e.POST("/slack/command", slackHandler.HandleSlashCommand)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	e.Logger.Fatal(e.Start(":" + port))
}
