package orm

import (
	"time"
)

type User struct {
	ID            int64  `gorm:"primary_key"`
	Name          string `gorm:"unique;not null"`
	Email         string `gorm:"not null"`
	EmailVerified bool
	Passhash      string `gorm:"not null"`
	Salt          string `gorm:"not null"`
	Iter          int    `gorm:"not null"`
	FamilyName    string
	GivenName     string
	Picture       string
	Nickname      string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Organization struct {
	Name        string `gorm:"primary_key"`
	Type        string `gorm:"not null"`
	Address     string
	Phone       string
	AdminUserID int64 `gorm:"type:bigint REFERENCES users(id)"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Controller struct {
	Region    string `gorm:"primary_key"`
	Address   string `gorm:"unique;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
