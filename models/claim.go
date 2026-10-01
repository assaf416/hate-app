package models

import (
	"database/sql"

	"insurance/db"
)

type Claim struct {
	ID          int64
	ClientID    int64
	PolicyID    int64
	ClaimNumber string
	Description string
	Amount      float64
	Status      string
	FiledDate   string
	CreatedAt   string

	ClientName   string
	PolicyNumber string
}

func ListClaims() ([]Claim, error) {
	rows, err := db.DB.Query(`
		SELECT cl.id, cl.client_id, cl.policy_id, cl.claim_number, cl.description, cl.amount, cl.status, cl.filed_date, cl.created_at, c.full_name, p.policy_number
		FROM claims cl
		JOIN clients c ON c.id = cl.client_id
		JOIN policies p ON p.id = cl.policy_id
		ORDER BY cl.id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Claim
	for rows.Next() {
		var c Claim
		if err := rows.Scan(&c.ID, &c.ClientID, &c.PolicyID, &c.ClaimNumber, &c.Description, &c.Amount, &c.Status, &c.FiledDate, &c.CreatedAt, &c.ClientName, &c.PolicyNumber); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func ListClaimsByClient(clientID int64) ([]Claim, error) {
	rows, err := db.DB.Query(`
		SELECT cl.id, cl.client_id, cl.policy_id, cl.claim_number, cl.description, cl.amount, cl.status, cl.filed_date, cl.created_at, p.policy_number
		FROM claims cl JOIN policies p ON p.id = cl.policy_id
		WHERE cl.client_id = ? ORDER BY cl.id DESC`, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Claim
	for rows.Next() {
		var c Claim
		if err := rows.Scan(&c.ID, &c.ClientID, &c.PolicyID, &c.ClaimNumber, &c.Description, &c.Amount, &c.Status, &c.FiledDate, &c.CreatedAt, &c.PolicyNumber); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func ListClaimsByPolicy(policyID int64) ([]Claim, error) {
	rows, err := db.DB.Query(`
		SELECT cl.id, cl.client_id, cl.policy_id, cl.claim_number, cl.description, cl.amount, cl.status, cl.filed_date, cl.created_at, c.full_name
		FROM claims cl JOIN clients c ON c.id = cl.client_id
		WHERE cl.policy_id = ? ORDER BY cl.id DESC`, policyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Claim
	for rows.Next() {
		var c Claim
		if err := rows.Scan(&c.ID, &c.ClientID, &c.PolicyID, &c.ClaimNumber, &c.Description, &c.Amount, &c.Status, &c.FiledDate, &c.CreatedAt, &c.ClientName); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func GetClaim(id int64) (*Claim, error) {
	var c Claim
	err := db.DB.QueryRow(`
		SELECT cl.id, cl.client_id, cl.policy_id, cl.claim_number, cl.description, cl.amount, cl.status, cl.filed_date, cl.created_at, c2.full_name, p.policy_number
		FROM claims cl
		JOIN clients c2 ON c2.id = cl.client_id
		JOIN policies p ON p.id = cl.policy_id
		WHERE cl.id = ?`, id).
		Scan(&c.ID, &c.ClientID, &c.PolicyID, &c.ClaimNumber, &c.Description, &c.Amount, &c.Status, &c.FiledDate, &c.CreatedAt, &c.ClientName, &c.PolicyNumber)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func CreateClaim(c *Claim) (int64, error) {
	res, err := db.DB.Exec(`INSERT INTO claims (client_id, policy_id, claim_number, description, amount, status, filed_date) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		c.ClientID, c.PolicyID, c.ClaimNumber, c.Description, c.Amount, c.Status, c.FiledDate)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func UpdateClaim(c *Claim) error {
	_, err := db.DB.Exec(`UPDATE claims SET client_id = ?, policy_id = ?, claim_number = ?, description = ?, amount = ?, status = ?, filed_date = ? WHERE id = ?`,
		c.ClientID, c.PolicyID, c.ClaimNumber, c.Description, c.Amount, c.Status, c.FiledDate, c.ID)
	return err
}

func DeleteClaim(id int64) error {
	_, err := db.DB.Exec(`DELETE FROM claims WHERE id = ?`, id)
	return err
}
