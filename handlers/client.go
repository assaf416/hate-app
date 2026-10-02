package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"insurance/models"
	clientview "insurance/templates/client"
	"insurance/templates/dashboard"
)

func Dashboard(c echo.Context) error {
	clients, err := models.ListClients()
	if err != nil {
		return err
	}
	policies, err := models.ListPolicies()
	if err != nil {
		return err
	}
	claims, err := models.ListClaims()
	if err != nil {
		return err
	}
	payments, err := models.ListPayments()
	if err != nil {
		return err
	}
	return Render(c, "dashboard", dashboard.Dashboard(len(clients), len(policies), len(claims), len(payments)))
}

func ListClients(c echo.Context) error {
	page := ParsePage(c)
	clients, total, err := models.ListClientsPage(page, PageSize)
	if err != nil {
		return err
	}
	pageInfo := BuildPageInfo(page, total, "/clients")
	return Render(c, "clients", clientview.List(clients, c.QueryParam("new") == "1", pageInfo))
}

func NewClientForm(c echo.Context) error {
	return clientview.Form(nil).Render(c.Request().Context(), c.Response())
}

func EditClientForm(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	cl, err := models.GetClient(id)
	if err != nil {
		return err
	}
	if cl == nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	return clientview.Form(cl).Render(c.Request().Context(), c.Response())
}

func CloseModal(c echo.Context) error {
	return clientview.Empty().Render(c.Request().Context(), c.Response())
}

func CreateClient(c echo.Context) error {
	cl := &models.Client{
		FullName:   c.FormValue("full_name"),
		NationalID: c.FormValue("national_id"),
		Email:      c.FormValue("email"),
		Phone:      c.FormValue("phone"),
		Address:    c.FormValue("address"),
	}
	if _, err := models.CreateClient(cl); err != nil {
		return err
	}
	return RedirectToReferer(c, "/clients")
}

func UpdateClient(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	cl := &models.Client{
		ID:         id,
		FullName:   c.FormValue("full_name"),
		NationalID: c.FormValue("national_id"),
		Email:      c.FormValue("email"),
		Phone:      c.FormValue("phone"),
		Address:    c.FormValue("address"),
	}
	if err := models.UpdateClient(cl); err != nil {
		return err
	}
	return RedirectToReferer(c, "/clients")
}

func DeleteClient(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := models.DeleteClient(id); err != nil {
		return err
	}
	return c.NoContent(http.StatusOK)
}

func ShowClient(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	cl, err := models.GetClient(id)
	if err != nil {
		return err
	}
	if cl == nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	policies, err := models.ListPoliciesByClient(id)
	if err != nil {
		return err
	}
	claims, err := models.ListClaimsByClient(id)
	if err != nil {
		return err
	}
	payments, err := models.ListPaymentsByClient(id)
	if err != nil {
		return err
	}
	attachments, err := models.ListAttachments("client", id)
	if err != nil {
		return err
	}
	calls, err := models.ListPhoneCallsByClient(id)
	if err != nil {
		return err
	}
	return Render(c, "clients", clientview.Detail(*cl, policies, claims, payments, attachments, calls))
}
