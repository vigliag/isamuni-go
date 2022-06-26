package web

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (ctl *Controller) homeH(c echo.Context) error {
	stats, err := ctl.model.GetSiteStats()
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "home.html", H{"stats": stats})
}
