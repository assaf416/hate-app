package models

import (
	"database/sql"

	"insurance/db"
)

type Client struct {
	ID         int64
	FullName   string
	NationalID string
	Email      string
	Phone      string
	Address    string
	CreatedAt  string
}

func ListClients() ([]Client, error) {
	rows, err := db.DB.Query(`SELECT id, full_name, national_id, email, phone, address, created_at FROM clients ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Client
	for rows.Next() {
		var c Client
		if err := rows.Scan(&c.ID, &c.FullName, &c.NationalID, &c.Email, &c.Phone, &c.Address, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

var clientSortColumns = map[string]string{
	"name":        "full_name",
	"national_id": "national_id",
	"email":       "email",
	"phone":       "phone",
	"address":     "address",
}

// ListClientsPage returns one page of clients plus the total row count, for
// server-side pagination, optionally filtered by free-text search and
// sorted by a whitelisted column.
func ListClientsPage(page, pageSize int, search, sortKey, dirKey string) ([]Client, int, error) {
	query := `SELECT id, full_name, national_id, email, phone, address, created_at, COUNT(*) OVER() AS total FROM clients WHERE 1=1`
	var args []any
	if search != "" {
		like := "%" + search + "%"
		query += ` AND (full_name LIKE ? OR national_id LIKE ? OR email LIKE ? OR phone LIKE ? OR address LIKE ?)`
		args = append(args, like, like, like, like, like)
	}
	query += ` ORDER BY ` + buildOrderBy(sortKey, dirKey, clientSortColumns, "id DESC") + ` LIMIT ? OFFSET ?`
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []Client
	total := 0
	for rows.Next() {
		var c Client
		if err := rows.Scan(&c.ID, &c.FullName, &c.NationalID, &c.Email, &c.Phone, &c.Address, &c.CreatedAt, &total); err != nil {
			return nil, 0, err
		}
		out = append(out, c)
	}
	return out, total, rows.Err()
}

// ListClientsAll returns every client matching the given filter/sort, with
// no pagination limit -- used for exports.
func ListClientsAll(search, sortKey, dirKey string) ([]Client, error) {
	query := `SELECT id, full_name, national_id, email, phone, address, created_at FROM clients WHERE 1=1`
	var args []any
	if search != "" {
		like := "%" + search + "%"
		query += ` AND (full_name LIKE ? OR national_id LIKE ? OR email LIKE ? OR phone LIKE ? OR address LIKE ?)`
		args = append(args, like, like, like, like, like)
	}
	query += ` ORDER BY ` + buildOrderBy(sortKey, dirKey, clientSortColumns, "id DESC")

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Client
	for rows.Next() {
		var c Client
		if err := rows.Scan(&c.ID, &c.FullName, &c.NationalID, &c.Email, &c.Phone, &c.Address, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func GetClient(id int64) (*Client, error) {
	var c Client
	err := db.DB.QueryRow(`SELECT id, full_name, national_id, email, phone, address, created_at FROM clients WHERE id = ?`, id).
		Scan(&c.ID, &c.FullName, &c.NationalID, &c.Email, &c.Phone, &c.Address, &c.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func CreateClient(c *Client) (int64, error) {
	res, err := db.DB.Exec(`INSERT INTO clients (full_name, national_id, email, phone, address) VALUES (?, ?, ?, ?, ?)`,
		c.FullName, c.NationalID, c.Email, c.Phone, c.Address)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func UpdateClient(c *Client) error {
	_, err := db.DB.Exec(`UPDATE clients SET full_name = ?, national_id = ?, email = ?, phone = ?, address = ? WHERE id = ?`,
		c.FullName, c.NationalID, c.Email, c.Phone, c.Address, c.ID)
	return err
}

func DeleteClient(id int64) error {
	_, err := db.DB.Exec(`DELETE FROM clients WHERE id = ?`, id)
	return err
}
