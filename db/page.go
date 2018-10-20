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

//Kind or namespace of a page
const (
	PageUser      PageType = 0
	PageCompany            = iota
	PageCommunity          = iota
	PageWiki               = iota
)

//Page represents a page in the database
type Page struct {
	gorm.Model
	Title    string `gorm:"not null"`
	Slug     string `gorm:"unique"`
	Short    string
	Content  string
	Type     PageType `gorm:"type:int" form:"type"`
	OwnerID  uint
	Owner    User
	Location string
	Area     string
	Sector   string
	Website  string
}

//FindPage returns a page of a given type by ID or null
func FindPage(id uint, ptype PageType) *Page {
	var page Page
	res := Db.First(&page, " id = ? and type = ? ", id, ptype)
	if res.Error != nil {
		return nil
	}
	return &page
}

func (p PageType) Int() int {
	return int(p)
}

func (p PageType) CatName() string {
	switch p {
	case PageUser:
		return "professionals"
	case PageCommunity:
		return "communities"
	case PageCompany:
		return "companies"
	case PageWiki:
		return "wiki"
	}
	return ""
}
