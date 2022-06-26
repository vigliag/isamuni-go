package web

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo-contrib/session"

	"github.com/labstack/echo/v4"
	"github.com/vigliag/isamuni-go/mail"
	"github.com/vigliag/isamuni-go/model"
)

func (ctl *Controller) SendEmailVerification(u *model.User) error {
	if u.Email == nil {
		return fmt.Errorf("Can't verify mail, no mail set for user")
	}

	token, err := ctl.model.CreateToken(u.ID, 18)
	if err != nil {
		fmt.Println("Could not create token")
		return err
	}

	confirmationurl := ctl.appURL + "/confirmMail?token=" + token

	err = ctl.mailer.SendMail(mail.ConfirmationEmail(u.Username, *u.Email, confirmationurl))
	if err != nil {
		return err
	}

	return nil
}

func (ctl *Controller) mailVerificationH(c echo.Context) error {
	tokenValue := c.QueryParam("token")

	if tokenValue == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid token")
	}

	t, err := ctl.model.GetToken(tokenValue)
	if err != nil {
		c.Logger().Error("Could not find valid token")
		return err
	}

	u := ctl.model.RetrieveUser(t.Identifier)
	if u == nil {
		c.Logger().Error("Could not find user corresponding to token")
		return err
	}

	u.EmailVerified = true
	err = ctl.model.Db.Save(&u).Error
	if err != nil {
		c.Logger().Error("Could not set email as verified")
		return err
	}

	err = ctl.model.DeleteToken(tokenValue)
	if err != nil {
		c.Logger().Error("Could not delete the token after use")
		return err
	}

	s, err := session.Get("session", c)
	if err != nil {
		return err
	}
	s.AddFlash("Email correctly confirmed!")
	s.Save(c.Request(), c.Response())

	return c.Redirect(http.StatusFound, "/")
}
