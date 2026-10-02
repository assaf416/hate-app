// Package stepdefinitions implements the Go glue code behind the Hebrew
// Gherkin scenarios in features/*.feature. Each scenario gets a fresh SQLite
// file and a fresh in-process HTTP server so scenarios never leak state.
package stepdefinitions

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/cucumber/godog"

	"insurance/db"
	"insurance/router"
)

type testState struct {
	server        *httptest.Server
	dbPath        string
	lastStatus    int
	lastBody      string
	lastPaymentID int64
}

func (s *testState) reset() error {
	f, err := os.CreateTemp("", "insurance-test-*.db")
	if err != nil {
		return err
	}
	path := f.Name()
	f.Close()
	os.Remove(path)

	if err := db.Init(path); err != nil {
		return err
	}
	s.dbPath = path
	s.server = httptest.NewServer(router.New())
	return nil
}

func (s *testState) cleanup() {
	if s.server != nil {
		s.server.Close()
	}
	if s.dbPath != "" {
		os.Remove(s.dbPath)
	}
}

func (s *testState) get(path string) error {
	resp, err := http.Get(s.server.URL + path)
	if err != nil {
		return err
	}
	return s.capture(resp)
}

func (s *testState) postForm(path string, form url.Values) error {
	resp, err := http.PostForm(s.server.URL+path, form)
	if err != nil {
		return err
	}
	return s.capture(resp)
}

func (s *testState) putForm(path string, form url.Values) error {
	req, err := http.NewRequest(http.MethodPut, s.server.URL+path, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	return s.capture(resp)
}

func (s *testState) delete(path string) error {
	req, err := http.NewRequest(http.MethodDelete, s.server.URL+path, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	return s.capture(resp)
}

func (s *testState) uploadFile(path, fieldFileName, content string) error {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("file", fieldFileName)
	if err != nil {
		return err
	}
	if _, err := part.Write([]byte(content)); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, s.server.URL+path, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	return s.capture(resp)
}

func (s *testState) capture(resp *http.Response) error {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	s.lastStatus = resp.StatusCode
	s.lastBody = string(body)
	return nil
}

func pageURL(name string) (string, error) {
	switch name {
	case "לקוחות":
		return "/clients", nil
	case "פוליסות":
		return "/policies", nil
	case "תביעות":
		return "/claims", nil
	case "תשלומים":
		return "/payments", nil
	case "לוח הבקרה":
		return "/", nil
	default:
		return "", fmt.Errorf("עמוד לא מוכר: %s", name)
	}
}

func clientIDByName(name string) (int64, error) {
	var id int64
	err := db.DB.QueryRow(`SELECT id FROM clients WHERE full_name = ?`, name).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("לקוח לא נמצא: %s", name)
	}
	return id, err
}

func policyIDByNumber(number string) (int64, error) {
	var id int64
	err := db.DB.QueryRow(`SELECT id FROM policies WHERE policy_number = ?`, number).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("פוליסה לא נמצאה: %s", number)
	}
	return id, err
}

func claimIDByNumber(number string) (int64, error) {
	var id int64
	err := db.DB.QueryRow(`SELECT id FROM claims WHERE claim_number = ?`, number).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("תביעה לא נמצאה: %s", number)
	}
	return id, err
}

func attachmentIDByName(fileName string) (int64, error) {
	var id int64
	err := db.DB.QueryRow(`SELECT id FROM attachments WHERE file_name = ?`, fileName).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("קובץ לא נמצא: %s", fileName)
	}
	return id, err
}

