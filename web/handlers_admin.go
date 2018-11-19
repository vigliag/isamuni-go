package web

import (
	"net/http"

	"github.com/vigliag/isamuni-go/model"

	"github.com/labstack/echo"
)

func (ctl *Controller) adminH(c echo.Context) error {
	u := currentUser(c)
	if u == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not logged in")
	}

	var unapprovedVersions []model.ContentVersion
	ctl.model.Db.
		Preload("Page").
		Preload("User").
		Joins("JOIN pages ON pages.id = content_versions.page_id").
		Where("content_versions.id > pages.approved_version_id").
		Find(&unapprovedVersions)

	return c.Render(http.StatusFound, "admin.html", H{"unapproved": unapprovedVersions})
}
