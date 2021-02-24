package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"log"
	"rabbit-shorten-url/internal/db/mysql"
	"rabbit-shorten-url/internal/url"
)

func main() {
	app := fiber.New()
	db := mysql.New(mysql.Config{
		Username: "rabbit",
		Password: "password",
		Database: "rabbit",
		Ip:       "localhost",
		Port:     "6603",
	})
	dbClient, err := db.Connect()
	if err != nil {
		panic(err)
	}
	urlService := url.New(dbClient)

	app.Use(logger.New())

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Get("/:code", urlService.Redirect)
	app.Post("/", urlService.Create)

	log.Fatal(app.Listen(":3000"))
}
