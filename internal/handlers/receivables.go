package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/meta"
	inertia "github.com/romsar/gonertia/v3"

	"github.com/puppe1990/leilao-erp/internal/domain"
	"github.com/puppe1990/leilao-erp/internal/store"
)

type ReceivablesHandler struct {
	renderer *cais.Renderer
	store    store.Store
	site     meta.Site
	cfg      cais.Config
	inertia  *inertia.Inertia
}

func NewReceivablesHandler(renderer *cais.Renderer, s store.Store, site meta.Site, cfg cais.Config, i *inertia.Inertia) *ReceivablesHandler {
	return &ReceivablesHandler{renderer: renderer, store: s, site: site, cfg: cfg, inertia: i}
}

func (h *ReceivablesHandler) Index(w http.ResponseWriter, r *http.Request) {
	receivables, err := h.store.ListReceivables()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	accounts, err := h.store.ListCashAccounts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows := make([]map[string]any, 0, len(receivables))
	for _, rec := range receivables {
		receivedAt := ""
		if rec.ReceivedAt != nil {
			receivedAt = *rec.ReceivedAt
		}
		var saleID any
		if rec.SaleID != nil {
			saleID = *rec.SaleID
		}
		open := rec.Status == "open"
		canDelete := open && rec.SaleID == nil
		rows = append(rows, map[string]any{
			"id":          rec.ID,
			"description": rec.Description,
			"amount":      domain.FormatBRL(rec.AmountCents),
			"amountRaw":   formatCashInput(rec.AmountCents),
			"dueOn":       rec.DueOn,
			"status":      rec.Status,
			"statusLabel": receivableStatusLabel(rec.Status),
			"saleId":      saleID,
			"receivedAt":  receivedAt,
			"canCancel":   open,
			"canSettle":   open,
			"canEdit":     open,
			"canDelete":   canDelete,
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

	_ = h.inertia.Render(w, r, "Receivables/Index", withCompany(h.store, inertia.Props{
		"site":         meta.ForRequest(h.site, r),
		"receivables":  rows,
		"cashAccounts": accountOptions,
	}))
}

func (h *ReceivablesHandler) Settle(w http.ResponseWriter, r *http.Request, id int64) {
	if err := parseFormOrJSON(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	accountID, _ := strconv.ParseInt(strings.TrimSpace(r.FormValue("cash_account_id")), 10, 64)
	receivedAt := strings.TrimSpace(r.FormValue("received_at"))
	if receivedAt == "" {
		receivedAt = time.Now().UTC().Format("2006-01-02")
	}
	if len(receivedAt) == 10 {
		receivedAt += "T12:00:00Z"
	}

	if accountID <= 0 {
		accounts, err := h.store.ListCashAccounts()
		if err == nil && len(accounts) == 1 {
			accountID = accounts[0].ID
		}
	}

	if accountID <= 0 {
		ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{
			"cash_account_id": "Selecione a conta de caixa",
		})
		r = r.WithContext(ctx)
		h.Index(w, r)
		return
	}

	if err := h.store.SettleReceivable(id, accountID, receivedAt); err != nil {
		ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{
			"form": err.Error(),
		})
		r = r.WithContext(ctx)
		h.Index(w, r)
		return
	}

	h.inertia.Redirect(w, r, "/receivables", http.StatusSeeOther)
}

func receivableStatusLabel(status string) string {
	switch status {
	case "open":
		return "Aberto"
	case "received":
		return "Recebido"
	case "cancelled":
		return "Cancelado"
	default:
		return status
	}
}

func (h *ReceivablesHandler) Create(w http.ResponseWriter, r *http.Request) {
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
	if _, err := h.store.CreateReceivable(store.CreateReceivableInput{
		Description: r.FormValue("description"),
		AmountCents: amount,
		DueOn:       strings.TrimSpace(r.FormValue("due_on")),
	}); err != nil {
		ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{"form": err.Error()})
		r = r.WithContext(ctx)
		h.Index(w, r)
		return
	}
	h.inertia.Redirect(w, r, "/receivables", http.StatusSeeOther)
}

func (h *ReceivablesHandler) Cancel(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.store.CancelReceivable(id); err != nil {
		ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{"form": err.Error()})
		r = r.WithContext(ctx)
		h.Index(w, r)
		return
	}
	h.inertia.Redirect(w, r, "/receivables", http.StatusSeeOther)
}

func (h *ReceivablesHandler) Update(w http.ResponseWriter, r *http.Request, id int64) {
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
	if err := h.store.UpdateReceivable(id, store.CreateReceivableInput{
		Description: r.FormValue("description"),
		AmountCents: amount,
		DueOn:       strings.TrimSpace(r.FormValue("due_on")),
	}); err != nil {
		ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{"form": err.Error()})
		r = r.WithContext(ctx)
		h.Index(w, r)
		return
	}
	h.inertia.Redirect(w, r, "/receivables", http.StatusSeeOther)
}

func (h *ReceivablesHandler) Destroy(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.store.DeleteReceivable(id); err != nil {
		ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{"form": err.Error()})
		r = r.WithContext(ctx)
		h.Index(w, r)
		return
	}
	h.inertia.Redirect(w, r, "/receivables", http.StatusSeeOther)
}
