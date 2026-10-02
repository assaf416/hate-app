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

// ListClientsPage returns one page of clients plus the total row count,
// for server-side pagination.
func ListClientsPage(page, pageSize int) ([]Client, int, error) {
	rows, err := db.DB.Query(`
		SELECT id, full_name, national_id, email, phone, address, created_at, COUNT(*) OVER() AS total
		FROM clients ORDER BY id DESC LIMIT ? OFFSET ?`, pageSize, (page-1)*pageSize)
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
