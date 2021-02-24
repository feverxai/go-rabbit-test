package mysql

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type MySql interface {
	Connect() *gorm.DB
}

type service struct {
	Config
}

func New(config Config) *service {
	return &service{
		Config: config,
	}
}

func (s *service) Connect() (*gorm.DB, error) {
	// refer https://github.com/go-sql-driver/mysql#dsn-data-source-name for details
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", s.Username, s.Password, s.Ip, s.Port, s.Database)
	return gorm.Open(mysql.Open(dsn), &gorm.Config{})
}
