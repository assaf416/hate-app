package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"insurance/models"
	phonecallview "insurance/templates/phonecall"
)

func ListPhoneCalls(c echo.Context) error {
	page := ParsePage(c)
	lq := ParseListQuery(c)
	calls, total, err := models.ListPhoneCallsPage(page, PageSize, lq.Search, lq.Sort, lq.Dir)
	if err != nil {
		return err
	}
	clients, err := models.ListClients()
	if err != nil {
		return err
	}
	pageInfo := BuildPageInfo(page, total, "/recordings", lq)
	return Render(c, "recordings", phonecallview.List(calls, clients, c.QueryParam("new") == "1", pageInfo))
}

func NewPhoneCallForm(c echo.Context) error {
	clients, err := models.ListClients()
	if err != nil {
		return err
	}
	presetClientID, _ := strconv.ParseInt(c.QueryParam("client_id"), 10, 64)
	return phonecallview.Form(nil, clients, presetClientID).Render(c.Request().Context(), c.Response())
}

func EditPhoneCallForm(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	pc, err := models.GetPhoneCall(id)
	if err != nil {
		return err
	}
	if pc == nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	clients, err := models.ListClients()
	if err != nil {
		return err
	}
	return phonecallview.Form(pc, clients, 0).Render(c.Request().Context(), c.Response())
}

func ClosePhoneCallModal(c echo.Context) error {
	return phonecallview.Empty().Render(c.Request().Context(), c.Response())
}

func parsePhoneCallForm(c echo.Context) (*models.PhoneCall, error) {
	clientID, err := strconv.ParseInt(c.FormValue("client_id"), 10, 64)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid client_id")
	}
	return &models.PhoneCall{
		ClientID:     clientID,
		Title:        c.FormValue("title"),
		RecordingURL: c.FormValue("recording_url"),
		RecordedAt:   c.FormValue("recorded_at"),
	}, nil
}

func CreatePhoneCall(c echo.Context) error {
	pc, err := parsePhoneCallForm(c)
	if err != nil {
		return err
	}
	if _, err := models.CreatePhoneCall(pc); err != nil {
		return err
	}
	return RedirectToReferer(c, "/recordings")
}

func UpdatePhoneCall(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	pc, err := parsePhoneCallForm(c)
	if err != nil {
		return err
	}
	pc.ID = id
	if err := models.UpdatePhoneCall(pc); err != nil {
		return err
	}
	return RedirectToReferer(c, "/recordings")
}

func DeletePhoneCall(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := models.DeletePhoneCall(id); err != nil {
		return err
	}
	return c.NoContent(http.StatusOK)
}
