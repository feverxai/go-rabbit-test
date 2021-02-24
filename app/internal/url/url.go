package url

import (
	"github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"net/http"
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

// CreateRequest handle incoming post request to create new shorten url with expiry date ms (epoch millisecond)
type CreateRequest struct {
	Url          string `json:"url"`
	ExpiryDateMs int64  `json:"expiry_date_ms"`
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
		return c.Status(http.StatusBadRequest).JSON(ErrResponse{err.Error()})
	}

	if err := validation.Validate(req.Url,
		validation.Required,           // not empty
		validation.By(checkBlockList), // is a block list
		is.URL,                        // is a valid URL
	); err != nil {
		return c.Status(http.StatusBadRequest).JSON(ErrResponse{err.Error()})
	}

	if req.ExpiryDateMs > 0 && req.ExpiryDateMs <= time.Now().Unix()*int64(time.Millisecond) {
		return c.Status(http.StatusBadRequest).JSON(ErrResponse{Error: "expiry_date_ms must be future"})
	}
	expiryDate := time.Unix(0, req.ExpiryDateMs*int64(time.Millisecond))

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

	result := u.db.Create(&url)
	if result.Error != nil {
		return c.Status(http.StatusBadRequest).JSON(ErrResponse{result.Error.Error()})
	}

	return c.Status(http.StatusCreated).JSON(CreateResponse{url.ShortCode})
}

// Redirect is used to find valid service from shorten service then redirect to (302)
func (u *service) Redirect(c *fiber.Ctx) error {
	code := c.Params("code")

	var url models.Url
	u.db.First(&url, "short_code", code)

	return c.Redirect(url.FullUrl)
}
