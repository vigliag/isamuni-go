package web

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gosimple/slug"
	"github.com/labstack/echo"
	"github.com/vigliag/isamuni-go/model"
)

// PageURL returns the url for a give page
func PageURL(p *model.Page) string {
	typeStr := CatUrl(p.Type)
	return fmt.Sprintf("/%s/%d", typeStr, p.ID)
}

func CatUrl(ptype model.PageType) string {
	typeStr := "page"
	switch ptype {
	case model.PageUser:
		typeStr = "professionals"
	case model.PageCommunity:
		typeStr = "communities"
	case model.PageCompany:
		typeStr = "companies"
	case model.PageWiki:
		typeStr = "wiki"
	}
	return typeStr
}

func CatName(ptype model.PageType) string {
	name := "page"
	switch ptype {
	case model.PageUser:
		name = "professionisti"
	case model.PageCommunity:
		name = "community"
	case model.PageCompany:
		name = "aziende"
	case model.PageWiki:
		name = "wiki"
	}
	return name
}

func (ctl *Controller) indexPageH(ptype model.PageType) echo.HandlerFunc {
	return func(c echo.Context) error {
		pages, err := ctl.model.GetPagesOfType(ptype)
		if err != nil {
			return err
		}
		title := strings.Title(CatName(ptype))
		return c.Render(200, "pageIndex.html", H{"pages": pages, "title": title, "cat": ptype})
	}
}

func (ctl *Controller) newPageH(ptype model.PageType) echo.HandlerFunc {
	return func(c echo.Context) error {
		u := currentUser(c)
		if u == nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Not logged in")
		}

		p := model.Page{}
		p.Type = ptype
		return c.Render(200, "pageEdit.html", H{"page": p})
	}
}

func (ctl *Controller) showPageH(ptype model.PageType) echo.HandlerFunc {
	return func(c echo.Context) error {
		u := currentUser(c)
		id := intParameter(c, "id")
		if id == 0 {
			return echo.NewHTTPError(404, "Invalid ID")
		}

		page := ctl.model.FindPage(uint(id), ptype)
		if page == nil {
			log.Printf("Page not found for %v\n", id)
			return echo.NewHTTPError(404, "Page not found")
		}

		return c.Render(200, "pageShow.html",
			H{"page": page, "pageURL": PageURL(page),
				"content": RenderMarkdown(page.Content),
				"canEdit": ctl.model.CanEdit(page, u),
			})
	}
}

func (ctl *Controller) mePageH(c echo.Context) error {
	u := currentUser(c)
	if u == nil {
		return c.Redirect(http.StatusFound, "/login")
	}

	page := ctl.model.UserPage(u)
	if page == nil {
		fmt.Printf("User page not found")
		page = &model.Page{
			Title: u.Username,
		}
	}

	// generate an url to edit the page, if the page does not exists
	action := "/pages"
	if page.ID != 0 {
		action = fmt.Sprintf("/pages/%d", page.ID)
	}
	shownContent := page.Content
	shownVersion := "Current"
	return c.Render(200, "profileEdit.html", H{"page": page, "action": action, "shownContent": shownContent, "shownVersion": shownVersion, "user": u})
}

// shows edit form for a page
// if user is admin, it should also show a list of versions, containing:
// - the current version
// - all versions greater than current
// and the form should be populated with the contents of the latest version
// the version shown in the editor should be clearly indicated
func (ctl *Controller) editPageH(ptype model.PageType) echo.HandlerFunc {
	return func(c echo.Context) error {
		u := currentUser(c)
		if u == nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Not logged in")
		}

		id := intParameter(c, "id")
		if id == 0 {
			return echo.NewHTTPError(404, "Invalid ID")
		}

		page := ctl.model.FindPage(uint(id), ptype)
		if page == nil {
			log.Printf("Page not found for %v\n", id)
			return echo.NewHTTPError(404, "Page not found")
		}
		if !ctl.model.CanEdit(page, u) {
			return echo.NewHTTPError(http.StatusUnauthorized, "Can't edit this page")
		}

		versions, err := ctl.model.FindNewerPageVersions(page)
		if err != nil {
			log.Println(err)
		}

		shownContent := page.Content
		shownVersion := "Current"
		if len(versions) > 0 {
			shownContent = versions[0].Content
			shownVersion = fmt.Sprintf("revision %v by %v at %v", versions[0].ID, versions[0].User.Username, versions[0].UpdatedAt)
		}

		action := "/pages"
		if page.ID != 0 {
			action = fmt.Sprintf("/pages/%d", page.ID)
		}
		return c.Render(200, "pageEdit.html",
			H{"page": page, "versions": versions,
				"shownContent": shownContent, "shownVersion": shownVersion, "action": action})
	}
}

// receives POST.
// Id parameter is the page to edit
//
// if Page has an owner, only the owner can edit it
// if Page has no owner, on save a new ContentVersion is created and,
//    if user is admin, the version is also saved in the Page model
//    if user is not admin, a notification is generated
func (ctl *Controller) updatePageH(c echo.Context) error {
	u := currentUser(c)
	if u == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not logged in")
	}

	iptype, err := strconv.Atoi(c.FormValue("type"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid page type")
	}
	ptype := model.PageType(iptype)

	var p model.Page

	pid := intParameter(c, "id")

	if pid != 0 {
		//Trying to update an existing page
		res := ctl.model.Db.Find(&p, pid)

		if res.Error != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Editing an invalid page")
		}

		if !ctl.model.CanEdit(&p, u) {
			return echo.NewHTTPError(http.StatusBadRequest, "Only the owner of this page can edit it")
		}
	} else {
		// New page

		if ptype == model.PageUser {
			if ctl.model.UserPage(u) != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "This user has a page already")
			}
			p.OwnerID = u.ID
		}
	}

	// Assign values from the request
	p.Title = c.FormValue("title")
	if s := c.FormValue("slug"); s != "" {
		p.Slug = s
	} else {
		p.Slug = slug.Make(p.Title)
	}
	p.Content = c.FormValue("content")
	p.Type = ptype
	p.ID = uint(pid)

	// Save the page
	err = ctl.model.SavePage(&p, u)
	if err != nil {
		log.Println(err)
		return c.Render(http.StatusBadRequest, "pageEdit.html", H{"page": p, "error": "Could not save page"})
	}

	// Index the page
	err = ctl.index.IndexPage(&p)
	if err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, PageURL(&p))
}
