package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	claimview "insurance/templates/claim"
	"insurance/models"
)

func ListClaims(c echo.Context) error {
	claims, err := models.ListClaims()
	if err != nil {
		return err
	}
	policies, err := models.ListPolicies()
	if err != nil {
		return err
	}
	return Render(c, "claims", claimview.List(claims, policies, c.QueryParam("new") == "1"))
}

func NewClaimForm(c echo.Context) error {
	policies, err := models.ListPolicies()
	if err != nil {
		return err
	}
	return claimview.Form(nil, policies).Render(c.Request().Context(), c.Response())
}

func EditClaimForm(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	cl, err := models.GetClaim(id)
	if err != nil {
		return err
	}
	if cl == nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	policies, err := models.ListPolicies()
	if err != nil {
		return err
	}
	return claimview.Form(cl, policies).Render(c.Request().Context(), c.Response())
}

func CloseClaimModal(c echo.Context) error {
	return claimview.Empty().Render(c.Request().Context(), c.Response())
}

func parseClaimForm(c echo.Context) (*models.Claim, error) {
	policyID, err := strconv.ParseInt(c.FormValue("policy_id"), 10, 64)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid policy_id")
	}
	policy, err := models.GetPolicy(policyID)
	if err != nil {
		return nil, err
	}
	if policy == nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "policy not found")
	}
	amount, _ := strconv.ParseFloat(c.FormValue("amount"), 64)
	return &models.Claim{
		ClientID:    policy.ClientID,
		PolicyID:    policyID,
		ClaimNumber: c.FormValue("claim_number"),
		Description: c.FormValue("description"),
		Amount:      amount,
		Status:      c.FormValue("status"),
		FiledDate:   c.FormValue("filed_date"),
	}, nil
}

func CreateClaim(c echo.Context) error {
	cl, err := parseClaimForm(c)
	if err != nil {
		return err
	}
	if _, err := models.CreateClaim(cl); err != nil {
		return err
	}
	return RedirectToReferer(c, "/claims")
}

func UpdateClaim(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	cl, err := parseClaimForm(c)
	if err != nil {
		return err
	}
	cl.ID = id
	if err := models.UpdateClaim(cl); err != nil {
		return err
	}
	return RedirectToReferer(c, "/claims")
}

func DeleteClaim(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := models.DeleteClaim(id); err != nil {
		return err
	}
	return c.NoContent(http.StatusOK)
}
