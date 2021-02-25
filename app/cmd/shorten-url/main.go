package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"gorm.io/gorm"
	"log"
	"os"
	"os/signal"
	"rabbit-shorten-url/internal/db/mysql"
	"rabbit-shorten-url/internal/url"
	"time"
)

func main() {

	db := mysql.New(mysql.Config{
		Username: os.Getenv("DB_USERNAME"),
		Password: os.Getenv("DB_PASSWORD"),
		Database: os.Getenv("DB_DATABASE"),
		Ip:       os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
	})
	dbClient, err := db.Connect()
	if err != nil {
		log.Fatal(err)
	}

	app := Setup(dbClient)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		_ = <-c
		fmt.Println("Gracefully shutting down...")
		_ = app.Shutdown()
	}()

	if err := app.Listen(":3000"); err != nil {
		log.Panic(err)
	}

	fmt.Println("Running cleanup tasks...")
	if err := db.Close(dbClient); err != nil {
		log.Panic(err)
	}
}

func Setup(dbClient *gorm.DB) *fiber.App {
	app := fiber.New()

	urlService := url.New(dbClient)

	app.Use(logger.New())
	app.Use(cache.New(cache.Config{
		Expiration: 30 * time.Minute,
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Get("/:code", urlService.Redirect)
	app.Post("/", urlService.Create)

	// group route for admin auth
	admin := app.Group("/admin", basicauth.New(basicauth.Config{
		Users: map[string]string{
			"admin": "demo",
		},
	}))
	admin.Get("/urls/:code?", urlService.List)
	admin.Delete("/urls/:code", urlService.SoftDelete)

	return app
}