// InitializeScenario registers every Hebrew step used under features/*.feature.
func InitializeScenario(sc *godog.ScenarioContext) {
	state := &testState{}

	sc.Before(func(ctx context.Context, s *godog.Scenario) (context.Context, error) {
		return ctx, state.reset()
	})
	sc.After(func(ctx context.Context, s *godog.Scenario, err error) (context.Context, error) {
		state.cleanup()
		return ctx, nil
	})

	sc.Given(`^מסד הנתונים ריק$`, func() error { return nil })

	sc.Then(`^התגובה תקינה$`, func() error {
		if state.lastStatus < 200 || state.lastStatus >= 300 {
			return fmt.Errorf("סטטוס לא תקין: %d, גוף: %s", state.lastStatus, state.lastBody)
		}
		return nil
	})

	sc.Then(`^הטקסט "([^"]+)" מופיע בעמוד$`, func(text string) error {
		if !strings.Contains(state.lastBody, text) {
			return fmt.Errorf("הטקסט %q לא נמצא בעמוד", text)
		}
		return nil
	})

	sc.When(`^אני פותח את עמוד "([^"]+)"$`, func(name string) error {
		path, err := pageURL(name)
		if err != nil {
			return err
		}
		return state.get(path)
	})

	// --- Clients ---

	sc.When(`^אני יוצר לקוח עם שם "([^"]+)" תעודת זהות "([^"]+)" אימייל "([^"]+)" טלפון "([^"]+)" וכתובת "([^"]+)"$`,
		func(name, nationalID, email, phone, address string) error {
			return state.postForm("/clients", url.Values{
				"full_name":   {name},
				"national_id": {nationalID},
				"email":       {email},
				"phone":       {phone},
				"address":     {address},
			})
		})

	sc.Then(`^הלקוח "([^"]+)" מופיע ברשימת הלקוחות$`, func(name string) error {
		if err := state.get("/clients"); err != nil {
			return err
		}
		if !strings.Contains(state.lastBody, name) {
			return fmt.Errorf("הלקוח %q לא נמצא ברשימה", name)
		}
		return nil
	})

	sc.Then(`^הלקוח "([^"]+)" לא מופיע ברשימת הלקוחות$`, func(name string) error {
		if err := state.get("/clients"); err != nil {
			return err
		}
		if strings.Contains(state.lastBody, name) {
			return fmt.Errorf("הלקוח %q עדיין מופיע ברשימה", name)
		}
		return nil
	})

	sc.When(`^אני פותח את כרטיס הלקוח "([^"]+)"$`, func(name string) error {
		id, err := clientIDByName(name)
		if err != nil {
			return err
		}
		return state.get("/clients/" + strconv.FormatInt(id, 10))
	})

	sc.When(`^אני מעדכן את הלקוח "([^"]+)" לשם "([^"]+)"$`, func(oldName, newName string) error {
		id, err := clientIDByName(oldName)
		if err != nil {
			return err
		}
		return state.putForm("/clients/"+strconv.FormatInt(id, 10), url.Values{
			"full_name":   {newName},
			"national_id": {"000000000"},
			"email":       {""},
			"phone":       {""},
			"address":     {""},
		})
	})

	sc.When(`^אני מוחק את הלקוח "([^"]+)"$`, func(name string) error {
		id, err := clientIDByName(name)
		if err != nil {
			return err
		}
		return state.delete("/clients/" + strconv.FormatInt(id, 10))
	})

	// --- Policies ---

	sc.When(`^אני יוצר פוליסה מספר "([^"]+)" מסוג "([^"]+)" ללקוח "([^"]+)" בתאריכים "([^"]+)" עד "([^"]+)" בפרמיה "([^"]+)"$`,
		func(number, ptype, clientName, start, end, premium string) error {
			clientID, err := clientIDByName(clientName)
			if err != nil {
				return err
			}
			return state.postForm("/policies", url.Values{
				"client_id":     {strconv.FormatInt(clientID, 10)},
				"policy_number": {number},
				"policy_type":   {ptype},
				"start_date":    {start},
				"end_date":      {end},
				"premium":       {premium},
				"status":        {"פעילה"},
			})
		})

	sc.Then(`^הפוליסה "([^"]+)" מופיעה ברשימת הפוליסות$`, func(number string) error {
		if err := state.get("/policies"); err != nil {
			return err
		}
		if !strings.Contains(state.lastBody, number) {
			return fmt.Errorf("הפוליסה %q לא נמצאה ברשימה", number)
		}
		return nil
	})

	sc.Then(`^הפוליסה "([^"]+)" לא מופיעה ברשימת הפוליסות$`, func(number string) error {
		if err := state.get("/policies"); err != nil {
			return err
		}
		if strings.Contains(state.lastBody, number) {
			return fmt.Errorf("הפוליסה %q עדיין מופיעה ברשימה", number)
		}
		return nil
	})

	sc.When(`^אני מעדכן את הפוליסה "([^"]+)" לסטטוס "([^"]+)"$`, func(number, status string) error {
		id, err := policyIDByNumber(number)
		if err != nil {
			return err
		}
		var p struct {
			clientID                                     int64
			ptype, start, end, statusCur                 string
			premium                                       float64
		}
		err = db.DB.QueryRow(`SELECT client_id, policy_type, start_date, end_date, premium FROM policies WHERE id = ?`, id).
			Scan(&p.clientID, &p.ptype, &p.start, &p.end, &p.premium)
		if err != nil {
			return err
		}
		return state.putForm("/policies/"+strconv.FormatInt(id, 10), url.Values{
			"client_id":     {strconv.FormatInt(p.clientID, 10)},
			"policy_number": {number},
			"policy_type":   {p.ptype},
			"start_date":    {p.start},
			"end_date":      {p.end},
			"premium":       {strconv.FormatFloat(p.premium, 'f', 2, 64)},
			"status":        {status},
		})
	})

	sc.Then(`^הפוליסה "([^"]+)" מופיעה ברשימת הפוליסות בסטטוס "([^"]+)"$`, func(number, status string) error {
		if err := state.get("/policies"); err != nil {
			return err
		}
		if !strings.Contains(state.lastBody, number) {
			return fmt.Errorf("הפוליסה %q לא נמצאה ברשימה", number)
		}
		if !strings.Contains(state.lastBody, status) {
			return fmt.Errorf("הסטטוס %q לא נמצא בעמוד", status)
		}
		return nil
	})

	sc.When(`^אני מוחק את הפוליסה "([^"]+)"$`, func(number string) error {
		id, err := policyIDByNumber(number)
		if err != nil {
			return err
		}
		return state.delete("/policies/" + strconv.FormatInt(id, 10))
	})

	sc.Then(`^הפוליסה "([^"]+)" מופיעה ברשימת הפוליסות כטרם מאושרת$`, func(number string) error {
		if err := state.get("/policies"); err != nil {
			return err
		}
		if !strings.Contains(state.lastBody, number) {
			return fmt.Errorf("הפוליסה %q לא נמצאה ברשימה", number)
		}
		if !strings.Contains(state.lastBody, "טרם אושרה") {
			return fmt.Errorf("הפוליסה %q לא מסומנת כטרם מאושרת", number)
		}
		return nil
	})

	sc.When(`^אני מאשר את הפוליסה "([^"]+)" על ידי משתמש מספר "([^"]+)"$`, func(number, userID string) error {
		id, err := policyIDByNumber(number)
		if err != nil {
			return err
		}
		return state.postForm("/policies/"+strconv.FormatInt(id, 10)+"/approve", url.Values{
			"approved_by_user_id": {userID},
		})
	})

	sc.Then(`^הפוליסה "([^"]+)" מופיעה ברשימת הפוליסות כמאושרת$`, func(number string) error {
		if err := state.get("/policies"); err != nil {
			return err
		}
		if !strings.Contains(state.lastBody, number) {
			return fmt.Errorf("הפוליסה %q לא נמצאה ברשימה", number)
		}
		if !strings.Contains(state.lastBody, "מאושרת") {
			return fmt.Errorf("הפוליסה %q לא מסומנת כמאושרת", number)
		}
		return nil
	})

	sc.Then(`^כרטיס הפוליסה "([^"]+)" מציג שהיא אושרה על ידי משתמש מספר "([^"]+)"$`, func(number, userID string) error {
		id, err := policyIDByNumber(number)
		if err != nil {
			return err
		}
		if err := state.get("/policies/" + strconv.FormatInt(id, 10)); err != nil {
			return err
		}
		if !strings.Contains(state.lastBody, userID) {
			return fmt.Errorf("מזהה המשתמש המאשר %q לא מופיע בכרטיס הפוליסה", userID)
		}
		return nil
	})

	// --- Claims ---

	sc.When(`^אני יוצר תביעה מספר "([^"]+)" על הפוליסה "([^"]+)" בתאריך "([^"]+)" בסכום "([^"]+)" ותיאור "([^"]+)"$`,
		func(number, policyNumber, date, amount, description string) error {
			policyID, err := policyIDByNumber(policyNumber)
			if err != nil {
				return err
			}
			return state.postForm("/claims", url.Values{
				"policy_id":    {strconv.FormatInt(policyID, 10)},
				"claim_number": {number},
				"description":  {description},
				"amount":       {amount},
				"status":       {"פתוחה"},
				"filed_date":   {date},
			})
		})

	sc.Then(`^התביעה "([^"]+)" מופיעה ברשימת התביעות בסטטוס "([^"]+)"$`, func(number, status string) error {
		if err := state.get("/claims"); err != nil {
			return err
		}
		if !strings.Contains(state.lastBody, number) {
			return fmt.Errorf("התביעה %q לא נמצאה ברשימה", number)
		}
		if !strings.Contains(state.lastBody, status) {
			return fmt.Errorf("הסטטוס %q לא נמצא בעמוד", status)
		}
		return nil
	})

	sc.When(`^אני מעדכן את התביעה "([^"]+)" לסטטוס "([^"]+)"$`, func(number, status string) error {
		id, err := claimIDByNumber(number)
		if err != nil {
			return err
		}
		var c struct {
			policyID            int64
			description, filed  string
			amount              float64
		}
		err = db.DB.QueryRow(`SELECT policy_id, description, amount, filed_date FROM claims WHERE id = ?`, id).
			Scan(&c.policyID, &c.description, &c.amount, &c.filed)
		if err != nil {
			return err
		}
		return state.putForm("/claims/"+strconv.FormatInt(id, 10), url.Values{
			"policy_id":    {strconv.FormatInt(c.policyID, 10)},
			"claim_number": {number},
			"description":  {c.description},
			"amount":       {strconv.FormatFloat(c.amount, 'f', 2, 64)},
			"status":       {status},
			"filed_date":   {c.filed},
		})
	})

	sc.When(`^אני מוחק את התביעה "([^"]+)"$`, func(number string) error {
		id, err := claimIDByNumber(number)
		if err != nil {
			return err
		}
		return state.delete("/claims/" + strconv.FormatInt(id, 10))
	})

	sc.Then(`^התביעה "([^"]+)" לא מופיעה ברשימת התביעות$`, func(number string) error {
		if err := state.get("/claims"); err != nil {
			return err
		}
		if strings.Contains(state.lastBody, number) {
			return fmt.Errorf("התביעה %q עדיין מופיעה ברשימה", number)
		}
		return nil
	})

	// --- Payments ---

	sc.When(`^אני יוצר תשלום על הפוליסה "([^"]+)" בתאריך "([^"]+)" בסכום "([^"]+)" באמצעי "([^"]+)"$`,
		func(policyNumber, date, amount, method string) error {
			policyID, err := policyIDByNumber(policyNumber)
			if err != nil {
				return err
			}
			if err := state.postForm("/payments", url.Values{
				"policy_id":    {strconv.FormatInt(policyID, 10)},
				"payment_date": {date},
				"amount":       {amount},
				"method":       {method},
				"status":       {"שולם"},
			}); err != nil {
				return err
			}
			return db.DB.QueryRow(`SELECT id FROM payments ORDER BY id DESC LIMIT 1`).Scan(&state.lastPaymentID)
		})

	sc.Then(`^התשלום על הפוליסה "([^"]+)" מופיע ברשימת התשלומים בסכום "([^"]+)"$`, func(policyNumber, amount string) error {
		if err := state.get("/payments"); err != nil {
			return err
		}
		if !strings.Contains(state.lastBody, policyNumber) || !strings.Contains(state.lastBody, amount) {
			return fmt.Errorf("התשלום על %q בסכום %q לא נמצא ברשימה", policyNumber, amount)
		}
		return nil
	})

	sc.When(`^אני מעדכן את התשלום האחרון לסטטוס "([^"]+)"$`, func(status string) error {
		var p struct {
			clientID, policyID int64
			date, method       string
			amount             float64
		}
		err := db.DB.QueryRow(`SELECT client_id, policy_id, amount, payment_date, method FROM payments WHERE id = ?`, state.lastPaymentID).
			Scan(&p.clientID, &p.policyID, &p.amount, &p.date, &p.method)
		if err != nil {
			return err
		}
		return state.putForm("/payments/"+strconv.FormatInt(state.lastPaymentID, 10), url.Values{
			"policy_id":    {strconv.FormatInt(p.policyID, 10)},
			"payment_date": {p.date},
			"amount":       {strconv.FormatFloat(p.amount, 'f', 2, 64)},
			"method":       {p.method},
			"status":       {status},
		})
	})

	sc.Then(`^התשלום על הפוליסה "([^"]+)" מופיע ברשימת התשלומים בסטטוס "([^"]+)"$`, func(policyNumber, status string) error {
		if err := state.get("/payments"); err != nil {
			return err
		}
		if !strings.Contains(state.lastBody, policyNumber) {
			return fmt.Errorf("התשלום על %q לא נמצא ברשימה", policyNumber)
		}
		if !strings.Contains(state.lastBody, status) {
			return fmt.Errorf("הסטטוס %q לא נמצא בעמוד", status)
		}
		return nil
	})

	sc.When(`^אני מוחק את התשלום האחרון$`, func() error {
		return state.delete("/payments/" + strconv.FormatInt(state.lastPaymentID, 10))
	})

	sc.Then(`^אין תשלום בסכום "([^"]+)" ברשימת התשלומים$`, func(amount string) error {
		if err := state.get("/payments"); err != nil {
			return err
		}
		if strings.Contains(state.lastBody, amount) {
			return fmt.Errorf("תשלום בסכום %q עדיין מופיע ברשימה", amount)
		}
		return nil
	})

	// --- Attachments ---

	sc.When(`^אני מעלה קובץ בשם "([^"]+)" ללקוח "([^"]+)"$`, func(fileName, clientName string) error {
		clientID, err := clientIDByName(clientName)
		if err != nil {
			return err
		}
		path := fmt.Sprintf("/attachments/upload?entity_type=client&entity_id=%d", clientID)
		return state.uploadFile(path, fileName, "תוכן בדיקה")
	})

	sc.Then(`^הקובץ "([^"]+)" מופיע בכרטיס הלקוח "([^"]+)"$`, func(fileName, clientName string) error {
		id, err := clientIDByName(clientName)
		if err != nil {
			return err
		}
		if err := state.get("/clients/" + strconv.FormatInt(id, 10)); err != nil {
			return err
		}
		if !strings.Contains(state.lastBody, fileName) {
			return fmt.Errorf("הקובץ %q לא נמצא בכרטיס הלקוח", fileName)
		}
		return nil
	})

	sc.When(`^אני מוחק את הקובץ "([^"]+)" מהלקוח "([^"]+)"$`, func(fileName, clientName string) error {
		id, err := attachmentIDByName(fileName)
		if err != nil {
			return err
		}
		return state.delete("/attachments/" + strconv.FormatInt(id, 10))
	})

	sc.Then(`^הקובץ "([^"]+)" לא מופיע בכרטיס הלקוח "([^"]+)"$`, func(fileName, clientName string) error {
		id, err := clientIDByName(clientName)
		if err != nil {
			return err
		}
		if err := state.get("/clients/" + strconv.FormatInt(id, 10)); err != nil {
			return err
		}
		if strings.Contains(state.lastBody, fileName) {
			return fmt.Errorf("הקובץ %q עדיין מופיע בכרטיס הלקוח", fileName)
		}
		return nil
	})
}
