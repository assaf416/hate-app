package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"insurance/models"
	paymentview "insurance/templates/payment"
)

func ListPayments(c echo.Context) error {
	page := ParsePage(c)
	lq := ParseListQuery(c)
	payments, total, err := models.ListPaymentsPage(page, PageSize, lq.Search, lq.Status, lq.Sort, lq.Dir)
	if err != nil {
		return err
	}
	policies, err := models.ListPolicies()
	if err != nil {
		return err
	}
	pageInfo := BuildPageInfo(page, total, "/payments", lq)
	return Render(c, "payments", paymentview.List(payments, policies, c.QueryParam("new") == "1", pageInfo))
}

func NewPaymentForm(c echo.Context) error {
	policies, err := models.ListPolicies()
	if err != nil {
		return err
	}
	return paymentview.Form(nil, policies).Render(c.Request().Context(), c.Response())
}

func EditPaymentForm(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	pm, err := models.GetPayment(id)
	if err != nil {
		return err
	}
	if pm == nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	policies, err := models.ListPolicies()
	if err != nil {
		return err
	}
	return paymentview.Form(pm, policies).Render(c.Request().Context(), c.Response())
}

func ClosePaymentModal(c echo.Context) error {
	return paymentview.Empty().Render(c.Request().Context(), c.Response())
}

func parsePaymentForm(c echo.Context) (*models.Payment, error) {
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
	return &models.Payment{
		ClientID:    policy.ClientID,
		PolicyID:    policyID,
		Amount:      amount,
		PaymentDate: c.FormValue("payment_date"),
		Method:      c.FormValue("method"),
		Status:      c.FormValue("status"),
	}, nil
}

func CreatePayment(c echo.Context) error {
	pm, err := parsePaymentForm(c)
	if err != nil {
		return err
	}
	if _, err := models.CreatePayment(pm); err != nil {
		return err
	}
	return RedirectToReferer(c, "/payments")
}

func UpdatePayment(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	pm, err := parsePaymentForm(c)
	if err != nil {
		return err
	}
	pm.ID = id
	if err := models.UpdatePayment(pm); err != nil {
		return err
	}
	return RedirectToReferer(c, "/payments")
}

func DeletePayment(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := models.DeletePayment(id); err != nil {
		return err
	}
	return c.NoContent(http.StatusOK)
}
