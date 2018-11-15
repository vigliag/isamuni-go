package model

import (
	"encoding/base64"

	"github.com/gorilla/securecookie"
)

func GenRandomString() string {
	return base64.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(16))
}
