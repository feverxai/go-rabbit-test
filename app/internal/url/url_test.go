package url

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"io/ioutil"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"
)

type TSuite struct {
	suite.Suite
	DB   *gorm.DB
	mock sqlmock.Sqlmock
}

func TestUrlSuite(t *testing.T) {
	suite.Run(t, new(TSuite))
}

func (s *TSuite) SetupSuite() {
	db, mock, err := sqlmock.New()
	s.mock = mock
	require.NoError(s.T(), err)
	s.DB, err = gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	require.NoError(s.T(), err)

	s.DB.Debug()
}

func (s *TSuite) TestCreateUrl_ShouldReturnBodyParserError() {
	u := New(s.DB)
	app := fiber.New()
	app.Post("/", u.Create)

	const reqBody = `{
		"url": "https://docs.gofiber.io/",
		"expiry": "not allow"
	}`

	req := httptest.NewRequest("POST", "/", strings.NewReader(reqBody))
	req.Header.Add("Content-Type", "application/json")
	res, _ := app.Test(req, -1)

	s.Assert().Equal(fiber.StatusBadRequest, res.StatusCode)
}

func (s *TSuite) TestCreateUrl_UrlIsNotValid() {
	u := New(s.DB)
	app := fiber.New()
	app.Post("/", u.Create)

	const reqBody = `{
		"url": "not a valid",
		"expiry": 0
	}`

	req := httptest.NewRequest("POST", "/", strings.NewReader(reqBody))
	req.Header.Add("Content-Type", "application/json")

	res, _ := app.Test(req, -1)
	body, _ := ioutil.ReadAll(res.Body)

	s.Assert().Equal(fiber.StatusBadRequest, res.StatusCode)
	s.Assert().Contains(string(body), is.ErrURL.Message())
}

func (s *TSuite) TestCreateUrl_UrlIsBlockList() {
	u := New(s.DB)
	app := fiber.New()
	app.Post("/", u.Create)

	const reqBody = `{
		"url": "https://www.facebook.com/",
		"expiry": 0
	}`

	req := httptest.NewRequest("POST", "/", strings.NewReader(reqBody))
	req.Header.Add("Content-Type", "application/json")

	res, _ := app.Test(req, -1)
	body, _ := ioutil.ReadAll(res.Body)

	s.Assert().Equal(fiber.StatusBadRequest, res.StatusCode)
	s.Assert().Contains(string(body), ErrURLBlockList.Error())
}

func (s *TSuite) TestCreateUrl_Success() {
	u := New(s.DB)
	app := fiber.New()
	app.Post("/", u.Create)

	rs := sqlmock.NewRows([]string{"short_code"})
	s.mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `urls` WHERE `short_code` = ? ORDER BY `urls`.`short_code` LIMIT 1")).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(rs)

	s.mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `urls` (`short_code`,`full_url`,`expiry_date`,`hits`,`is_deleted`) VALUES (?,?,?,?,?)")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	var reqBody = `{
		"url": "https://docs.gofiber.io/",
		"expiry": 24
	}`

	req := httptest.NewRequest("POST", "/", strings.NewReader(reqBody))
	req.Header.Add("Content-Type", "application/json")

	res, _ := app.Test(req, -1)

	s.Assert().Equal(fiber.StatusCreated, res.StatusCode)
}

func (s *TSuite) TestCreateUrl_SuccessButShortCodeIsDuplicated() {
	u := New(s.DB)
	app := fiber.New()
	app.Post("/", u.Create)

	rs := sqlmock.NewRows([]string{"short_code"})
	s.mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `urls` WHERE `short_code` = ? ORDER BY `urls`.`short_code` LIMIT 1")).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(rs.AddRow("duplicated"))

	s.mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `urls` WHERE `short_code` = ? ORDER BY `urls`.`short_code` LIMIT 1")).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(rs)

	s.mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `urls` (`short_code`,`full_url`,`expiry_date`,`hits`,`is_deleted`) VALUES (?,?,?,?,?)")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	var reqBody = `{
		"url": "https://docs.gofiber.io/",
		"expiry": 24
	}`

	req := httptest.NewRequest("POST", "/", strings.NewReader(reqBody))
	req.Header.Add("Content-Type", "application/json")

	res, _ := app.Test(req, -1)

	s.Assert().Equal(fiber.StatusCreated, res.StatusCode)
}

func (s *TSuite) TestRedirectUrl_ShortCodeIsExpired() {
	u := New(s.DB)
	app := fiber.New()
	app.Get("/:code", u.Redirect)
	shortCode := "test1234"

	rs := sqlmock.NewRows([]string{"short_code", "full_url", "expiry_date", "hits", "is_deleted"})
	s.mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `urls` WHERE `short_code` = ? ORDER BY `urls`.`short_code` LIMIT 1")).
		WithArgs(shortCode).
		WillReturnRows(rs.AddRow(shortCode, "https://www.google.com", time.Now().Add(-1*time.Hour), 0, 0))

	req := httptest.NewRequest("GET", "/"+shortCode, nil)
	req.Header.Add("Content-Type", "application/json")

	res, _ := app.Test(req, -1)

	s.Assert().Equal(fiber.StatusGone, res.StatusCode)
}

func (s *TSuite) TestRedirectUrl_ShortCodeIsDeleted() {
	u := New(s.DB)
	app := fiber.New()
	app.Get("/:code", u.Redirect)
	shortCode := "test1234"

	rs := sqlmock.NewRows([]string{"short_code", "full_url", "expiry_date", "hits", "is_deleted"})
	s.mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `urls` WHERE `short_code` = ? ORDER BY `urls`.`short_code` LIMIT 1")).
		WithArgs(shortCode).
		WillReturnRows(rs.AddRow(shortCode, "https://www.google.com", time.Now().Add(time.Hour), 0, 1))

	req := httptest.NewRequest("GET", "/"+shortCode, nil)
	req.Header.Add("Content-Type", "application/json")

	res, _ := app.Test(req, -1)

	s.Assert().Equal(fiber.StatusGone, res.StatusCode)
}

func (s *TSuite) TestRedirectUrl_Success() {
	u := New(s.DB)
	app := fiber.New()
	app.Get("/:code", u.Redirect)
	shortCode := "test1234"
	hits := 0
	rs := sqlmock.NewRows([]string{"short_code", "full_url", "expiry_date", "hits", "is_deleted"})
	s.mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `urls` WHERE `short_code` = ? ORDER BY `urls`.`short_code` LIMIT 1")).
		WithArgs(shortCode).
		WillReturnRows(rs.AddRow(shortCode, "https://www.google.com", time.Now().Add(time.Hour), hits, 0))
	s.mock.ExpectExec(regexp.QuoteMeta("UPDATE `urls` SET `hits`=? WHERE `short_code` = ?")).
		WithArgs(hits + 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("GET", "/"+shortCode, nil)
	req.Header.Add("Content-Type", "application/json")

	res, _ := app.Test(req, -1)

	s.Assert().Equal(fiber.StatusFound, res.StatusCode)
}

func (s *TSuite) AfterTest(_, _ string) {
	require.NoError(s.T(), s.mock.ExpectationsWereMet())
}