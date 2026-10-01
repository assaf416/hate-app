package handlers

import (
	"github.com/labstack/echo/v4"

	chatview "insurance/templates/chat"
)

func Chat(c echo.Context) error {
	return Render(c, "chat", chatview.Page())
}
