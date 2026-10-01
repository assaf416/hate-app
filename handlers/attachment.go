package handlers

import (
	"encoding/base64"
	"io"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/labstack/echo/v4"

	"insurance/models"
	clientview "insurance/templates/client"
	policyview "insurance/templates/policy"
)

var allowedEntityTypes = map[string]bool{
	"client": true,
	"policy": true,
	"claim":  true,
}

// allowedAttachmentExt restricts uploads to the document types clients
// actually send us (Word, Excel, PDF, plus common image scans).
var allowedAttachmentExt = map[string]bool{
	".pdf":  true,
	".doc":  true,
	".docx": true,
	".xls":  true,
	".xlsx": true,
	".png":  true,
	".jpg":  true,
	".jpeg": true,
}

func UploadAttachment(c echo.Context) error {
	entityType := c.QueryParam("entity_type")
	if !allowedEntityTypes[entityType] {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid entity_type")
	}
	entityID, err := strconv.ParseInt(c.QueryParam("entity_id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid entity_id")
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "missing file")
	}
	ext := filepath.Ext(fileHeader.Filename)
	if !allowedAttachmentExt[ext] {
		return echo.NewHTTPError(http.StatusBadRequest, "סוג קובץ לא נתמך")
	}

	src, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	content, err := io.ReadAll(src)
	if err != nil {
		return err
	}

	att := &models.Attachment{
		EntityType:    entityType,
		EntityID:      entityID,
		FileName:      filepath.Base(fileHeader.Filename),
		ContentType:   fileHeader.Header.Get("Content-Type"),
		ContentBase64: base64.StdEncoding.EncodeToString(content),
		SizeBytes:     fileHeader.Size,
	}
	if _, err := models.CreateAttachment(att); err != nil {
		return err
	}

	attachments, err := models.ListAttachments(entityType, entityID)
	if err != nil {
		return err
	}

	switch entityType {
	case "policy":
		return policyview.AttachmentList(entityID, attachments).Render(c.Request().Context(), c.Response())
	default:
		return clientview.AttachmentList(entityType, entityID, attachments).Render(c.Request().Context(), c.Response())
	}
}

func DeleteAttachmentHandler(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	att, err := models.GetAttachment(id)
	if err != nil {
		return err
	}
	if att == nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	if err := models.DeleteAttachment(id); err != nil {
		return err
	}
	return c.NoContent(http.StatusOK)
}

func DownloadAttachment(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	att, err := models.GetAttachment(id)
	if err != nil {
		return err
	}
	if att == nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	content, err := base64.StdEncoding.DecodeString(att.ContentBase64)
	if err != nil {
		return err
	}
	c.Response().Header().Set("Content-Disposition", `attachment; filename="`+att.FileName+`"`)
	return c.Blob(http.StatusOK, att.ContentType, content)
}
