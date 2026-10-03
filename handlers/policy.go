package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"insurance/models"
	policyview "insurance/templates/policy"
)

func ListPolicies(c echo.Context) error {
	page := ParsePage(c)
	lq := ParseListQuery(c)
	policies, total, err := models.ListPoliciesPage(page, PageSize, lq.Search, lq.Status, lq.Sort, lq.Dir)
	if err != nil {
		return err
	}
	clients, err := models.ListClients()
	if err != nil {
		return err
	}
	pageInfo := BuildPageInfo(page, total, "/policies", lq)
	return Render(c, "policies", policyview.List(policies, clients, c.QueryParam("new") == "1", pageInfo))
}

func NewPolicyForm(c echo.Context) error {
	clients, err := models.ListClients()
	if err != nil {
		return err
	}
	presetClientID, _ := strconv.ParseInt(c.QueryParam("client_id"), 10, 64)
	return policyview.Form(nil, clients, presetClientID).Render(c.Request().Context(), c.Response())
}

func EditPolicyForm(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	p, err := models.GetPolicy(id)
	if err != nil {
		return err
	}
	if p == nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	clients, err := models.ListClients()
	if err != nil {
		return err
	}
	return policyview.Form(p, clients, 0).Render(c.Request().Context(), c.Response())
}

func ClosePolicyModal(c echo.Context) error {
	return policyview.Empty().Render(c.Request().Context(), c.Response())
}

func parsePolicyForm(c echo.Context) (*models.Policy, error) {
	clientID, err := strconv.ParseInt(c.FormValue("client_id"), 10, 64)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid client_id")
	}
	premium, _ := strconv.ParseFloat(c.FormValue("premium"), 64)
	return &models.Policy{
		ClientID:     clientID,
		PolicyNumber: c.FormValue("policy_number"),
		PolicyType:   c.FormValue("policy_type"),
		StartDate:    c.FormValue("start_date"),
		EndDate:      c.FormValue("end_date"),
		Premium:      premium,
		Status:       c.FormValue("status"),
	}, nil
}

func CreatePolicy(c echo.Context) error {
	p, err := parsePolicyForm(c)
	if err != nil {
		return err
	}
	id, err := models.CreatePolicy(p)
	if err != nil {
		return err
	}
	// Land on the new policy's own page so the client can attach documents
	// (contract, ID scans, etc.) right away as part of the creation flow.
	c.Response().Header().Set("HX-Redirect", "/policies/"+strconv.FormatInt(id, 10))
	return c.NoContent(http.StatusOK)
}

func UpdatePolicy(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	p, err := parsePolicyForm(c)
	if err != nil {
		return err
	}
	p.ID = id
	if err := models.UpdatePolicy(p); err != nil {
		return err
	}
	return RedirectToReferer(c, "/policies")
}

func ApprovePolicy(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	approvedByUserID, err := strconv.ParseInt(c.FormValue("approved_by_user_id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid approved_by_user_id")
	}
	if err := models.ApprovePolicy(id, approvedByUserID); err != nil {
		return err
	}
	p, err := models.GetPolicy(id)
	if err != nil {
		return err
	}
	return policyview.ApprovalPanel(*p).Render(c.Request().Context(), c.Response())
}

func DeletePolicy(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := models.DeletePolicy(id); err != nil {
		return err
	}
	return c.NoContent(http.StatusOK)
}

func ShowPolicy(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	p, err := models.GetPolicy(id)
	if err != nil {
		return err
	}
	if p == nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	claims, err := models.ListClaimsByPolicy(id)
	if err != nil {
		return err
	}
	payments, err := models.ListPaymentsByPolicy(id)
	if err != nil {
		return err
	}
	attachments, err := models.ListAttachments("policy", id)
	if err != nil {
		return err
	}
	return Render(c, "policies", policyview.Detail(*p, claims, payments, attachments))
}
