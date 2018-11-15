package web

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/vigliag/isamuni-go/index"

	"github.com/spf13/viper"

	rice "github.com/GeertJohan/go.rice"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/middleware"
	"github.com/vigliag/isamuni-go/model"
)

type Controller struct {
	model *model.Model
	index *index.Index
}

func NewController(model *model.Model, index *index.Index) *Controller {
	return &Controller{model, index}
}

// Helpers
/////////////

func intParameter(c echo.Context, param string) int {
	id, err := strconv.Atoi(c.Param(param))
	if err != nil {
		return 0
	}
	return id
}

func currentUser(c echo.Context) *model.User {
	if u, ok := c.Get("currentUser").(*model.User); ok {
		return u
	}
	return nil
}

func serveTemplate(templateName string) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Render(200, templateName+".html", H{})
		return nil
	}
}

func customHTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
	}
	if err := c.Render(code, "error.html", H{"code": code}); err != nil {
		c.HTML(http.StatusInternalServerError, "<h3>Internal Server Error</h3>")
		c.Logger().Error(err)
	}
	c.Logger().Error(err)
}

// CreateServer attaches the app's routes and middlewares to an Echo server
func CreateServer(r *echo.Echo, ctl *Controller) *echo.Echo {
	t := loadTemplates()
	r.Renderer = t
	r.HTTPErrorHandler = customHTTPErrorHandler

	// Obtain a session key to use for signing cookies
	// If no session_key was set, generate one (useful in development)
	sessKey := []byte(viper.GetString("SESSION_KEY"))
	if len(sessKey) == 0 {
		fmt.Println("SESSION_KEY not set, generating new session key")
		sessKey = securecookie.GenerateRandomKey(32)
	}

	//Initialize a CookieStore to save sessions(inlined NewCookieStore method)
	//Contents of the sessions will be signed but not encrypted cookies
	//We set SameSite as an additional measure to prevent CSRF attacks
	cs := &sessions.CookieStore{
		Codecs: securecookie.CodecsFromPairs(sessKey),
		Options: &sessions.Options{
			Path:     "/",
			MaxAge:   86400 * 30, //30 days
			SameSite: http.SameSiteLaxMode,
		},
	}
	cs.MaxAge(cs.Options.MaxAge)

	r.Pre(middleware.RemoveTrailingSlash())
	r.Use(session.Middleware(cs))
	r.Use(middleware.Logger())
	r.Use(ctl.setCurrentUserMiddleware)

	staticBox := rice.MustFindBox("static")
	staticFileServer := http.StripPrefix("/static/", http.FileServer(staticBox.HTTPBox()))
	r.GET("/static/*", echo.WrapHandler(staticFileServer))

	r.GET("/", ctl.homeH)

	r.GET("/login", ctl.loginPage)
	r.GET("/logout", ctl.loginPage)
	r.POST("/login", ctl.loginWithEmail)
	r.POST("/logout", ctl.logout)

	r.GET("/login/facebook", ctl.redirectToFacebookLogin)
	r.GET("/oauth/fb", ctl.completeFacebookLogin)

	r.GET("/professionals/:id", ctl.showPageH(model.PageUser))
	r.GET("/wiki/:id", ctl.showPageH(model.PageWiki))
	r.GET("/companies/:id", ctl.showPageH(model.PageCompany))
	r.GET("/communities/:id", ctl.showPageH(model.PageCommunity))

	r.GET("/professionals/new", ctl.newPageH(model.PageUser))
	r.GET("/wiki/new", ctl.newPageH(model.PageWiki))
	r.GET("/companies/new", ctl.newPageH(model.PageCompany))
	r.GET("/communities/new", ctl.newPageH(model.PageCommunity))

	r.GET("/professionals/:id/edit", ctl.editPageH(model.PageUser))
	r.GET("/wiki/:id/edit", ctl.editPageH(model.PageWiki))
	r.GET("/companies/:id/edit", ctl.editPageH(model.PageCompany))
	r.GET("/communities/:id/edit", ctl.editPageH(model.PageCommunity))

	r.GET("/professionals", ctl.indexPageH(model.PageUser))
	r.GET("/wiki", ctl.indexPageH(model.PageWiki))
	r.GET("/companies", ctl.indexPageH(model.PageCompany))
	r.GET("/communities", ctl.indexPageH(model.PageCommunity))

	r.GET("/me", ctl.mePageH)

	r.GET("/search", ctl.searchH)
	r.GET("/privacy", serveTemplate("privacy"))

	r.POST("/pages", ctl.updatePageH)
	r.POST("/pages/:id", ctl.updatePageH)
	return r
}
