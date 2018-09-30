package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gosimple/slug"
	"github.com/microcosm-cc/bluemonday"
	blackfriday "gopkg.in/russross/blackfriday.v2"

	"github.com/vigliag/isamuni-go/db"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/middleware"
)

// Helpers
/////////////

type H map[string]interface{}

//Template is our template type, that exposes a custom Render method
type Template struct {
	templates *template.Template
}

//Render renders a template given its name, and the data to pass to the template engine
func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	if viewContext, isMap := data.(H); isMap {
		viewContext["currentUser"] = c.Get("currentUser")
	}
	return t.templates.ExecuteTemplate(w, name, data)
}

func RenderMarkdown(m string) template.HTML {
	unsafe := blackfriday.Run([]byte(m))
	html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)
	return template.HTML(html)
}

func getSessionKey() []byte {
	sessKey := os.Getenv("SESSION_KEY")
	if sessKey == "" {
		fmt.Println("SESSION_KEY not set, generating new session key")
		return securecookie.GenerateRandomKey(32)
	}
	return []byte(sessKey)
}

func serveTemplate(templateName string) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Render(200, templateName+".html", H{})
		return nil
	}
}

func PageUrl(p *db.Page) string {
	return fmt.Sprintf("/%s/%d", p.Type.CatName(), p.ID)
}

// Handlers
//////////////////

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

func indexPageHandler(ptype db.PageType) echo.HandlerFunc {
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

func newPageHandler(ptype db.PageType) echo.HandlerFunc {
	return func(c echo.Context) error {
		p := db.Page{}
		p.Type = ptype
		return c.Render(200, "pageEdit.html", H{"page": p})
	}
}

func showPageHandler(ptype db.PageType) echo.HandlerFunc {
	return func(c echo.Context) error {
		idParam := c.Param("id")
		if idParam == "" {
			log.Println("Invalid " + idParam)
			return echo.NewHTTPError(404, "Invalid ID")
		}

		id, err := strconv.Atoi(idParam)
		if err != nil {
			log.Println("Invalid " + idParam)
			return echo.NewHTTPError(404, "Invalid ID")
		}

		page := db.FindPage(uint(id), ptype)
		if page == nil {
			log.Println("Page not found for " + idParam)
			return echo.NewHTTPError(404, "Page not found")
		}

		return c.Render(200, "pageShow.html",
			H{"page": page, "pageURL": PageUrl(page), "content": RenderMarkdown(page.Content)})
	}
}

func editPageHandler(ptype db.PageType) echo.HandlerFunc {
	return func(c echo.Context) error {
		idParam := c.Param("id")
		if idParam == "" {
			log.Println("Invalid " + idParam)
			return echo.NewHTTPError(404, "Invalid ID")
		}

		id, err := strconv.Atoi(idParam)
		if err != nil {
			log.Println("Invalid " + idParam)
			return echo.NewHTTPError(404, "Invalid ID")
		}

		page := db.FindPage(uint(id), ptype)
		if page == nil {
			log.Println("Page not found for " + idParam)
			return echo.NewHTTPError(404, "Page not found")
		}

		return c.Render(200, "pageEdit.html", H{"page": page})
	}
}

func mePageHandler(c echo.Context) error {
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

func updatePageHandler(c echo.Context) error {
	u := currentUser(c)
	if u == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not logged in")
	}

	ptype, err := db.ParsePageType(c.FormValue("type"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid page type")
	}

	var p db.Page

	pid, _ := strconv.Atoi(c.FormValue("id"))

	if pid != 0 {
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

// Middlewares
//////////////////

func setCurrentUserMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get("session", c)
		if err == nil && sess.Values["userid"] != nil && sess.Values["email"] != nil {
			user := db.RetrieveUser(sess.Values["userid"].(uint), sess.Values["email"].(string))
			c.Set("currentUser", user)
		}
		return next(c)
	}
}

func currentUser(c echo.Context) *db.User {
	if u, ok := c.Get("currentUser").(*db.User); ok {
		return u
	}
	return nil
}

// Startup
///////////////

func createServer(r *echo.Echo) {
	t := &Template{
		templates: template.Must(template.ParseGlob("templates/*.html")),
	}
	r.Renderer = t

	//Initialize a CookieStore (inlined NewCookieStore method)
	//Cookies are signed but not encrypted
	//We rely on SameSite to prevent CSRF attacks
	cs := &sessions.CookieStore{
		Codecs: securecookie.CodecsFromPairs(getSessionKey()),
		Options: &sessions.Options{
			Path:     "/",
			MaxAge:   86400 * 30,
			SameSite: http.SameSiteLaxMode,
		},
	}
	cs.MaxAge(cs.Options.MaxAge)

	r.Use(session.Middleware(cs))
	r.Use(middleware.Logger())
	r.Use(setCurrentUserMiddleware)

	r.Static("/static", "static")
	r.GET("/", serveTemplate("home"))

	r.GET("/login", serveTemplate("login"))
	r.POST("/login", loginEmail)

	r.GET("/professionals/:id", showPageHandler(db.PageUser))
	r.GET("/wiki/:id", showPageHandler(db.PageWiki))
	r.GET("/companies/:id", showPageHandler(db.PageCompany))
	r.GET("/communities/:id", showPageHandler(db.PageCommunity))

	r.GET("/professionals/new", newPageHandler(db.PageUser))
	r.GET("/wiki/new", newPageHandler(db.PageWiki))
	r.GET("/companies/new", newPageHandler(db.PageCompany))
	r.GET("/communities/new", newPageHandler(db.PageCommunity))

	r.GET("/professionals/:id/edit", editPageHandler(db.PageUser))
	r.GET("/wiki/:id/edit", editPageHandler(db.PageWiki))
	r.GET("/companies/:id/edit", editPageHandler(db.PageCompany))
	r.GET("/communities/:id/edit", editPageHandler(db.PageCommunity))

	r.GET("/professionals", indexPageHandler(db.PageUser))
	r.GET("/wiki", indexPageHandler(db.PageWiki))
	r.GET("/companies", indexPageHandler(db.PageCompany))
	r.GET("/communities", indexPageHandler(db.PageCommunity))

	r.GET("/me", mePageHandler)

	r.POST("/pages", updatePageHandler)
	return
}

func main() {
	db.Connect()

	r := echo.New()
	createServer(r)

	r.Use(middleware.Recover())
	db.RegisterEmail("vigliag", "vigliag@gmail.com", "password")

	http.ListenAndServe(":8080", r)
	fmt.Println("Server started")
}
