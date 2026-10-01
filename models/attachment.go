package models

import (
	"database/sql"
	"fmt"
	"strings"

	"insurance/db"
)

type Attachment struct {
	ID            int64
	EntityType    string // "client" | "policy" | "claim"
	EntityID      int64
	FileName      string
	ContentType   string
	ContentBase64 string
	SizeBytes     int64
	UploadedAt    string
}

// IconClass returns a Bootstrap Icons class matching the file's extension,
// so Word/Excel/PDF/image attachments are visually distinguishable in the UI.
func (a Attachment) IconClass() string {
	switch {
	case strings.HasSuffix(strings.ToLower(a.FileName), ".pdf"):
		return "bi bi-file-earmark-pdf-fill text-danger"
	case strings.HasSuffix(strings.ToLower(a.FileName), ".doc"), strings.HasSuffix(strings.ToLower(a.FileName), ".docx"):
		return "bi bi-file-earmark-word-fill text-primary"
	case strings.HasSuffix(strings.ToLower(a.FileName), ".xls"), strings.HasSuffix(strings.ToLower(a.FileName), ".xlsx"):
		return "bi bi-file-earmark-excel-fill text-success"
	case strings.HasSuffix(strings.ToLower(a.FileName), ".png"), strings.HasSuffix(strings.ToLower(a.FileName), ".jpg"), strings.HasSuffix(strings.ToLower(a.FileName), ".jpeg"):
		return "bi bi-file-earmark-image-fill text-warning"
	default:
		return "bi bi-file-earmark-fill text-secondary"
	}
}

// SizeLabel renders the attachment size as a human-friendly KB/MB string.
func (a Attachment) SizeLabel() string {
	kb := float64(a.SizeBytes) / 1024
	if kb < 1024 {
		return fmt.Sprintf("%.0f KB", kb)
	}
	return fmt.Sprintf("%.1f MB", kb/1024)
}

func ListAttachments(entityType string, entityID int64) ([]Attachment, error) {
	rows, err := db.DB.Query(`SELECT id, entity_type, entity_id, file_name, content_type, size_bytes, uploaded_at FROM attachments WHERE entity_type = ? AND entity_id = ? ORDER BY id DESC`, entityType, entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Attachment
	for rows.Next() {
		var a Attachment
		if err := rows.Scan(&a.ID, &a.EntityType, &a.EntityID, &a.FileName, &a.ContentType, &a.SizeBytes, &a.UploadedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func GetAttachment(id int64) (*Attachment, error) {
	var a Attachment
	err := db.DB.QueryRow(`SELECT id, entity_type, entity_id, file_name, content_type, content_base64, size_bytes, uploaded_at FROM attachments WHERE id = ?`, id).
		Scan(&a.ID, &a.EntityType, &a.EntityID, &a.FileName, &a.ContentType, &a.ContentBase64, &a.SizeBytes, &a.UploadedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func CreateAttachment(a *Attachment) (int64, error) {
	res, err := db.DB.Exec(`INSERT INTO attachments (entity_type, entity_id, file_name, content_type, content_base64, size_bytes) VALUES (?, ?, ?, ?, ?, ?)`,
		a.EntityType, a.EntityID, a.FileName, a.ContentType, a.ContentBase64, a.SizeBytes)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func DeleteAttachment(id int64) error {
	_, err := db.DB.Exec(`DELETE FROM attachments WHERE id = ?`, id)
	return err
}
