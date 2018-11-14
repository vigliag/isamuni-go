package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindPage(t *testing.T) {
	m := New(ConnectTestDB())

	p := Page{
		Content: "Ciao",
		Type:    PageCompany,
		Title:   "Example company",
	}

	res := m.Db.Save(&p)
	assert.Nil(t, res.Error)
	assert.NotZero(t, p.ID)

	rpage := m.FindPage(p.ID, p.Type)
	assert.NotNil(t, rpage)
	assert.Equal(t, p.Title, rpage.Title)

	p2 := m.FindPage(12515, PageCompany)
	assert.Nil(t, p2)

	m.Db.Close()
}
