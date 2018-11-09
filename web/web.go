package web

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"

	rice "github.com/GeertJohan/go.rice"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/middleware"
	"github.com/microcosm-cc/bluemonday"
	"github.com/vigliag/isamuni-go/model"
	blackfriday "gopkg.in/russross/blackfriday.v2"
)

// Helpers
/////////////

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

func loginPage(c echo.Context) error {
	redirParam := c.QueryParam("redir")
	c.Render(200, "login.html", H{"redir": redirParam})
	return nil
}

func serveTemplate(templateName string) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Render(200, templateName+".html", H{})
		return nil
	}
}

func PageURL(p *model.Page) string {
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
			user := model.RetrieveUser(sess.Values["userid"].(uint), sess.Values["email"].(string))
			c.Set("currentUser", user)
		}
		return next(c)
	}
}

func currentUser(c echo.Context) *model.User {
	if u, ok := c.Get("currentUser").(*model.User); ok {
		return u
	}
	return nil
}

// Startup
///////////////

func CreateServer(r *echo.Echo) *echo.Echo {
	t := loadTemplates()
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

	r.Pre(middleware.RemoveTrailingSlash())

	r.Use(session.Middleware(cs))
	r.Use(middleware.Logger())
	r.Use(setCurrentUserMiddleware)

	staticBox := rice.MustFindBox("static")
	staticFileServer := http.StripPrefix("/static/", http.FileServer(staticBox.HTTPBox()))
	r.GET("/static/*", echo.WrapHandler(staticFileServer))

	r.GET("/", serveTemplate("home"))

	r.GET("/login", loginPage)
	r.GET("/logout", loginPage)
	r.POST("/login", loginWithEmail)
	r.POST("/logout", logout)

	r.GET("/professionals/:id", showPageH(model.PageUser))
	r.GET("/wiki/:id", showPageH(model.PageWiki))
	r.GET("/companies/:id", showPageH(model.PageCompany))
	r.GET("/communities/:id", showPageH(model.PageCommunity))

	r.GET("/professionals/new", newPageH(model.PageUser))
	r.GET("/wiki/new", newPageH(model.PageWiki))
	r.GET("/companies/new", newPageH(model.PageCompany))
	r.GET("/communities/new", newPageH(model.PageCommunity))

	r.GET("/professionals/:id/edit", editPageH(model.PageUser))
	r.GET("/wiki/:id/edit", editPageH(model.PageWiki))
	r.GET("/companies/:id/edit", editPageH(model.PageCompany))
	r.GET("/communities/:id/edit", editPageH(model.PageCommunity))

	r.GET("/professionals", indexPageH(model.PageUser))
	r.GET("/wiki", indexPageH(model.PageWiki))
	r.GET("/companies", indexPageH(model.PageCompany))
	r.GET("/communities", indexPageH(model.PageCommunity))

	r.GET("/me", mePageH)

	r.POST("/pages", updatePageH)
	r.POST("/pages/:id", updatePageH)
	return r
}
