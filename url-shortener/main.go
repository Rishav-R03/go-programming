package main

import (
	"log"
	"os"
	"url-shortener/api/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func setupRoutes(app *fiber.App) {
	app.Get("/:url", routes.ResolveURL)
	app.Post("/api/v1", routes.ShortenURL)
}

func main() {
	_ = godotenv.Load()
	app := fiber.New()
	app.Use(logger.New())
	setupRoutes(app)
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "3000"
	}
	address := port
	if address[0] != ':' {
		address = ":" + address
	}
	log.Fatal(app.Listen(os.Getenv("APP_PORT")))
}
