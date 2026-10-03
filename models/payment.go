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

var paymentSortColumns = map[string]string{
	"payment_date": "pm.payment_date",
	"client":       "c.full_name",
	"policy":       "p.policy_number",
	"method":       "pm.method",
	"amount":       "pm.amount",
	"status":       "pm.status",
}

// ListPaymentsPage returns one page of payments plus the total row count,
// optionally filtered by free-text search and/or status, and sorted by a
// whitelisted column.
func ListPaymentsPage(page, pageSize int, search, status, sortKey, dirKey string) ([]Payment, int, error) {
	query := `
		SELECT pm.id, pm.client_id, pm.policy_id, pm.amount, pm.payment_date, pm.method, pm.status, pm.created_at, c.full_name, p.policy_number, COUNT(*) OVER() AS total
		FROM payments pm
		JOIN clients c ON c.id = pm.client_id
		JOIN policies p ON p.id = pm.policy_id WHERE 1=1`
	var args []any
	if search != "" {
		like := "%" + search + "%"
		query += ` AND (c.full_name LIKE ? OR p.policy_number LIKE ? OR pm.method LIKE ?)`
		args = append(args, like, like, like)
	}
	if status != "" {
		query += ` AND pm.status = ?`
		args = append(args, status)
	}
	query += ` ORDER BY ` + buildOrderBy(sortKey, dirKey, paymentSortColumns, "pm.id DESC") + ` LIMIT ? OFFSET ?`
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := db.DB.Query(query, args...)
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

// ListPaymentsAll returns every payment matching the given filter/sort, with
// no pagination limit -- used for exports.
func ListPaymentsAll(search, status, sortKey, dirKey string) ([]Payment, error) {
	query := `
		SELECT pm.id, pm.client_id, pm.policy_id, pm.amount, pm.payment_date, pm.method, pm.status, pm.created_at, c.full_name, p.policy_number
		FROM payments pm
		JOIN clients c ON c.id = pm.client_id
		JOIN policies p ON p.id = pm.policy_id WHERE 1=1`
	var args []any
	if search != "" {
		like := "%" + search + "%"
		query += ` AND (c.full_name LIKE ? OR p.policy_number LIKE ? OR pm.method LIKE ?)`
		args = append(args, like, like, like)
	}
	if status != "" {
		query += ` AND pm.status = ?`
		args = append(args, status)
	}
	query += ` ORDER BY ` + buildOrderBy(sortKey, dirKey, paymentSortColumns, "pm.id DESC")

	rows, err := db.DB.Query(query, args...)
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
