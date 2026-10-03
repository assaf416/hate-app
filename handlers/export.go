package handlers

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/xuri/excelize/v2"

	"insurance/models"
)

func formatAmount(v float64) string {
	return strconv.FormatFloat(v, 'f', 2, 64)
}

// writeXLSX builds a single-sheet workbook from headers+rows and streams it
// to the client as a downloadable .xlsx file.
func writeXLSX(c echo.Context, filename string, headers []string, rows [][]string) error {
	f := excelize.NewFile()
	defer f.Close()

	const sheet = "Sheet1"
	headerStyle, err := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	if err != nil {
		return err
	}

	for i, h := range headers {
		cell, err := excelize.CoordinatesToCellName(i+1, 1)
		if err != nil {
			return err
		}
		if err := f.SetCellValue(sheet, cell, h); err != nil {
			return err
		}
	}
	lastCol, err := excelize.CoordinatesToCellName(len(headers), 1)
	if err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "A1", lastCol, headerStyle); err != nil {
		return err
	}

	for r, row := range rows {
		for col, val := range row {
			cell, err := excelize.CoordinatesToCellName(col+1, r+2)
			if err != nil {
				return err
			}
			if err := f.SetCellValue(sheet, cell, val); err != nil {
				return err
			}
		}
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return err
	}

	c.Response().Header().Set("Content-Disposition", `attachment; filename="`+url.QueryEscape(filename)+`"`)
	return c.Blob(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

func ExportClients(c echo.Context) error {
	lq := ParseListQuery(c)
	clients, err := models.ListClientsAll(lq.Search, lq.Sort, lq.Dir)
	if err != nil {
		return err
	}
	headers := []string{"שם מלא", "תעודת זהות", "אימייל", "טלפון", "כתובת"}
	rows := make([][]string, len(clients))
	for i, cl := range clients {
		rows[i] = []string{cl.FullName, cl.NationalID, cl.Email, cl.Phone, cl.Address}
	}
	return writeXLSX(c, "לקוחות.xlsx", headers, rows)
}

func ExportPolicies(c echo.Context) error {
	lq := ParseListQuery(c)
	policies, err := models.ListPoliciesAll(lq.Search, lq.Status, lq.Sort, lq.Dir)
	if err != nil {
		return err
	}
	headers := []string{"מספר פוליסה", "לקוח", "סוג", "תחילה", "סיום", "פרמיה", "סטטוס", "אישור"}
	rows := make([][]string, len(policies))
	for i, p := range policies {
		approval := "טרם אושרה"
		if p.IsApproved() {
			approval = "מאושרת"
		}
		rows[i] = []string{
			p.PolicyNumber, p.ClientName, p.PolicyType, p.StartDate, p.EndDate,
			formatAmount(p.Premium), p.Status, approval,
		}
	}
	return writeXLSX(c, "פוליסות.xlsx", headers, rows)
}

func ExportClaims(c echo.Context) error {
	lq := ParseListQuery(c)
	claims, err := models.ListClaimsAll(lq.Search, lq.Status, lq.Sort, lq.Dir)
	if err != nil {
		return err
	}
	headers := []string{"מספר תביעה", "לקוח", "פוליסה", "תאריך הגשה", "סכום", "סטטוס"}
	rows := make([][]string, len(claims))
	for i, cl := range claims {
		rows[i] = []string{cl.ClaimNumber, cl.ClientName, cl.PolicyNumber, cl.FiledDate, formatAmount(cl.Amount), cl.Status}
	}
	return writeXLSX(c, "תביעות.xlsx", headers, rows)
}

func ExportPayments(c echo.Context) error {
	lq := ParseListQuery(c)
	payments, err := models.ListPaymentsAll(lq.Search, lq.Status, lq.Sort, lq.Dir)
	if err != nil {
		return err
	}
	headers := []string{"תאריך", "לקוח", "פוליסה", "אמצעי תשלום", "סכום", "סטטוס"}
	rows := make([][]string, len(payments))
	for i, pm := range payments {
		rows[i] = []string{pm.PaymentDate, pm.ClientName, pm.PolicyNumber, pm.Method, formatAmount(pm.Amount), pm.Status}
	}
	return writeXLSX(c, "תשלומים.xlsx", headers, rows)
}

func ExportPhoneCalls(c echo.Context) error {
	lq := ParseListQuery(c)
	calls, err := models.ListPhoneCallsAll(lq.Search, lq.Sort, lq.Dir)
	if err != nil {
		return err
	}
	headers := []string{"כותרת", "לקוח", "תאריך שיחה", "קישור להקלטה"}
	rows := make([][]string, len(calls))
	for i, call := range calls {
		rows[i] = []string{call.Title, call.ClientName, call.RecordedAt, call.RecordingURL}
	}
	return writeXLSX(c, "יומן_הקלטות.xlsx", headers, rows)
}
