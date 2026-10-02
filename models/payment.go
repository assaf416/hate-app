package models

import (
	"database/sql"

	"insurance/db"
)

type Payment struct {
	ID          int64
	ClientID    int64
	PolicyID    int64
	Amount      float64
	PaymentDate string
	Method      string
	Status      string
	CreatedAt   string

	ClientName   string
	PolicyNumber string
}

func ListPayments() ([]Payment, error) {
	rows, err := db.DB.Query(`
		SELECT pm.id, pm.client_id, pm.policy_id, pm.amount, pm.payment_date, pm.method, pm.status, pm.created_at, c.full_name, p.policy_number
		FROM payments pm
		JOIN clients c ON c.id = pm.client_id
		JOIN policies p ON p.id = pm.policy_id
		ORDER BY pm.id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Payment
	for rows.Next() {
		var p Payment
		if err := rows.Scan(&p.ID, &p.ClientID, &p.PolicyID, &p.Amount, &p.PaymentDate, &p.Method, &p.Status, &p.CreatedAt, &p.ClientName, &p.PolicyNumber); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// ListPaymentsPage returns one page of payments plus the total row count.
func ListPaymentsPage(page, pageSize int) ([]Payment, int, error) {
	rows, err := db.DB.Query(`
		SELECT pm.id, pm.client_id, pm.policy_id, pm.amount, pm.payment_date, pm.method, pm.status, pm.created_at, c.full_name, p.policy_number, COUNT(*) OVER() AS total
		FROM payments pm
		JOIN clients c ON c.id = pm.client_id
		JOIN policies p ON p.id = pm.policy_id
		ORDER BY pm.id DESC LIMIT ? OFFSET ?`, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []Payment
	total := 0
	for rows.Next() {
		var p Payment
		if err := rows.Scan(&p.ID, &p.ClientID, &p.PolicyID, &p.Amount, &p.PaymentDate, &p.Method, &p.Status, &p.CreatedAt, &p.ClientName, &p.PolicyNumber, &total); err != nil {
			return nil, 0, err
		}
		out = append(out, p)
	}
	return out, total, rows.Err()
}

func ListPaymentsByClient(clientID int64) ([]Payment, error) {
	rows, err := db.DB.Query(`
		SELECT pm.id, pm.client_id, pm.policy_id, pm.amount, pm.payment_date, pm.method, pm.status, pm.created_at, p.policy_number
		FROM payments pm JOIN policies p ON p.id = pm.policy_id
		WHERE pm.client_id = ? ORDER BY pm.id DESC`, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Payment
	for rows.Next() {
		var p Payment
		if err := rows.Scan(&p.ID, &p.ClientID, &p.PolicyID, &p.Amount, &p.PaymentDate, &p.Method, &p.Status, &p.CreatedAt, &p.PolicyNumber); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func ListPaymentsByPolicy(policyID int64) ([]Payment, error) {
	rows, err := db.DB.Query(`
		SELECT pm.id, pm.client_id, pm.policy_id, pm.amount, pm.payment_date, pm.method, pm.status, pm.created_at, c.full_name
		FROM payments pm JOIN clients c ON c.id = pm.client_id
		WHERE pm.policy_id = ? ORDER BY pm.id DESC`, policyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Payment
	for rows.Next() {
		var p Payment
		if err := rows.Scan(&p.ID, &p.ClientID, &p.PolicyID, &p.Amount, &p.PaymentDate, &p.Method, &p.Status, &p.CreatedAt, &p.ClientName); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func GetPayment(id int64) (*Payment, error) {
	var p Payment
	err := db.DB.QueryRow(`
		SELECT pm.id, pm.client_id, pm.policy_id, pm.amount, pm.payment_date, pm.method, pm.status, pm.created_at, c.full_name, pol.policy_number
		FROM payments pm
		JOIN clients c ON c.id = pm.client_id
		JOIN policies pol ON pol.id = pm.policy_id
		WHERE pm.id = ?`, id).
		Scan(&p.ID, &p.ClientID, &p.PolicyID, &p.Amount, &p.PaymentDate, &p.Method, &p.Status, &p.CreatedAt, &p.ClientName, &p.PolicyNumber)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func CreatePayment(p *Payment) (int64, error) {
	res, err := db.DB.Exec(`INSERT INTO payments (client_id, policy_id, amount, payment_date, method, status) VALUES (?, ?, ?, ?, ?, ?)`,
		p.ClientID, p.PolicyID, p.Amount, p.PaymentDate, p.Method, p.Status)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func UpdatePayment(p *Payment) error {
	_, err := db.DB.Exec(`UPDATE payments SET client_id = ?, policy_id = ?, amount = ?, payment_date = ?, method = ?, status = ? WHERE id = ?`,
		p.ClientID, p.PolicyID, p.Amount, p.PaymentDate, p.Method, p.Status, p.ID)
	return err
}

func DeletePayment(id int64) error {
	_, err := db.DB.Exec(`DELETE FROM payments WHERE id = ?`, id)
	return err
}
