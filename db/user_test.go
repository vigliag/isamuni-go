package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUser_CheckPassword(t *testing.T) {
	u := User{}
	u.SetPassword("password")

	assert.True(t, u.CheckPassword("password"))
	assert.False(t, u.CheckPassword("wrong"))

	u.SetPassword("")
	assert.False(t, u.CheckPassword(""))
	assert.False(t, u.CheckPassword("wrong"))
}

func TestDB_Login(t *testing.T) {
	ConnectTestDB()

	u1, err := RegisterEmail("vigliag", "vigliag@gmail.com", "password")
	assert.NoError(t, err)

	u2 := LoginEmail("vigliag@gmail.com", "password")
	assert.Equal(t, u1.ID, u2.ID)

	u3 := LoginEmail("nonexistent", "password")
	assert.Nil(t, u3)

	u4 := LoginEmail("vigliag@gmail.com", "wrongpassword")
	assert.Nil(t, u4)

	u5, err := RegisterEmail("vigliag", "vigliag@gmail.com", "password")
	assert.Error(t, err)
	assert.Nil(t, u5)
}
