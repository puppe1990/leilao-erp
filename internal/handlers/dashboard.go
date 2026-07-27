package handlers

import (
	"net/http"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/httpx"
	"github.com/puppe1990/cais/pkg/cais/meta"
	inertia "github.com/romsar/gonertia/v3"

	"github.com/puppe1990/leilao-erp/internal/domain"
	"github.com/puppe1990/leilao-erp/internal/store"
)

type DashboardData struct {
	meta.Site
	TotalCashFormatted       string
	OpenPayablesFormatted    string
	OpenReceivablesFormatted string
	MonthProfitFormatted     string
	OverduePayables          int
	OverdueReceivables       int
	LotCount                 int
	CtaLot                   bool
	Env                      string
}

type DashboardHandler struct {
	renderer *cais.Renderer
	store    store.Store
	site     meta.Site
	cfg      cais.Config
	inertia  *inertia.Inertia
}

func NewDashboardHandler(renderer *cais.Renderer, s store.Store, site meta.Site, cfg cais.Config, i *inertia.Inertia) *DashboardHandler {
	return &DashboardHandler{renderer: renderer, store: s, site: site, cfg: cfg, inertia: i}
}

func (h *DashboardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	summary, err := h.store.DashboardSummary()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	balances := make([]map[string]any, 0, len(summary.CashBalances))
	for _, b := range summary.CashBalances {
		balances = append(balances, map[string]any{
			"id":        b.ID,
			"name":      b.Name,
			"cents":     b.Cents,
			"formatted": domain.FormatBRL(b.Cents),
		})
	}

	ctaLot := summary.LotCount == 0
	totalCashFormatted := domain.FormatBRL(summary.TotalCashCents)
	openPayablesFormatted := domain.FormatBRL(summary.OpenPayablesCents)
	openReceivablesFormatted := domain.FormatBRL(summary.OpenReceivablesCents)
	monthProfitFormatted := domain.FormatBRL(summary.MonthProfitCents)

	if h.inertia != nil {
		_ = h.inertia.Render(w, r, "Dashboard", withCompany(h.store, inertia.Props{
			"site":                     meta.ForRequest(h.site, r),
			"balances":                 balances,
			"totalCashFormatted":       totalCashFormatted,
			"openPayablesFormatted":    openPayablesFormatted,
			"openReceivablesFormatted": openReceivablesFormatted,
			"monthProfitFormatted":     monthProfitFormatted,
			"overduePayables":          summary.OverduePayables,
			"overdueReceivables":       summary.OverdueReceivables,
			"lotCount":                 summary.LotCount,
			"ctaLot":                   ctaLot,
			"env":                      h.cfg.Env,
		}))
		return
	}
	httpx.RenderOrError(w, h.renderer, "base", "dashboard", DashboardData{
		Site:                     meta.ForRequest(h.site, r),
		TotalCashFormatted:       totalCashFormatted,
		OpenPayablesFormatted:    openPayablesFormatted,
		OpenReceivablesFormatted: openReceivablesFormatted,
		MonthProfitFormatted:     monthProfitFormatted,
		OverduePayables:          summary.OverduePayables,
		OverdueReceivables:       summary.OverdueReceivables,
		LotCount:                 summary.LotCount,
		CtaLot:                   ctaLot,
		Env:                      h.cfg.Env,
	}, h.cfg)
}
