package model

import (
	"encoding/base64"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/jinzhu/gorm"
)

//Token is a token, used in oauth, or for email confirmation
type Token struct {
	gorm.Model

	// Specifies what this token is about (eg. a userid)
	Identifier uint

	// After this time, the token is no more valid
	Expiration time.Time

	// Autogenerated random value
	Value string
}

func DeleteExpiredTokens() error {
	return Db.Where("expiration < ?", time.Now()).Delete(Token{}).Error
}

func CreateToken(identifier uint) (string, error) {
	state := base64.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(16))
	t := Token{
		Expiration: time.Now().Add(time.Minute * 10),
		Value:      state,
		Identifier: identifier,
	}
	return state, Db.Save(&t).Error
}

func GetToken(value string) (*Token, error) {
	err := DeleteExpiredTokens()
	if err != nil {
		return nil, err
	}

	t := new(Token)

	res := Db.First(&t, "state = ?", value)
	if res.Error != nil {
		return nil, res.Error
	}

	return t, nil
}

func DeleteToken(value string) error {
	return Db.Where("value =  ?", value).Delete(Token{}).Error
}
