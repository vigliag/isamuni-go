package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strconv"

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

//H is the context for a template
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

//RenderMarkdown renders markdown to safe HTML for use in a template
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

func intParameter(c echo.Context, param string) int {
	id, err := strconv.Atoi(c.Param(param))
	if err != nil {
		return 0
	}
	return id
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

func createServer(r *echo.Echo) *echo.Echo {
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

	r.GET("/professionals/:id", showPageH(db.PageUser))
	r.GET("/wiki/:id", showPageH(db.PageWiki))
	r.GET("/companies/:id", showPageH(db.PageCompany))
	r.GET("/communities/:id", showPageH(db.PageCommunity))

	r.GET("/professionals/new", newPageH(db.PageUser))
	r.GET("/wiki/new", newPageH(db.PageWiki))
	r.GET("/companies/new", newPageH(db.PageCompany))
	r.GET("/communities/new", newPageH(db.PageCommunity))

	r.GET("/professionals/:id/edit", editPageH(db.PageUser))
	r.GET("/wiki/:id/edit", editPageH(db.PageWiki))
	r.GET("/companies/:id/edit", editPageH(db.PageCompany))
	r.GET("/communities/:id/edit", editPageH(db.PageCommunity))

	r.GET("/professionals", indexPageH(db.PageUser))
	r.GET("/wiki", indexPageH(db.PageWiki))
	r.GET("/companies", indexPageH(db.PageCompany))
	r.GET("/communities", indexPageH(db.PageCommunity))

	r.GET("/me", mePageH)

	r.POST("/pages", updatePageH)
	return r
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
