package ormapi

import (
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

// Data saved to persistent sql db, also used for API calls

type User struct {
	Name          string `gorm:"primary_key"`
	Email         string `gorm:"unique;not null"`
	EmailVerified bool
	Passhash      string `gorm:"not null"`
	Salt          string `gorm:"not null"`
	Iter          int    `gorm:"not null"`
	FamilyName    string
	GivenName     string
	Picture       string
	Nickname      string
	CreatedAt     time.Time `json:",omitempty"`
	UpdatedAt     time.Time `json:",omitempty"`
}

type Organization struct {
	Name          string `gorm:"primary_key"`
	Type          string `gorm:"not null"`
	Address       string
	Phone         string
	AdminUsername string    `gorm:"type:text REFERENCES users(name)"`
	CreatedAt     time.Time `json:",omitempty"`
	UpdatedAt     time.Time `json:",omitempty"`
}

type Controller struct {
	Region    string    `gorm:"primary_key"`
	Address   string    `gorm:"unique;not null"`
	CreatedAt time.Time `json:",omitempty"`
	UpdatedAt time.Time `json:",omitempty"`
}

// Structs used for API calls

type RolePerm struct {
	Role     string `json:"role"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

type Role struct {
	Org      string `form:"org" json:"org"`
	Username string `form:"username" json:"username"`
	Role     string `form:"role" json:"role"`
}

type UserLogin struct {
	Username string `form:"username" json:"username"`
	Password string `form:"password" json:"password"`
}

type NewPassword struct {
	Password string `form:"password" json:"password"`
}

// Structs used in replies

type Result struct {
	Message string `json:"message"`
}

type ResultID struct {
	Message string `json:"message"`
	ID      int64  `json:"id"`
}

type ResultName struct {
	Message string `json:"message"`
	Name    string `json:"name"`
}

// all data is for full create/delete

type AllData struct {
	Controllers []Controller   `json:"controllers,omitempty"`
	Orgs        []Organization `json:"orgs,omitempty"`
	Roles       []Role         `json:"roles,omitempty"`
	// not supported yet
	RegionData []RegionData `json:"regiondata,omitempty"`
}

type RegionData struct {
	Region  string                    `json:"region,omitempty"`
	AppData edgeproto.ApplicationData `json:"appdata,omitempty"`
}
