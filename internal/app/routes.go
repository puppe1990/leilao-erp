package app

import (
	"net/http"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/middleware"

	"github.com/puppe1990/leilao-erp/internal/handlers"
)

func registerRoutes(r *cais.Router, deps Deps, cfg cais.Config) {
	home := handlers.NewHomeHandler(deps.Renderer, deps.Site, deps.Catalog, cfg, deps.Inertia)
	contact := handlers.NewContactHandler(deps.Renderer, deps.Store, deps.Site, deps.Catalog, cfg, deps.Inertia)
	dashboard := handlers.NewDashboardHandler(deps.Renderer, deps.Store, deps.Site, cfg, deps.Inertia)
	lots := handlers.NewLotsHandler(deps.Renderer, deps.Store, deps.Site, cfg, deps.Inertia)
	stock := handlers.NewStockHandler(deps.Renderer, deps.Store, deps.Site, cfg, deps.Inertia)
	sales := handlers.NewSalesHandler(deps.Renderer, deps.Store, deps.Site, cfg, deps.Inertia)
	cash := handlers.NewCashHandler(deps.Renderer, deps.Store, deps.Site, cfg, deps.Inertia)
	payables := handlers.NewPayablesHandler(deps.Renderer, deps.Store, deps.Site, cfg, deps.Inertia)
	receivables := handlers.NewReceivablesHandler(deps.Renderer, deps.Store, deps.Site, cfg, deps.Inertia)
	configH := handlers.NewConfigHandler(deps.Renderer, deps.Store, deps.Site, cfg, deps.Inertia)
	auth := handlers.NewAuthHandler(deps.Renderer, deps.Store, deps.Site, deps.Store.Sessions(), cfg, deps.Catalog, deps.Inertia)

	loginLimit := middleware.NewRateLimiter(10, cfg)
	resetLimit := middleware.NewRateLimiter(10, cfg)
	contactLimit := middleware.NewRateLimiter(20, cfg)

	r.Get("/", home.ServeHTTP)
	r.Get("/contact", contact.Get)
	r.Post("/contact", contactLimit.Middleware(http.HandlerFunc(contact.Post)).ServeHTTP)
	r.Get("/login", auth.Login)
	r.Post("/login", loginLimit.Middleware(http.HandlerFunc(auth.LoginPost)).ServeHTTP)
	// Public signup disabled — single admin is seeded in development.
	r.Get("/forgot-password", auth.ForgotPassword)
	r.Post("/forgot-password", resetLimit.Middleware(http.HandlerFunc(auth.ForgotPasswordPost)).ServeHTTP)
	r.Get("/reset-password", auth.ResetPassword)
	r.Post("/reset-password", resetLimit.Middleware(http.HandlerFunc(auth.ResetPasswordPost)).ServeHTTP)
	r.Post("/logout", auth.LogoutPost)
	r.Get("/dashboard", middleware.RequireAuthFunc("/login", dashboard.ServeHTTP))

	r.Get("/lots", middleware.RequireAuthFunc("/login", lots.Index))
	r.Get("/lots/new", middleware.RequireAuthFunc("/login", lots.New))
	r.Post("/lots", middleware.RequireAuthFunc("/login", lots.Create))
	r.Get("/lots/{id}", middleware.RequireAuthFunc("/login", cais.IntParam("id", lots.Show)))
	r.Get("/lots/{id}/edit", middleware.RequireAuthFunc("/login", cais.IntParam("id", lots.Edit)))
	r.Post("/lots/{id}", middleware.RequireAuthFunc("/login", cais.IntParam("id", lots.Update)))
	r.Post("/lots/{id}/delete", middleware.RequireAuthFunc("/login", cais.IntParam("id", lots.Destroy)))
	r.Post("/lots/{id}/costs", middleware.RequireAuthFunc("/login", cais.IntParam("id", lots.AddCost)))
	r.Post("/lots/{id}/items/{itemId}", middleware.RequireAuthFunc("/login", cais.IntParam("id", lots.UpdateItem)))

	r.Get("/stock", middleware.RequireAuthFunc("/login", stock.Index))

	r.Get("/sales", middleware.RequireAuthFunc("/login", sales.Index))
	r.Get("/sales/new", middleware.RequireAuthFunc("/login", sales.New))
	r.Post("/sales", middleware.RequireAuthFunc("/login", sales.Create))
	r.Get("/sales/{id}", middleware.RequireAuthFunc("/login", cais.IntParam("id", sales.Show)))
	r.Get("/sales/{id}/edit", middleware.RequireAuthFunc("/login", cais.IntParam("id", sales.Edit)))
	r.Post("/sales/{id}", middleware.RequireAuthFunc("/login", cais.IntParam("id", sales.Update)))
	r.Post("/sales/{id}/delete", middleware.RequireAuthFunc("/login", cais.IntParam("id", sales.Destroy)))
	r.Post("/sales/{id}/cancel", middleware.RequireAuthFunc("/login", cais.IntParam("id", sales.Cancel)))

	r.Get("/cash", middleware.RequireAuthFunc("/login", cash.Index))
	r.Post("/cash/entries", middleware.RequireAuthFunc("/login", cash.CreateManual))
	r.Post("/cash/entries/{id}/delete", middleware.RequireAuthFunc("/login", cais.IntParam("id", cash.DestroyEntry)))
	r.Post("/cash/accounts", middleware.RequireAuthFunc("/login", cash.CreateAccount))
	r.Post("/cash/accounts/{id}", middleware.RequireAuthFunc("/login", cais.IntParam("id", cash.UpdateAccount)))
	r.Post("/cash/accounts/{id}/delete", middleware.RequireAuthFunc("/login", cais.IntParam("id", cash.DestroyAccount)))

	r.Get("/payables", middleware.RequireAuthFunc("/login", payables.Index))
	r.Post("/payables", middleware.RequireAuthFunc("/login", payables.Create))
	r.Post("/payables/{id}/settle", middleware.RequireAuthFunc("/login", cais.IntParam("id", payables.Settle)))
	r.Post("/payables/{id}/cancel", middleware.RequireAuthFunc("/login", cais.IntParam("id", payables.Cancel)))

	r.Get("/receivables", middleware.RequireAuthFunc("/login", receivables.Index))
	r.Post("/receivables", middleware.RequireAuthFunc("/login", receivables.Create))
	r.Post("/receivables/{id}/settle", middleware.RequireAuthFunc("/login", cais.IntParam("id", receivables.Settle)))
	r.Post("/receivables/{id}/cancel", middleware.RequireAuthFunc("/login", cais.IntParam("id", receivables.Cancel)))

	r.Get("/config", middleware.RequireAuthFunc("/login", configH.Index))
	r.Post("/config/company", middleware.RequireAuthFunc("/login", configH.UpdateCompany))
	r.Post("/config/password", middleware.RequireAuthFunc("/login", configH.UpdatePassword))
}
