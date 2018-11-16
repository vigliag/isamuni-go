package model

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite" //register gorm sqlite dialect
	_ "github.com/mattn/go-sqlite3"
)

func Connect(dbPath string) *gorm.DB {
	var err error
	Db, err := gorm.Open("sqlite3", dbPath)
	if err != nil {
		panic(err)
	}

	Db.AutoMigrate(&User{}, &Page{}, &ContentVersion{}, &Token{})

	return Db
}

func ConnectTestDB() *gorm.DB {
	var err error
	Db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}

	Db.AutoMigrate(&User{}, &Page{}, &ContentVersion{}, &Token{})
	return Db
}
