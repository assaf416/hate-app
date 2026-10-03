package models

import (
	"database/sql"

	"insurance/db"
)

type PhoneCall struct {
	ID           int64
	ClientID     int64
	Title        string
	RecordingURL string
	RecordedAt   string
	CreatedAt    string

	ClientName string // joined, for display
}

func ListPhoneCalls() ([]PhoneCall, error) {
	rows, err := db.DB.Query(`
		SELECT pc.id, pc.client_id, pc.title, pc.recording_url, pc.recorded_at, pc.created_at, c.full_name
		FROM phone_calls pc JOIN clients c ON c.id = pc.client_id
		ORDER BY pc.recorded_at DESC, pc.id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []PhoneCall
	for rows.Next() {
		var p PhoneCall
		if err := rows.Scan(&p.ID, &p.ClientID, &p.Title, &p.RecordingURL, &p.RecordedAt, &p.CreatedAt, &p.ClientName); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

var phoneCallSortColumns = map[string]string{
	"title":       "pc.title",
	"client":      "c.full_name",
	"recorded_at": "pc.recorded_at",
}

// ListPhoneCallsPage returns one page of phone calls plus the total row
// count, optionally filtered by free-text search and sorted by a
// whitelisted column.
func ListPhoneCallsPage(page, pageSize int, search, sortKey, dirKey string) ([]PhoneCall, int, error) {
	query := `
		SELECT pc.id, pc.client_id, pc.title, pc.recording_url, pc.recorded_at, pc.created_at, c.full_name, COUNT(*) OVER() AS total
		FROM phone_calls pc JOIN clients c ON c.id = pc.client_id WHERE 1=1`
	var args []any
	if search != "" {
		like := "%" + search + "%"
		query += ` AND (pc.title LIKE ? OR c.full_name LIKE ?)`
		args = append(args, like, like)
	}
	query += ` ORDER BY ` + buildOrderBy(sortKey, dirKey, phoneCallSortColumns, "pc.recorded_at DESC, pc.id DESC") + ` LIMIT ? OFFSET ?`
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []PhoneCall
	total := 0
	for rows.Next() {
		var p PhoneCall
		if err := rows.Scan(&p.ID, &p.ClientID, &p.Title, &p.RecordingURL, &p.RecordedAt, &p.CreatedAt, &p.ClientName, &total); err != nil {
			return nil, 0, err
		}
		out = append(out, p)
	}
	return out, total, rows.Err()
}

func ListPhoneCallsByClient(clientID int64) ([]PhoneCall, error) {
	rows, err := db.DB.Query(`SELECT id, client_id, title, recording_url, recorded_at, created_at FROM phone_calls WHERE client_id = ? ORDER BY recorded_at DESC, id DESC`, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []PhoneCall
	for rows.Next() {
		var p PhoneCall
		if err := rows.Scan(&p.ID, &p.ClientID, &p.Title, &p.RecordingURL, &p.RecordedAt, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func GetPhoneCall(id int64) (*PhoneCall, error) {
	var p PhoneCall
	err := db.DB.QueryRow(`
		SELECT pc.id, pc.client_id, pc.title, pc.recording_url, pc.recorded_at, pc.created_at, c.full_name
		FROM phone_calls pc JOIN clients c ON c.id = pc.client_id WHERE pc.id = ?`, id).
		Scan(&p.ID, &p.ClientID, &p.Title, &p.RecordingURL, &p.RecordedAt, &p.CreatedAt, &p.ClientName)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func CreatePhoneCall(p *PhoneCall) (int64, error) {
	res, err := db.DB.Exec(`INSERT INTO phone_calls (client_id, title, recording_url, recorded_at) VALUES (?, ?, ?, ?)`,
		p.ClientID, p.Title, p.RecordingURL, p.RecordedAt)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func UpdatePhoneCall(p *PhoneCall) error {
	_, err := db.DB.Exec(`UPDATE phone_calls SET client_id = ?, title = ?, recording_url = ?, recorded_at = ? WHERE id = ?`,
		p.ClientID, p.Title, p.RecordingURL, p.RecordedAt, p.ID)
	return err
}

func DeletePhoneCall(id int64) error {
	_, err := db.DB.Exec(`DELETE FROM phone_calls WHERE id = ?`, id)
	return err
}
