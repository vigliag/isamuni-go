package web

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/vigliag/isamuni-go/model"
)

func homeH(c echo.Context) error {
	stats, err := model.GetSiteStats()
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "home.html", H{"stats": stats})
}
