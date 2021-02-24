package main

import (
	"github.com/gofiber/fiber/v2"
	"log"
	"rabbit-shorten-url/internal/url"
)

func main() {
	app := fiber.New()
	urlService := url.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Get("/:code", urlService.Redirect)
	app.Post("/", urlService.Create)

	log.Fatal(app.Listen(":3000"))
}
