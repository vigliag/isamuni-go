package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenUsage(t *testing.T) {
	m := Model{ConnectTestDB()}
	tval, err := m.CreateToken(1, 18)
	assert.NoError(t, err)
	assert.NotEmpty(t, tval)

	tok, err := m.GetToken(tval)
	assert.NoError(t, err)
	assert.Equal(t, tval, tok.Value)

	err = m.DeleteToken(tval)
	assert.NoError(t, err)

	_, err = m.GetToken(tval)
	assert.Error(t, err)
}
