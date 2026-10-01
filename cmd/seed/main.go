// Command seed populates insurance.db with sample Hebrew data for local
// development and demos: 100 clients, 1000 policies, and 300 attachments.
package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"insurance/db"
	"insurance/models"
)

const (
	numClients     = 100
	numPolicies    = 1000
	numAttachments = 300
)

var (
	firstNames = []string{
		"דני", "רותם", "משה", "שרה", "יוסי", "מיכל", "אבי", "נועה", "עידן", "טל",
		"יעל", "אורי", "ליאור", "שירה", "עומר", "הדר", "אלון", "קרן", "רון", "נטע",
		"גיל", "ענת", "דור", "מאיה", "איתי", "רוני", "נדב", "אפרת", "יובל", "שני",
	}
	lastNames = []string{
		"כהן", "לוי", "מזרחי", "פרץ", "ביטון", "אברהם", "דוד", "אזולאי", "גבאי", "עמר",
		"אוחיון", "שרעבי", "חדד", "מלכה", "רוזן", "פרידמן", "שפירא", "קפלן", "נחום", "אלבז",
	}
	cities = []string{
		"תל אביב", "ירושלים", "חיפה", "באר שבע", "נתניה", "אשדוד", "ראשון לציון",
		"פתח תקווה", "חולון", "רמת גן", "הרצליה", "כפר סבא", "רעננה", "מודיעין",
		"אילת", "נהריה", "עכו", "טבריה", "לוד", "רמלה",
	}
	streets = []string{"הרצל", "ויצמן", "בן גוריון", "רוטשילד", "ז'בוטינסקי", "סוקולוב", "ההגנה", "דיזנגוף"}

	policyTypes = []string{"רכב", "דירה", "חיים", "בריאות", "עסק", "נסיעות"}
	statuses    = []string{"פעילה", "פעילה", "פעילה", "מוקפאת", "מבוטלת", "הסתיימה"}

	paymentMethods = []string{"כרטיס אשראי", "הוראת קבע", "העברה בנקאית", "מזומן"}
	paymentStatus  = []string{"שולם", "שולם", "שולם", "ממתין", "נכשל"}

	claimStatuses = []string{"פתוחה", "בבדיקה", "אושרה", "נדחתה", "נסגרה"}
	claimDescs    = []string{"תאונת דרכים", "נזק לרכוש", "גניבה", "נזקי מים", "שבר בציוד", "פציעה אישית"}

	attachmentBaseNames = []string{
		"תעודת_זהות", "חוזה_ביטוח", "דוח_נזק", "אישור_העברה", "טופס_תביעה",
		"רישיון_רכב", "צילום_רכב", "דוח_שמאי", "חשבונית", "אישור_רפואי",
	}

	fileKinds = []struct {
		ext   string
		ctype string
		magic []byte
	}{
		{".pdf", "application/pdf", []byte("%PDF-1.4\n")},
		{".docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document", []byte("PK\x03\x04")},
		{".xlsx", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", []byte("PK\x03\x04")},
		{".png", "image/png", []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}},
		{".jpg", "image/jpeg", []byte{0xFF, 0xD8, 0xFF}},
	}
)

func main() {
	dbPath := flag.String("db", "insurance.db", "path to the SQLite database file")
	flag.Parse()

	if err := db.Init(*dbPath); err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	rng := rand.New(rand.NewSource(42))

	if err := reset(); err != nil {
		log.Fatalf("failed to reset tables: %v", err)
	}

	clientIDs := seedClients(rng)
	log.Printf("seeded %d clients", len(clientIDs))

	policyIDs := seedPolicies(rng, clientIDs)
	log.Printf("seeded %d policies", len(policyIDs))

	seedClaimsAndPayments(rng, policyIDs)
	log.Printf("seeded claims and payments for a sample of policies")

	seedAttachments(rng, clientIDs, policyIDs)
	log.Printf("seeded %d attachments", numAttachments)

	log.Println("done")
}

func reset() error {
	_, err := db.DB.Exec(`
		DELETE FROM attachments;
		DELETE FROM payments;
		DELETE FROM claims;
		DELETE FROM policies;
		DELETE FROM clients;
	`)
	return err
}

func pick[T any](rng *rand.Rand, items []T) T {
	return items[rng.Intn(len(items))]
}

