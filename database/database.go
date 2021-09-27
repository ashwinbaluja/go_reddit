package database

import (
	"gorm.io/gorm"
)

var (
	DBConn *gorm.DB
)

type User struct {
	gorm.Model
	Username       string `json:"username"`
	Bio            string `json:"bio"`
	HashedPassword []byte `json:"hashed"`
}

type Post struct {
	gorm.Model
	Title     string `json:"title"`
	Body      string `json:"body"`
	Upvotes   int    `json:"upvotes"`
	Downvotes int    `json:"downvotes"`
	UserID    int
}

type Subspace struct {
	gorm.Model
	Subspacename string `json:"subspacename"`
	UserID       int
}
