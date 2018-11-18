package web

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/vigliag/isamuni-go/model"

	"github.com/spf13/viper"

	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
)

func facebookOauthConfig() *oauth2.Config {
	clientID := viper.GetString("FB_CLIENT_ID")
	clientSecret := viper.GetString("FB_CLIENT_SECRET")
	redirectURL := viper.GetString("APP_URL") + "/oauth/fb"

	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"email"},
		Endpoint:     facebook.Endpoint,
		RedirectURL:  redirectURL,
	}
}

func (ctl *Controller) redirectToFacebookLogin(c echo.Context) error {
	state := model.GenRandomString(18)

	sess, _ := session.Get("session", c)
	sess.Values["oauth_state"] = state
	sess.Save(c.Request(), c.Response())

	url := facebookOauthConfig().AuthCodeURL(state, oauth2.AccessTypeOnline)
	return c.Redirect(http.StatusSeeOther, url)
}

type facebookUserData struct {
	ID    string  `json:"id"`
	Email *string `json:"email"`
	Name  string  `json:"name"`
}

// User is redirected back to our site after accepting Facebook's permission dialog
func (ctl *Controller) completeFacebookLogin(c echo.Context) error {
	code := c.FormValue("code")
	state := c.FormValue("state")

	sess, _ := session.Get("session", c)
	savedState, ok := sess.Values["oauth_state"]
	if !ok {
		return echo.ErrUnauthorized
	}

	delete(sess.Values, "oauth_state")
	sess.Save(c.Request(), c.Response())

	if state != savedState {
		return echo.ErrUnauthorized
	}

	// At this point we have a valid response
	// and we can continue with handling the login
	ctx := context.Background()
	conf := facebookOauthConfig()

	// Obtain the token for the user
	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		return err
	}

	// At this point the user is successfully authenticated
	// We can fetch his profile
	client := conf.Client(ctx, tok)
	res, err := client.Get("https://graph.facebook.com/me?fields=id,name,email")
	if err != nil {
		return err
	}

	var fbuser facebookUserData

	jsondata, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsondata, &fbuser)
	if err != nil {
		return err
	}

	user, err := ctl.model.LoginOrCreateFB(currentUser(c), fbuser.ID, fbuser.Name, fbuser.Email)
	if err != nil {
		return err
	}

	sess.Values["userid"] = user.ID
	sess.Values["session_token"] = user.SessionToken
	sess.Save(c.Request(), c.Response())
	return c.Redirect(http.StatusSeeOther, "/")
}

func (ctl *Controller) loginPage(c echo.Context) error {
	redirParam := c.QueryParam("redir")
	c.Render(200, "login.html", H{"redir": redirParam})
	return nil
}

func (ctl *Controller) logout(c echo.Context) error {
	sess, _ := session.Get("session", c)
	delete(sess.Values, "userid")
	delete(sess.Values, "session_token")
	sess.Save(c.Request(), c.Response())
	return c.Redirect(http.StatusSeeOther, "/")
}

func (ctl *Controller) loginWithEmail(c echo.Context) error {
	tplName := "login.html"
	email := c.FormValue("email")
	password := c.FormValue("password")
	redirect := c.FormValue("redir")

	if redirect == "" {
		redirect = "/"
	}

	if email == "" || password == "" {
		c.Logger().Error("Empty email or password")
		return c.Render(http.StatusBadRequest, tplName, H{"error": "Empty email or password"})
	}

	user := ctl.model.LoginEmail(email, password)
	if user == nil {
		c.Logger().Error("Invalid email or password")
		return c.Render(http.StatusNotFound, tplName, H{"error": "Invalid email or password"})
	}

	sess, _ := session.Get("session", c)
	sess.Values["userid"] = user.ID
	sess.Values["session_token"] = user.SessionToken
	sess.Save(c.Request(), c.Response())

	return c.Redirect(http.StatusSeeOther, redirect)
}

func (ctl *Controller) setCurrentUserMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// If there is no session, or the user is not logged in
		// then continue without setting the current user
		sess, err := session.Get("session", c)
		if err != nil || sess.Values["userid"] == nil {
			return next(c)
		}

		// Retrieve user from the database and put it into the request context
		user := ctl.model.RetrieveUser(sess.Values["userid"].(uint))

		// If the session has been expired (for example as a result of a password change)
		// then log out the user
		if user.SessionToken != "" && user.SessionToken != sess.Values["session_token"] {
			return ctl.logout(c)
		}

		c.Set("currentUser", user)
		return next(c)
	}
}

func (ctl *Controller) setMailH(c echo.Context) error {
	// TODO no tests yet
	u := currentUser(c)
	if u == nil {
		return c.Redirect(http.StatusFound, "/login")
	}

	email := strings.TrimSpace(c.FormValue("email"))

	// do nothing if the mail has not been updated
	if email == "" || (u.Email != nil && *u.Email == email) {
		return c.Redirect(http.StatusSeeOther, "/me")
	}

	// update the email
	u.Email = &email
	u.EmailVerified = false
	err := ctl.model.Db.Save(u).Error
	if err != nil {
		//TODO better error handling
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/me")
}
