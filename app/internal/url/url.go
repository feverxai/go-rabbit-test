package url

import (
	"errors"
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
	List(c *fiber.Ctx) error
	SoftDelete(c *fiber.Ctx) error
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

// SuccessResponse return response with message
type SuccessResponse struct {
	Message string `json:"message"`
}

var (
	ErrExpired  = errors.New("expired")
	ErrNotFound = errors.New("not found")
)

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

	var expiryDate *time.Time
	if req.Expiry > 0 {
		exp := time.Now().Add(req.Expiry * time.Hour)
		expiryDate = &exp
	}

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

	if url.ExpiryDate != nil && (url.ExpiryDate.Sub(time.Now()) <= 0 || url.IsDeleted) {
		return c.Status(fiber.StatusGone).JSON(ErrResponse{ErrExpired.Error()})
	}

	url.Hits += 1
	u.db.Model(&url).Update("hits", url.Hits)

	return c.Redirect(url.FullUrl)
}

// List is used to list details by short_code or keyword on full_url
func (u *service) List(c *fiber.Ctx) error {
	code := c.Params("code")
	fullUrl := c.Query("full_url")
	var url []models.Url

	if code != "" {
		result := u.db.First(&url, "short_code", code)
		if result.RowsAffected <= 0 {
			return c.Status(fiber.StatusNotFound).JSON(ErrResponse{ErrNotFound.Error()})
		}
		return c.JSON(url[0])
	}

	// init chain orm
	tx := u.db
	if fullUrl != "" {
		tx = u.db.Where("full_url LIKE ?", "%"+fullUrl+"%")
	}

	tx.Find(&url)

	return c.JSON(url)
}

// SoftDelete is used to mark flag is_deleted = true by short_code
func (u *service) SoftDelete(c *fiber.Ctx) error {
	code := c.Params("code")

	result := u.db.Model(&models.Url{}).Where("short_code = ?", code).Update("is_deleted", true)
	if result.RowsAffected <= 0 {
		return c.Status(fiber.StatusNotFound).JSON(ErrResponse{ErrNotFound.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse{code + " has been deleted"})
}
