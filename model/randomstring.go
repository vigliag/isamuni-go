package model

import (
	"encoding/base32"

	"github.com/gorilla/securecookie"
)

func GenRandomString(length int) string {
	return base32.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(length))
}
