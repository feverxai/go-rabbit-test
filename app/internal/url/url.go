package url

import (
	"github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"rabbit-shorten-url/internal/url/models"
	"time"
)

// Service interface for url package
type Service interface {
	Redirect(c *fiber.Ctx) error
	Create(c *fiber.Ctx) error
}

type service struct {
	db *gorm.DB
}

// New initial url service with dbClient
func New(dbClient *gorm.DB) *service {
	return &service{
		db: dbClient,
	}
}

// CreateRequest handle incoming post request to create new shorten url with expiry (hour)
type CreateRequest struct {
	Url    string        `json:"url"`
	Expiry time.Duration `json:"expiry"`
}

// CreateResponse return shorten url of incoming request
type CreateResponse struct {
	ShortenUrl string `json:"shorten_url"`
}

// ErrResponse return error response with message
type ErrResponse struct {
	Error string `json:"error"`
}

// Create is used to generate shorten service from request
func (u *service) Create(c *fiber.Ctx) error {
	// shortCodeLength define characters of shortCode
	shortCodeLength := 8
	req := new(CreateRequest)

	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrResponse{err.Error()})
	}

	if err := validation.Validate(req.Url,
		validation.Required,           // not empty
		validation.By(checkBlockList), // is a block list
		is.URL,                        // is a valid URL
	); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrResponse{err.Error()})
	}

	expiryDate := time.Now().Add(req.Expiry * time.Hour)

	var shortCode string
	isShortCodeDuplicated := true
	for isShortCodeDuplicated {
		// check if short_code is duplicated or not
		shortCode = generateRandomString(shortCodeLength)
		result := u.db.First(&models.Url{}, "short_code", shortCode)
		if result.RowsAffected <= 0 {
			isShortCodeDuplicated = false
		}
	}

	url := models.Url{
		ShortCode:  shortCode,
		FullUrl:    req.Url,
		ExpiryDate: expiryDate,
	}

	u.db.Create(&url)

	return c.Status(fiber.StatusCreated).JSON(CreateResponse{c.Hostname() + "/" + url.ShortCode})
}

// Redirect is used to find valid service from shorten service then redirect to (302)
func (u *service) Redirect(c *fiber.Ctx) error {
	code := c.Params("code")

	var url models.Url
	u.db.First(&url, "short_code", code)

	if url.ExpiryDate.Sub(time.Now()) <= 0 || url.IsDeleted {
		return c.Status(fiber.StatusGone).SendString("expired")
	}

	url.Hits += 1
	u.db.Model(&url).Update("hits", url.Hits)

	return c.Redirect(url.FullUrl)
}
