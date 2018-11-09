package model

import (
	"database/sql/driver"

	"github.com/gosimple/slug"

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

//Page represents either:
// a User's profile
// a Company
// a Community
// a Wiki page
type Page struct {
	gorm.Model
	Title string `gorm:"not null"`
	Slug  string `gorm:"unique"`
	Short string

	Type PageType `gorm:"type:int"`

	// If not null, the owner is the only one who can edit the page
	OwnerID uint
	Owner   User

	// Exact location
	Location string

	// City, used for filtering
	City string

	Sector  string
	Website string

	// Unparsed contents of the page
	Content string

	// Parsed metadata
	Parsed string

	ContentVersion []ContentVersion

	// If there is no approved version, then the page should not be publicly listed
	// It doesn't need to be a foreign key, it is only useful to find if there are
	// unapproved versions of the page
	ApprovedVersionID uint
}

func (p *Page) assignDataItem(name, content string) {

}

func (p *Page) SetFieldsToParsedContent() {
	parsed := normalizeHeaders(parseContent(p.Content, "dati"))

	p.Short = parsed["short"]
	p.City = parsed["city"]
	p.Website = parsed["website"]
}

type ContentVersion struct {
	gorm.Model

	PageID uint
	Page   Page

	UserID uint
	User   User

	Content string
}

func FindNewerPageVersions(page *Page) ([]ContentVersion, error) {
	var versions []ContentVersion
	res := Db.
		Preload("User").
		Order("id desc").
		Find(&versions, "page_id = ? and id > ?", page.ID, page.ApprovedVersionID)
	return versions, res.Error
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

func CanApproveEdits(p *Page, u *User) bool {
	return u != nil && (u.Role == "admin" || u.ID == p.OwnerID)
}

func CanEdit(p *Page, u *User) bool {
	return u != nil && (u.Role == "admin" || p.OwnerID == 0 || u.ID == p.OwnerID)
}

func SavePage(p *Page, u *User) error {
	if p.Slug == "" {
		p.Slug = slug.Make(p.Title)
	}

	// If the page is new, we save it to the database first
	// It will have no ApprovedVersionID at this stage, and will be unlisted
	if p.ID == 0 {
		if err := Db.Save(p).Error; err != nil {
			return err
		}
	}

	// Create and save a new ContentVersion for the page
	cv := ContentVersion{
		Content: p.Content,
		PageID:  p.ID,
		UserID:  u.ID,
	}
	if err := Db.Save(&cv).Error; err != nil {
		return err
	}

	// If an admin or owner is submitting this version, approve it
	// Assign the contentVersion to the page and parse the page's contents
	if CanApproveEdits(p, u) {
		p.Content = cv.Content
		p.ApprovedVersionID = cv.ID
		p.SetFieldsToParsedContent()
		return Db.Save(p).Error
	}

	return nil
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

func AllPages() ([]Page, error) {
	var pages []Page
	return pages, Db.Find(&pages).Error
}
