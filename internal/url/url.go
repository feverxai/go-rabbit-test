package url

import (
	"github.com/gofiber/fiber/v2"
)

type Url interface {
	Redirect(c *fiber.Ctx) error
	Create(c *fiber.Ctx) error
}

type url struct{}

func New() *url {
	return &url{}
}

// Redirect is used to find valid url from shorten url then redirect to (302)
func (u *url) Redirect(c *fiber.Ctx) error {
	code := c.Params("code")
	redirect := "https://www.google.com"
	if code == "xai" {
		redirect = "https://docs.gofiber.io/"
	}

	return c.Redirect(redirect)
}

// Create is used to generate shorten url from request
func (u *url) Create(c *fiber.Ctx) error {
	return c.JSON("")
}
