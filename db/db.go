package db

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	_ "github.com/mattn/go-sqlite3"
)

var (
	Db *gorm.DB
)

func Connect() *gorm.DB {
	var err error
	Db, err = gorm.Open("sqlite3", "database.db")
	if err != nil {
		panic(err)
	}

	Db.AutoMigrate(&User{}, &Page{})

	return Db
}

func ConnectTestDB() {
	var err error
	Db, err = gorm.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}

	Db.AutoMigrate(&User{}, &Page{})
}
