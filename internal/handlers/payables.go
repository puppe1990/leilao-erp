package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/meta"
	"github.com/puppe1990/leilao-erp/internal/domain"
	"github.com/puppe1990/leilao-erp/internal/store"
	inertia "github.com/romsar/gonertia/v3"
)

type PayablesHandler struct {
	renderer *cais.Renderer
	store    store.Store
	site     meta.Site
	cfg      cais.Config
	inertia  *inertia.Inertia
}

func NewPayablesHandler(renderer *cais.Renderer, s store.Store, site meta.Site, cfg cais.Config, i *inertia.Inertia) *PayablesHandler {
	return &PayablesHandler{renderer: renderer, store: s, site: site, cfg: cfg, inertia: i}
}

func (h *PayablesHandler) Index(w http.ResponseWriter, r *http.Request) {
	payables, err := h.store.ListPayables()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	accounts, err := h.store.ListCashAccounts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows := make([]map[string]any, 0, len(payables))
	for _, p := range payables {
		paidAt := ""
		if p.PaidAt != nil {
			paidAt = *p.PaidAt
		}
		var lotID any
		if p.LotID != nil {
			lotID = *p.LotID
		}
		rows = append(rows, map[string]any{
			"id":          p.ID,
			"description": p.Description,
			"amount":      domain.FormatBRL(p.AmountCents),
			"dueOn":       p.DueOn,
			"status":      p.Status,
			"statusLabel": payableStatusLabel(p.Status),
			"lotId":       lotID,
			"paidAt":      paidAt,
			"canSettle":   p.Status == "open",
			"canCancel":   p.Status == "open",
		})
	}

	accountOptions := make([]map[string]any, 0, len(accounts))
	for _, a := range accounts {
		accountOptions = append(accountOptions, map[string]any{
			"id":   a.ID,
			"name": a.Name,
			"kind": a.Kind,
		})
	}

	_ = h.inertia.Render(w, r, "Payables/Index", withCompany(h.store, inertia.Props{
		"site":         meta.ForRequest(h.site, r),
		"payables":     rows,
		"cashAccounts": accountOptions,
	}))
}

func (h *PayablesHandler) Settle(w http.ResponseWriter, r *http.Request, id int64) {
	if err := parseFormOrJSON(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	accountID, _ := strconv.ParseInt(strings.TrimSpace(r.FormValue("cash_account_id")), 10, 64)
	paidAt := strings.TrimSpace(r.FormValue("paid_at"))
	if paidAt == "" {
		paidAt = time.Now().UTC().Format("2006-01-02")
	}
	if len(paidAt) == 10 {
		paidAt = paidAt + "T12:00:00Z"
	}

	if accountID <= 0 {
		// Prefer first cash account if only one exists (UI may omit select)
		accounts, err := h.store.ListCashAccounts()
		if err == nil && len(accounts) == 1 {
			accountID = accounts[0].ID
		}
	}

	if accountID <= 0 {
		ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{
			"cash_account_id": "Selecione a conta de caixa",
		})
		// Re-render index with error
		r = r.WithContext(ctx)
		h.Index(w, r)
		return
	}

	if err := h.store.SettlePayable(id, accountID, paidAt); err != nil {
		ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{
			"form": err.Error(),
		})
		r = r.WithContext(ctx)
		h.Index(w, r)
		return
	}

	h.inertia.Redirect(w, r, "/payables", http.StatusSeeOther)
}

func (h *PayablesHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := parseFormOrJSON(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	amount, err := domain.ParseBRLToCents(r.FormValue("amount"))
	if err != nil {
		ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{"amount": "Valor inválido"})
		r = r.WithContext(ctx)
		h.Index(w, r)
		return
	}
	lotID, _ := strconv.ParseInt(r.FormValue("lot_id"), 10, 64)
	if _, err := h.store.CreatePayable(store.CreatePayableInput{
		Description: r.FormValue("description"),
		AmountCents: amount,
		DueOn:       strings.TrimSpace(r.FormValue("due_on")),
		LotID:       lotID,
	}); err != nil {
		ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{"form": err.Error()})
		r = r.WithContext(ctx)
		h.Index(w, r)
		return
	}
	h.inertia.Redirect(w, r, "/payables", http.StatusSeeOther)
}

func (h *PayablesHandler) Cancel(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.store.CancelPayable(id); err != nil {
		ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{"form": err.Error()})
		r = r.WithContext(ctx)
		h.Index(w, r)
		return
	}
	h.inertia.Redirect(w, r, "/payables", http.StatusSeeOther)
}
