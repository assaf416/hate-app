package models

import "insurance/db"

type SearchResult struct {
	Kind  string // "client" | "policy"
	ID    int64
	Title string
	Sub   string
}

func Search(query string) ([]SearchResult, error) {
	like := "%" + query + "%"
	var out []SearchResult

	clientRows, err := db.DB.Query(`SELECT id, full_name, national_id FROM clients WHERE full_name LIKE ? ORDER BY full_name LIMIT 8`, like)
	if err != nil {
		return nil, err
	}
	defer clientRows.Close()
	for clientRows.Next() {
		var id int64
		var name, nationalID string
		if err := clientRows.Scan(&id, &name, &nationalID); err != nil {
			return nil, err
		}
		out = append(out, SearchResult{Kind: "client", ID: id, Title: name, Sub: "ת.ז. " + nationalID})
	}
	if err := clientRows.Err(); err != nil {
		return nil, err
	}

	policyRows, err := db.DB.Query(`
		SELECT p.id, p.policy_number, c.full_name
		FROM policies p JOIN clients c ON c.id = p.client_id
		WHERE p.policy_number LIKE ? ORDER BY p.policy_number LIMIT 8`, like)
	if err != nil {
		return nil, err
	}
	defer policyRows.Close()
	for policyRows.Next() {
		var id int64
		var number, clientName string
		if err := policyRows.Scan(&id, &number, &clientName); err != nil {
			return nil, err
		}
		out = append(out, SearchResult{Kind: "policy", ID: id, Title: number, Sub: clientName})
	}
	return out, policyRows.Err()
}
