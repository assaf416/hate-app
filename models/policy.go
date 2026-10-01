package models

import (
	"database/sql"

	"insurance/db"
)

type Policy struct {
	ID           int64
	ClientID     int64
	PolicyNumber string
	PolicyType   string
	StartDate    string
	EndDate      string
	Premium      float64
	Status       string
	CreatedAt    string

	ClientName string // joined, for display
}

func ListPolicies() ([]Policy, error) {
	rows, err := db.DB.Query(`
		SELECT p.id, p.client_id, p.policy_number, p.policy_type, p.start_date, p.end_date, p.premium, p.status, p.created_at, c.full_name
		FROM policies p JOIN clients c ON c.id = p.client_id
		ORDER BY p.id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Policy
	for rows.Next() {
		var p Policy
		if err := rows.Scan(&p.ID, &p.ClientID, &p.PolicyNumber, &p.PolicyType, &p.StartDate, &p.EndDate, &p.Premium, &p.Status, &p.CreatedAt, &p.ClientName); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func ListPoliciesByClient(clientID int64) ([]Policy, error) {
	rows, err := db.DB.Query(`SELECT id, client_id, policy_number, policy_type, start_date, end_date, premium, status, created_at FROM policies WHERE client_id = ? ORDER BY id DESC`, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Policy
	for rows.Next() {
		var p Policy
		if err := rows.Scan(&p.ID, &p.ClientID, &p.PolicyNumber, &p.PolicyType, &p.StartDate, &p.EndDate, &p.Premium, &p.Status, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func GetPolicy(id int64) (*Policy, error) {
	var p Policy
	err := db.DB.QueryRow(`
		SELECT p.id, p.client_id, p.policy_number, p.policy_type, p.start_date, p.end_date, p.premium, p.status, p.created_at, c.full_name
		FROM policies p JOIN clients c ON c.id = p.client_id WHERE p.id = ?`, id).
		Scan(&p.ID, &p.ClientID, &p.PolicyNumber, &p.PolicyType, &p.StartDate, &p.EndDate, &p.Premium, &p.Status, &p.CreatedAt, &p.ClientName)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func CreatePolicy(p *Policy) (int64, error) {
	res, err := db.DB.Exec(`INSERT INTO policies (client_id, policy_number, policy_type, start_date, end_date, premium, status) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		p.ClientID, p.PolicyNumber, p.PolicyType, p.StartDate, p.EndDate, p.Premium, p.Status)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func UpdatePolicy(p *Policy) error {
	_, err := db.DB.Exec(`UPDATE policies SET client_id = ?, policy_number = ?, policy_type = ?, start_date = ?, end_date = ?, premium = ?, status = ? WHERE id = ?`,
		p.ClientID, p.PolicyNumber, p.PolicyType, p.StartDate, p.EndDate, p.Premium, p.Status, p.ID)
	return err
}

func DeletePolicy(id int64) error {
	_, err := db.DB.Exec(`DELETE FROM policies WHERE id = ?`, id)
	return err
}
