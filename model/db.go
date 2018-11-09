package model

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	_ "github.com/mattn/go-sqlite3"
)

var (
	Db *gorm.DB
)

func Connect(dbPath string) *gorm.DB {
	var err error
	Db, err = gorm.Open("sqlite3", dbPath)
	if err != nil {
		panic(err)
	}

	Db.AutoMigrate(&User{}, &Page{}, &ContentVersion{})

	return Db
}

func ConnectTestDB() {
	var err error
	Db, err = gorm.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}

	Db.AutoMigrate(&User{}, &Page{}, &ContentVersion{})
}

func Close() {
	Db.Close()
}
