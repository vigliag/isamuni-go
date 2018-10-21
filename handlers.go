package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gosimple/slug"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"github.com/vigliag/isamuni-go/db"
)

func logout(c echo.Context) error {
	sess, _ := session.Get("session", c)
	sess.Values["email"] = ""
	sess.Values["userid"] = uint(0)
	sess.Save(c.Request(), c.Response())
	return c.Redirect(http.StatusSeeOther, "/")
}

func loginWithEmail(c echo.Context) error {
	tplName := "login.html"
	email := c.FormValue("email")
	password := c.FormValue("password")
	redirect := c.FormValue("redir")

	if redirect == "" {
		redirect = "/"
	}

	if email == "" || password == "" {
		log.Println("Empty email or password")
		return c.Render(http.StatusBadRequest, tplName, H{"error": "Empty email or password"})
	}

	user := db.LoginEmail(email, password)
	if user == nil {
		log.Println("Invalid email or password")
		return c.Render(http.StatusNotFound, tplName, H{"error": "Invalid email or password"})
	}

	sess, _ := session.Get("session", c)
	sess.Values["email"] = email
	sess.Values["userid"] = user.ID
	sess.Save(c.Request(), c.Response())

	return c.Redirect(http.StatusSeeOther, redirect)
}

func indexPageH(ptype db.PageType) echo.HandlerFunc {
	return func(c echo.Context) error {
		var pages []db.Page
		res := db.Db.Order("title").Find(&pages, "type = ?", ptype)
		if err := res.Error; err != nil {
			return err
		}
		cat := ptype.CatName()
		title := strings.Title(cat)
		return c.Render(200, "pageIndex.html", H{"pages": pages, "title": title, "cat": cat})
	}
}

func newPageH(ptype db.PageType) echo.HandlerFunc {
	return func(c echo.Context) error {
		u := currentUser(c)
		if u == nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Not logged in")
		}

		p := db.Page{}
		p.Type = ptype
		return c.Render(200, "pageEdit.html", H{"page": p})
	}
}

func showPageH(ptype db.PageType) echo.HandlerFunc {
	return func(c echo.Context) error {
		u := currentUser(c)
		id := intParameter(c, "id")
		if id == 0 {
			return echo.NewHTTPError(404, "Invalid ID")
		}

		page := db.FindPage(uint(id), ptype)
		if page == nil {
			log.Printf("Page not found for %v\n", id)
			return echo.NewHTTPError(404, "Page not found")
		}

		return c.Render(200, "pageShow.html",
			H{"page": page, "pageURL": PageURL(page),
				"content": RenderMarkdown(page.Content),
				"canEdit": db.CanEdit(page, u),
			})
	}
}

func mePageH(c echo.Context) error {
	u := currentUser(c)
	if u == nil {
		return c.Redirect(http.StatusFound, "/login")
	}

	page := db.UserPage(u)
	if page == nil {
		page = &db.Page{
			Title: u.Username,
		}
	}

	action := "/pages"
	if page.ID != 0 {
		fmt.Sprintf("/pages/%d", page.ID)
	}
	return c.Render(200, "pageEdit.html", H{"page": page, "action": action})
}

// shows edit form for a page
// if user is admin, it should also show a list of versions, containing:
// - the current version
// - all versions greater than current
// and the form should be populated with the contents of the latest version
// the version shown in the editor should be clearly indicated
func editPageH(ptype db.PageType) echo.HandlerFunc {
	return func(c echo.Context) error {
		u := currentUser(c)
		if u == nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Not logged in")
		}

		id := intParameter(c, "id")
		if id == 0 {
			return echo.NewHTTPError(404, "Invalid ID")
		}

		page := db.FindPage(uint(id), ptype)
		if page == nil {
			log.Printf("Page not found for %v\n", id)
			return echo.NewHTTPError(404, "Page not found")
		}
		if !db.CanEdit(page, u) {
			return echo.NewHTTPError(http.StatusUnauthorized, "Can't edit this page")
		}

		versions, err := db.FindNewerPageVersions(page)
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
func updatePageH(c echo.Context) error {
	u := currentUser(c)
	if u == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not logged in")
	}

	iptype, err := strconv.Atoi(c.FormValue("type"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid page type")
	}
	ptype := db.PageType(iptype)

	var p db.Page

	pid := intParameter(c, "id")

	if pid != 0 {
		//Trying to update an existing page
		res := db.Db.Find(&p, pid)

		if res.Error != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Editing an invalid page")
		}

		if !db.CanEdit(&p, u) {
			return echo.NewHTTPError(http.StatusBadRequest, "Only the owner of this page can edit it")
		}
	} else {
		// New page

		if ptype == db.PageUser {
			if db.UserPage(u) != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "This user has a page already")
			} else {
				p.OwnerID = u.ID
			}
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
	err = db.SavePage(&p, u)
	if err != nil {
		log.Println(err)
		return c.Render(http.StatusBadRequest, "pageEdit.html", H{"page": p, "error": "Could not save page"})
	}

	return c.Redirect(http.StatusSeeOther, PageURL(&p))
}
