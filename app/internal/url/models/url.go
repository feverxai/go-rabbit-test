package models

import "time"

type Url struct {
	ShortCode  string `gorm:"primaryKey"`
	FullUrl    string
	ExpiryDate time.Time
	Hits       int
	IsDeleted  bool
}
