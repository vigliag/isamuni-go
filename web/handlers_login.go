package web

import (
	"log"
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"github.com/vigliag/isamuni-go/model"
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

	user := model.LoginEmail(email, password)
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
