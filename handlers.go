package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gosimple/slug"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"github.com/vigliag/isamuni-go/db"
)

func loginEmail(c echo.Context) error {
	tplName := "login.html"
	email := c.FormValue("email")
	password := c.FormValue("password")

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

	return c.Redirect(http.StatusSeeOther, "/")
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
		p := db.Page{}
		p.Type = ptype
		return c.Render(200, "pageEdit.html", H{"page": p})
	}
}

func showPageH(ptype db.PageType) echo.HandlerFunc {
	return func(c echo.Context) error {
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
			H{"page": page, "pageURL": PageUrl(page), "content": RenderMarkdown(page.Content)})
	}
}

func editPageH(ptype db.PageType) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := intParameter(c, "id")
		if id == 0 {
			return echo.NewHTTPError(404, "Invalid ID")
		}

		page := db.FindPage(uint(id), ptype)
		if page == nil {
			log.Printf("Page not found for %v\n", id)
			return echo.NewHTTPError(404, "Page not found")
		}

		return c.Render(200, "pageEdit.html", H{"page": page})
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
	return c.Render(200, "pageEdit.html", H{"page": page})
}

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

	pid, _ := strconv.Atoi(c.FormValue("id"))

	if pid != 0 {
		//Page exists

		res := db.Db.Find(&p, pid)
		if res.Error != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Editing an invalid page")
		}

		if p.Type != db.PageWiki && p.OwnerID != u.ID {
			return echo.NewHTTPError(http.StatusUnauthorized, "Can't edit this page")
		}

		if p.Type == db.PageUser && db.UserPage(u) != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "User has a page already")
		}

		p.OwnerID = u.ID
	}

	p.Title = c.FormValue("title")
	if s := c.FormValue("slug"); s != "" {
		p.Slug = s
	} else {
		p.Slug = slug.Make(p.Title)
	}

	p.Content = c.FormValue("content")
	p.Type = ptype
	p.ID = uint(pid)
	p.Short = c.FormValue("short")

	res := db.Db.Save(&p)
	if res.Error != nil {
		log.Println(res.Error)
		return c.Render(http.StatusBadRequest, "pageEdit.html", H{"page": p, "error": "Could not save page"})
	}

	return c.Redirect(http.StatusSeeOther, PageUrl(&p))
}
