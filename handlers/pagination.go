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

// ListQueryParams bundles the filter/sort query params shared by every
// list page: free-text search, status filter, and sort column/direction.
type ListQueryParams struct {
	Search string
	Status string
	Sort   string
	Dir    string
}

// ParseListQuery reads q/status/sort/dir from the request query string.
func ParseListQuery(c echo.Context) ListQueryParams {
	return ListQueryParams{
		Search: c.QueryParam("q"),
		Status: c.QueryParam("status"),
		Sort:   c.QueryParam("sort"),
		Dir:    c.QueryParam("dir"),
	}
}

// BuildPageInfo computes display info (total pages, etc.) for a paginated,
// filtered, and sorted list, ready to hand to the layout.Pagination,
// layout.SortableHeader, and layout.FilterBar components.
func BuildPageInfo(page, total int, basePath string, lq ListQueryParams) layout.PageInfo {
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
		Sort:       lq.Sort,
		Dir:        lq.Dir,
		Q:          lq.Search,
		Status:     lq.Status,
	}
}
