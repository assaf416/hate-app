package handlers

import (
	"github.com/labstack/echo/v4"

	"insurance/models"
	searchview "insurance/templates/search"
)

func Search(c echo.Context) error {
	q := c.QueryParam("q")
	var results []models.SearchResult
	if q != "" {
		var err error
		results, err = models.Search(q)
		if err != nil {
			return err
		}
	}
	return searchview.Results(q, results).Render(c.Request().Context(), c.Response())
}
