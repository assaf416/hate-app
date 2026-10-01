package handlers

import (
	"net/http"
	"net/url"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"

	"insurance/templates/layout"
)

// Render writes `content` wrapped in the full page layout for normal
// navigations, or just the fragment itself for HTMX-driven swaps.
func Render(c echo.Context, active string, content templ.Component) error {
	c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Response().WriteHeader(http.StatusOK)

	if c.Request().Header.Get("HX-Request") == "true" {
		return content.Render(c.Request().Context(), c.Response())
	}
	return layout.Base(active).Render(templ.WithChildren(c.Request().Context(), content), c.Response())
}

// RedirectToReferer sends an HX-Redirect back to the page the HTMX request
// came from, so modal forms work correctly regardless of which list/detail
// page they were opened from. Falls back to fallback if there's no Referer.
func RedirectToReferer(c echo.Context, fallback string) error {
	target := c.Request().Referer()
	if target == "" {
		target = fallback
	} else if u, err := url.Parse(target); err == nil {
		q := u.Query()
		q.Del("new")
		u.RawQuery = q.Encode()
		target = u.String()
	}
	c.Response().Header().Set("HX-Redirect", target)
	return c.NoContent(http.StatusOK)
}
