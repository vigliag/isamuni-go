package db

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/scrypt"
)

var (
	Db *gorm.DB
)

type User struct {
	gorm.Model
	Username       string `gorm:"unique"`
	HashedPassword string `gorm:"not null"`
	Salt           string `gorm:"not null"`
	Email          string `gorm:"unique"`
}

func UserPage(u *User) *Page {
	var page Page
	res := Db.Find(&page, "owner_id = ? and type = ?", u.ID, PageUser)
	if res.Error != nil {
		return nil
	}
	return &page
}

func RetrieveUser(id uint, email string) *User {
	var u User
	res := Db.First(&u, "email = ? and id = ?", email, id)
	if res.Error != nil {
		return nil
	}
	return &u
}

func LoginEmail(email string, password string) *User {
	var u User
	res := Db.First(&u, "email = ?", email)
	if res.Error != nil {
		return nil
	}

	if u.CheckPassword(password) == false {
		return nil
	}

	return &u
}

func RegisterEmail(username string, email string, password string) (*User, error) {
	u := User{
		Username: username,
		Email:    email,
	}
	u.SetPassword(password)
	res := Db.Save(&u)
	if res.Error != nil {
		return nil, res.Error
	}
	return &u, nil
}

func (u *User) SetPassword(password string) {
	salt := make([]byte, 8)
	_, err := rand.Read(salt)
	if err != nil {
		panic(err)
	}
	u.Salt = string(salt)
	u.HashedPassword = HashPassword(password, salt)
}

func (u *User) CheckPassword(password string) bool {
	if password == "" {
		return false
	}
	return HashPassword(password, []byte(u.Salt)) == u.HashedPassword
}

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

func HashPassword(password string, salt []byte) string {
	dk, err := scrypt.Key([]byte(password), salt, 32768, 8, 1, 32)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(dk)
}
