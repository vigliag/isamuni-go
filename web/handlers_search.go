package web

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/vigliag/isamuni-go/index"
	"github.com/vigliag/isamuni-go/model"
)

func (ctl *Controller) searchH(c echo.Context) error {
	query := c.FormValue("query")

	var resProfessionals []index.SearchResult
	var resCompanies []index.SearchResult
	var resCommunities []index.SearchResult

	if query == "" {
		return c.Render(http.StatusOK, "pageSearch.html", H{})
	}

	results, err := ctl.index.SearchPagesByQueryString(query)
	if err != nil {
		return err
	}

	for _, res := range results {
		switch res.Page.Type {
		case model.PageUser:
			resProfessionals = append(resProfessionals, res)
		case model.PageCommunity:
			resCommunities = append(resCommunities, res)
		case model.PageCompany:
			resCompanies = append(resCompanies, res)
		}
	}

	return c.Render(http.StatusOK, "pageSearch.html",
		H{"professionals": resProfessionals,
			"communities": resCommunities,
			"companies":   resCompanies,
			"query":       query})
}
