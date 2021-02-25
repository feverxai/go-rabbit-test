package models

import "time"

type Url struct {
	ShortCode  string     `gorm:"primaryKey" json:"short_code"`
	FullUrl    string     `json:"full_url"`
	ExpiryDate *time.Time `json:"expiry_date"`
	Hits       int        `json:"hits"`
	IsDeleted  bool       `json:"is_deleted"`
}
