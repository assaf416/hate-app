package router

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"insurance/handlers"
)

func New() *echo.Echo {
	e := echo.New()
	e.Use(middleware.Recover())

	e.Static("/static", "static")

	e.GET("/", handlers.Dashboard)
	e.GET("/search", handlers.Search)
	e.GET("/chat", handlers.Chat)

	e.GET("/clients", handlers.ListClients)
	e.GET("/clients/new", handlers.NewClientForm)
	e.GET("/clients/close", handlers.CloseModal)
	e.GET("/clients/:id", handlers.ShowClient)
	e.GET("/clients/:id/edit", handlers.EditClientForm)
	e.POST("/clients", handlers.CreateClient)
	e.PUT("/clients/:id", handlers.UpdateClient)
	e.DELETE("/clients/:id", handlers.DeleteClient)

	e.GET("/policies", handlers.ListPolicies)
	e.GET("/policies/new", handlers.NewPolicyForm)
	e.GET("/policies/close", handlers.ClosePolicyModal)
	e.GET("/policies/:id", handlers.ShowPolicy)
	e.GET("/policies/:id/edit", handlers.EditPolicyForm)
	e.POST("/policies", handlers.CreatePolicy)
	e.PUT("/policies/:id", handlers.UpdatePolicy)
	e.DELETE("/policies/:id", handlers.DeletePolicy)
	e.POST("/policies/:id/approve", handlers.ApprovePolicy)

	e.GET("/claims", handlers.ListClaims)
	e.GET("/claims/new", handlers.NewClaimForm)
	e.GET("/claims/close", handlers.CloseClaimModal)
	e.GET("/claims/:id/edit", handlers.EditClaimForm)
	e.POST("/claims", handlers.CreateClaim)
	e.PUT("/claims/:id", handlers.UpdateClaim)
	e.DELETE("/claims/:id", handlers.DeleteClaim)

	e.GET("/payments", handlers.ListPayments)
	e.GET("/payments/new", handlers.NewPaymentForm)
	e.GET("/payments/close", handlers.ClosePaymentModal)
	e.GET("/payments/:id/edit", handlers.EditPaymentForm)
	e.POST("/payments", handlers.CreatePayment)
	e.PUT("/payments/:id", handlers.UpdatePayment)
	e.DELETE("/payments/:id", handlers.DeletePayment)

	e.POST("/attachments/upload", handlers.UploadAttachment)
	e.DELETE("/attachments/:id", handlers.DeleteAttachmentHandler)
	e.GET("/attachments/:id/download", handlers.DownloadAttachment)

	return e
}
