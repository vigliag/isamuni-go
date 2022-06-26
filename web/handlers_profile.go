package web

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/vigliag/isamuni-go/model"
)

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
	var shownContent string
	action := "/pages"
	if page.ID != 0 {
		action = fmt.Sprintf("/pages/%d", page.ID)
		shownContent = page.Content
	} else {
		// if the page is new, display example content
		exampleContent, err := ctl.renderer.RenderString("exampleProfessional.html", H{})
		if err == nil {
			shownContent = exampleContent
		}
	}
	shownVersion := "Current"
	return c.Render(200, "profileEdit.html", H{"page": page, "action": action, "shownContent": shownContent, "shownVersion": shownVersion, "user": u})
}

func (ctl *Controller) setMailH(c echo.Context) error {
	// TODO no tests yet
	u := currentUser(c)
	if u == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not logged in")
	}

	email := strings.TrimSpace(c.FormValue("email"))

	// do nothing if the mail has not been updated
	if email == "" || (u.Email != nil && *u.Email == email) {
		return c.Redirect(http.StatusSeeOther, "/me")
	}

	// update the email
	u.Email = &email
	u.EmailVerified = false
	err := ctl.model.SaveUser(u)
	if err != nil {
		//TODO better error handling
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/me")
}

func (ctl *Controller) setPasswordH(c echo.Context) error {
	// TODO no tests yet
	u := currentUser(c)
	if u == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not logged in")
	}
	if u.Email != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "User must have an email set first")
	}

	s, err := session.Get("session", c)
	if err != nil {
		return err
	}

	currpwd := c.FormValue("currpwd")
	newpwd := c.FormValue("newpwd")

	if u.HashedPassword == "" || u.CheckPassword(currpwd) {
		u.SetPassword(newpwd)
		err := ctl.model.SaveUser(u)

		//TODO display errors better
		if err != nil {
			return err
		}

		// do not log user out of this session
		s.Values[SESSION_TOKEN_KEY] = u.SessionToken

		s.AddFlash("Password changed succesfully")
		s.Save(c.Request(), c.Response())

		return c.Redirect(http.StatusFound, "/me")
	}

	setFlash(c, "Error while changing password. Does the old password match?")
	return c.Redirect(http.StatusSeeOther, "/me")
}
