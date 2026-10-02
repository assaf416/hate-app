package handlers

import (
	"strconv"

	"github.com/labstack/echo/v4"

	"insurance/templates/layout"
)

const PageSize = 20

// ParsePage reads the "page" query param, defaulting to and clamping at 1.
func ParsePage(c echo.Context) int {
	page, err := strconv.Atoi(c.QueryParam("page"))
	if err != nil || page < 1 {
		return 1
	}
	return page
}

// BuildPageInfo computes display info (total pages, etc.) for a paginated list.
func BuildPageInfo(page, total int, basePath string) layout.PageInfo {
	totalPages := (total + PageSize - 1) / PageSize
	if totalPages < 1 {
		totalPages = 1
	}
	if page > totalPages {
		page = totalPages
	}
	return layout.PageInfo{
		Page:       page,
		PageSize:   PageSize,
		Total:      total,
		TotalPages: totalPages,
		BasePath:   basePath,
	}
}
