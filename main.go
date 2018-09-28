package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"vigliag/commwiki/db"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/middleware"
)

type H map[string]interface{}

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
		c.Render(200, templateName+".html", nil)
		return nil
	}
}

func loginEmail(c echo.Context) error {
	tplName := "loginPage.html"
	email := c.FormValue("email")
	password := c.FormValue("password")

	if email == "" || password == "" {
		log.Println("Empty email or password")
		c.Render(http.StatusBadRequest, tplName, H{"error": "Empty email or password"})
		return nil
	}

	user := db.LoginEmail(email, password)
	if user == nil {
		log.Println("Invalid email or password")
		c.Render(http.StatusNotFound, tplName, H{"error": "Invalid email or password"})
		return nil
	}

	sess, _ := session.Get("session", c)
	sess.Values["email"] = email
	sess.Values["userid"] = user.ID
	sess.Save(c.Request(), c.Response())

	c.String(200, fmt.Sprintf("Logged in with mail %s", user.Email))
	return nil
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

		c.Render(200, "showPage.html", H{"page": page})
		return nil
	}
}

func currentUser(c echo.Context) *db.User {
	sess, err := session.Get("session", c)
	if err != nil {
		return nil
	}
	return db.RetrieveUser(sess.Values["userid"].(uint), sess.Values["email"].(string))
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

	pid, _ := strconv.Atoi(c.FormValue("id"))

	var p db.Page
	p.Title = c.FormValue("title")
	p.Slug = c.FormValue("slug")
	p.Content = c.FormValue("content")
	p.Type = ptype
	p.ID = uint(pid)

	res := db.Db.Save(&p)
	if res.Error != nil {
		log.Println(res.Error)
		return echo.NewHTTPError(http.StatusBadRequest, "Could not update page")
	}

	return c.Redirect(http.StatusTemporaryRedirect, PageUrl(p))
}

func PageUrl(p db.Page) string {
	return fmt.Sprintf("/%s/%d", p.Type.CatName(), p.ID)
}

//Template is our template type, that exposes a custom Render method
type Template struct {
	templates *template.Template
}

//Render renders a template given its name, and the data to pass to the template engine
func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

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

	r.GET("/login", serveTemplate("loginPage"))
	r.POST("/login", loginEmail)

	r.GET("/professionals/:id", showPageHandler(db.PageUser))
	r.GET("/wiki/:id", showPageHandler(db.PageWiki))
	r.GET("/companies/:id", showPageHandler(db.PageCompany))
	r.GET("/communities/:id", showPageHandler(db.PageCommunity))

	r.POST("/professionals", updatePageHandler)
	r.POST("/wiki", updatePageHandler)
	r.POST("/companies", updatePageHandler)
	r.POST("/communities", updatePageHandler)
	return
}

func main() {
	db.Connect()

	r := echo.New()
	r.Use(middleware.Logger())
	r.Use(middleware.Recover())
	createServer(r)

	http.ListenAndServe(":8080", r)
	fmt.Println("Server started")
}
