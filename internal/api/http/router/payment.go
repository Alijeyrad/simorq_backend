package router

import (
	"github.com/Alijeyrad/simorq_backend/internal/api/http/handler"
	"github.com/gofiber/fiber/v3"
)

func (r *Router) registerPaymentRoutes(
	api fiber.Router,
	ph *handler.PaymentHandler,
	authRequired fiber.Handler,
	clinicHeader fiber.Handler,
) {
	// Public: ZarinPal callback (no auth)
	api.Get("/payments/verify", ph.Verify)

	// Auth required only (no clinic context — user-scoped wallet)
	payments := api.Group("/payments", authRequired)
	payments.Get("/wallet", ph.GetWallet)
	payments.Get("/transactions", ph.GetTransactions)

	// Auth + clinic context
	paymentsClinic := api.Group("/payments", authRequired, clinicHeader)
	paymentsClinic.Post("/pay", ph.Initiate)
	paymentsClinic.Post("/wallet/iban", ph.SetIBAN)
	paymentsClinic.Post("/withdraw", ph.Withdraw)
}
