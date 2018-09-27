package db

import (
	"database/sql/driver"

	"github.com/jinzhu/gorm"
)

//PageType tells if a Page is for a User, a Company or a Community
type PageType int64

//Scan reads a PageType from the database
func (u *PageType) Scan(value interface{}) error { *u = PageType(value.(int64)); return nil }

//Value serializes a PageType from the database
func (u PageType) Value() (driver.Value, error) { return int64(u), nil }

const (
	PageUser      PageType = 0
	PageCompany            = iota
	PageCommunity          = iota
	PageWiki               = iota
)

//Page represents a page in the database
type Page struct {
	gorm.Model
	Title   string   `gorm:"not null" binding:"required"`
	Slug    string   `gorm:"unique"`
	Content string   `binding:"required"`
	Type    PageType `gorm:"type:int" binding:"required"`
}

func FindPage(id uint, ptype PageType) *Page {
	var page Page
	res := Db.First(&page, " id = ? and type = ? ", id, ptype)
	if res.Error != nil {
		return nil
	}
	return &page
}