func seedClients(rng *rand.Rand) []int64 {
	ids := make([]int64, 0, numClients)
	for i := 1; i <= numClients; i++ {
		first := pick(rng, firstNames)
		last := pick(rng, lastNames)
		city := pick(rng, cities)
		street := pick(rng, streets)

		c := &models.Client{
			FullName:   first + " " + last,
			NationalID: fmt.Sprintf("%09d", rng.Intn(1_000_000_000)),
			Email:      fmt.Sprintf("client%03d@example.com", i),
			Phone:      fmt.Sprintf("05%d%07d", rng.Intn(9), rng.Intn(10_000_000)),
			Address:    fmt.Sprintf("%s %d, %s", street, rng.Intn(120)+1, city),
		}
		id, err := models.CreateClient(c)
		if err != nil {
			log.Fatalf("create client %d: %v", i, err)
		}
		ids = append(ids, id)
	}
	return ids
}

func randomDateRange(rng *rand.Rand) (start, end string) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	startDate := base.AddDate(0, rng.Intn(24), rng.Intn(28))
	endDate := startDate.AddDate(1, 0, 0)
	return startDate.Format("2006-01-02"), endDate.Format("2006-01-02")
}

func seedPolicies(rng *rand.Rand, clientIDs []int64) []int64 {
	ids := make([]int64, 0, numPolicies)
	for i := 1; i <= numPolicies; i++ {
		start, end := randomDateRange(rng)
		p := &models.Policy{
			ClientID:     pick(rng, clientIDs),
			PolicyNumber: fmt.Sprintf("POL-%05d", i),
			PolicyType:   pick(rng, policyTypes),
			StartDate:    start,
			EndDate:      end,
			Premium:      float64(rng.Intn(50000)+200) / 10,
			Status:       pick(rng, statuses),
		}
		id, err := models.CreatePolicy(p)
		if err != nil {
			log.Fatalf("create policy %d: %v", i, err)
		}
		ids = append(ids, id)
	}
	return ids
}

// seedClaimsAndPayments adds a handful of claims/payments against a random
// sample of policies, enough to make the dashboard and detail pages feel
// populated without doubling the seeding time.
func seedClaimsAndPayments(rng *rand.Rand, policyIDs []int64) {
	sampleSize := len(policyIDs) / 4
	for i := 1; i <= sampleSize; i++ {
		policyID := pick(rng, policyIDs)
		var clientID int64
		var policyNumber string
		err := db.DB.QueryRow(`SELECT client_id, policy_number FROM policies WHERE id = ?`, policyID).Scan(&clientID, &policyNumber)
		if err != nil {
			log.Fatalf("lookup policy %d: %v", policyID, err)
		}

		if rng.Intn(2) == 0 {
			claim := &models.Claim{
				ClientID:    clientID,
				PolicyID:    policyID,
				ClaimNumber: fmt.Sprintf("CLM-%05d", i),
				Description: pick(rng, claimDescs),
				Amount:      float64(rng.Intn(200000)+500) / 10,
				Status:      pick(rng, claimStatuses),
				FiledDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, rng.Intn(10), rng.Intn(28)).Format("2006-01-02"),
			}
			if _, err := models.CreateClaim(claim); err != nil {
				log.Fatalf("create claim %d: %v", i, err)
			}
		}

		payment := &models.Payment{
			ClientID:    clientID,
			PolicyID:    policyID,
			Amount:      float64(rng.Intn(5000)+50) / 10,
			PaymentDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, rng.Intn(10), rng.Intn(28)).Format("2006-01-02"),
			Method:      pick(rng, paymentMethods),
			Status:      pick(rng, paymentStatus),
		}
		if _, err := models.CreatePayment(payment); err != nil {
			log.Fatalf("create payment %d: %v", i, err)
		}
	}
}

func seedAttachments(rng *rand.Rand, clientIDs, policyIDs []int64) {
	for i := 1; i <= numAttachments; i++ {
		kind := pick(rng, fileKinds)
		baseName := pick(rng, attachmentBaseNames)
		fileName := fmt.Sprintf("%s_%03d%s", baseName, i, kind.ext)

		content := make([]byte, len(kind.magic)+rng.Intn(60_000)+2_000)
		copy(content, kind.magic)
		for j := len(kind.magic); j < len(content); j++ {
			content[j] = byte(rng.Intn(256))
		}

		var entityType string
		var entityID int64
		if rng.Intn(2) == 0 {
			entityType = "client"
			entityID = pick(rng, clientIDs)
		} else {
			entityType = "policy"
			entityID = pick(rng, policyIDs)
		}

		a := &models.Attachment{
			EntityType:    entityType,
			EntityID:      entityID,
			FileName:      fileName,
			ContentType:   kind.ctype,
			ContentBase64: base64.StdEncoding.EncodeToString(content),
			SizeBytes:     int64(len(content)),
		}
		if _, err := models.CreateAttachment(a); err != nil {
			log.Fatalf("create attachment %d: %v", i, err)
		}
	}
}
