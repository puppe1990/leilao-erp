package handlers

import (
	"fmt"
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

type LotsHandler struct {
	renderer *cais.Renderer
	store    store.Store
	site     meta.Site
	cfg      cais.Config
	inertia  *inertia.Inertia
}

func NewLotsHandler(renderer *cais.Renderer, s store.Store, site meta.Site, cfg cais.Config, i *inertia.Inertia) *LotsHandler {
	return &LotsHandler{renderer: renderer, store: s, site: site, cfg: cfg, inertia: i}
}

func (h *LotsHandler) Index(w http.ResponseWriter, r *http.Request) {
	lots, err := h.store.ListLots()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows := make([]map[string]any, 0, len(lots))
	for _, lot := range lots {
		rows = append(rows, map[string]any{
			"id":          lot.ID,
			"name":        lot.Name,
			"purchasedAt": lot.PurchasedAt,
			"status":      lot.Status,
			"statusLabel": lotStatusLabel(lot.Status),
			"totalCost":   domain.FormatBRL(lot.TotalCostCents),
			"itemCount":   lot.ItemCount,
		})
	}

	_ = h.inertia.Render(w, r, "Lots/Index", inertia.Props{
		"site": meta.ForRequest(h.site, r),
		"lots": rows,
	})
}

func (h *LotsHandler) New(w http.ResponseWriter, r *http.Request) {
	accounts, err := h.cashAccountProps()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = h.inertia.Render(w, r, "Lots/New", inertia.Props{
		"site":         meta.ForRequest(h.site, r),
		"cashAccounts": accounts,
	})
}

func (h *LotsHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := parseFormOrJSON(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	purchasedAt := strings.TrimSpace(r.FormValue("purchased_at"))
	itemTitle := strings.TrimSpace(r.FormValue("item_title"))
	qtyStr := strings.TrimSpace(r.FormValue("item_qty"))
	costLabel := strings.TrimSpace(r.FormValue("cost_label"))
	if costLabel == "" {
		costLabel = "Arremate"
	}
	costAmount := strings.TrimSpace(r.FormValue("cost_amount"))
	alreadyPaid := formBool(r.FormValue("already_paid"))
	cashAccountID, _ := strconv.ParseInt(strings.TrimSpace(r.FormValue("cash_account_id")), 10, 64)

	ve := make(inertia.ValidationErrors)
	if name == "" {
		ve["name"] = "Nome é obrigatório"
	}
	if purchasedAt == "" {
		ve["purchased_at"] = "Data da compra é obrigatória"
	}
	if itemTitle == "" {
		ve["item_title"] = "Título do item é obrigatório"
	}
	qty, err := strconv.Atoi(qtyStr)
	if err != nil || qty <= 0 {
		ve["item_qty"] = "Quantidade deve ser maior que zero"
	}
	amountCents, err := domain.ParseBRLToCents(costAmount)
	if err != nil || amountCents <= 0 {
		ve["cost_amount"] = "Valor do arremate inválido"
	}
	if alreadyPaid && cashAccountID <= 0 {
		ve["cash_account_id"] = "Selecione a conta de caixa"
	}

	if len(ve) > 0 {
		h.renderNewWithErrors(w, r, ve)
		return
	}

	paidAt := ""
	if alreadyPaid {
		paidAt = purchasedAt + "T12:00:00Z"
	}

	lotID, err := h.store.CreateLotPurchase(store.CreateLotInput{
		Name:        name,
		PurchasedAt: purchasedAt,
		ItemTitle:   itemTitle,
		ItemQty:     qty,
		Costs: []store.CostInput{
			{Label: costLabel, AmountCents: amountCents, AlreadyPaid: alreadyPaid},
		},
		CashAccountID: cashAccountID,
		PaidAt:        paidAt,
	})
	if err != nil {
		ve["form"] = err.Error()
		h.renderNewWithErrors(w, r, ve)
		return
	}

	h.inertia.Redirect(w, r, fmt.Sprintf("/lots/%d", lotID), http.StatusSeeOther)
}

func (h *LotsHandler) Show(w http.ResponseWriter, r *http.Request, id int64) {
	lot, err := h.store.FindLot(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	items, err := h.store.ListItemsByLot(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	costs, err := h.store.ListPurchaseCostsByLot(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	payables, err := h.store.ListPayablesByLot(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	accounts, err := h.cashAccountProps()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var totalCost int64
	costRows := make([]map[string]any, 0, len(costs))
	for _, c := range costs {
		totalCost += c.AmountCents
		costRows = append(costRows, map[string]any{
			"id":     c.ID,
			"label":  c.Label,
			"amount": domain.FormatBRL(c.AmountCents),
		})
	}

	itemRows := make([]map[string]any, 0, len(items))
	for _, it := range items {
		itemRows = append(itemRows, map[string]any{
			"id":       it.ID,
			"title":    it.Title,
			"unitCost": domain.FormatBRL(it.UnitCostCents),
			"status":   it.Status,
			"statusLabel": itemStatusLabel(it.Status),
		})
	}

	payableRows := make([]map[string]any, 0, len(payables))
	for _, p := range payables {
		payableRows = append(payableRows, map[string]any{
			"id":          p.ID,
			"description": p.Description,
			"amount":      domain.FormatBRL(p.AmountCents),
			"status":      p.Status,
			"statusLabel": payableStatusLabel(p.Status),
			"dueOn":       p.DueOn,
		})
	}

	auctionSource := ""
	if lot.AuctionSource != nil {
		auctionSource = *lot.AuctionSource
	}
	notes := ""
	if lot.Notes != nil {
		notes = *lot.Notes
	}

	_ = h.inertia.Render(w, r, "Lots/Show", inertia.Props{
		"site": meta.ForRequest(h.site, r),
		"lot": map[string]any{
			"id":            lot.ID,
			"name":          lot.Name,
			"purchasedAt":   lot.PurchasedAt,
			"status":        lot.Status,
			"statusLabel":   lotStatusLabel(lot.Status),
			"auctionSource": auctionSource,
			"notes":         notes,
			"totalCost":     domain.FormatBRL(totalCost),
		},
		"items":        itemRows,
		"costs":        costRows,
		"payables":     payableRows,
		"cashAccounts": accounts,
	})
}

func (h *LotsHandler) AddCost(w http.ResponseWriter, r *http.Request, id int64) {
	if err := parseFormOrJSON(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if _, err := h.store.FindLot(id); err != nil {
		http.NotFound(w, r)
		return
	}

	label := strings.TrimSpace(r.FormValue("cost_label"))
	if label == "" {
		label = "Custo extra"
	}
	costAmount := strings.TrimSpace(r.FormValue("cost_amount"))
	alreadyPaid := formBool(r.FormValue("already_paid"))
	cashAccountID, _ := strconv.ParseInt(strings.TrimSpace(r.FormValue("cash_account_id")), 10, 64)

	ve := make(inertia.ValidationErrors)
	amountCents, err := domain.ParseBRLToCents(costAmount)
	if err != nil || amountCents <= 0 {
		ve["cost_amount"] = "Valor inválido"
	}
	if alreadyPaid && cashAccountID <= 0 {
		ve["cash_account_id"] = "Selecione a conta de caixa"
	}
	if len(ve) > 0 {
		ctx := inertia.SetValidationErrors(r.Context(), ve)
		// re-show detail with errors
		h.Show(w, r.WithContext(ctx), id)
		return
	}

	paidAt := ""
	if alreadyPaid {
		paidAt = time.Now().UTC().Format(time.RFC3339)
	}

	if err := h.store.AddPurchaseCost(id, store.CostInput{
		Label:       label,
		AmountCents: amountCents,
		AlreadyPaid: alreadyPaid,
	}, cashAccountID, paidAt); err != nil {
		ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{"form": err.Error()})
		h.Show(w, r.WithContext(ctx), id)
		return
	}

	h.inertia.Redirect(w, r, fmt.Sprintf("/lots/%d", id), http.StatusSeeOther)
}

func (h *LotsHandler) renderNewWithErrors(w http.ResponseWriter, r *http.Request, ve inertia.ValidationErrors) {
	accounts, err := h.cashAccountProps()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ctx := inertia.SetValidationErrors(r.Context(), ve)
	// gonertia Render always writes 200; validation surfaces via props.errors
	_ = h.inertia.Render(w, r.WithContext(ctx), "Lots/New", inertia.Props{
		"site":         meta.ForRequest(h.site, r),
		"cashAccounts": accounts,
	})
}

func (h *LotsHandler) cashAccountProps() ([]map[string]any, error) {
	accounts, err := h.store.ListCashAccounts()
	if err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(accounts))
	for _, a := range accounts {
		out = append(out, map[string]any{
			"id":   a.ID,
			"name": a.Name,
			"kind": a.Kind,
		})
	}
	return out, nil
}

func formBool(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "on", "yes", "sim":
		return true
	default:
		return false
	}
}

func lotStatusLabel(status string) string {
	switch status {
	case "open":
		return "Aberto"
	case "closed":
		return "Fechado"
	default:
		return status
	}
}

func itemStatusLabel(status string) string {
	switch status {
	case "in_stock":
		return "Em estoque"
	case "reserved":
		return "Reservado"
	case "sold":
		return "Vendido"
	default:
		return status
	}
}

func payableStatusLabel(status string) string {
	switch status {
	case "open":
		return "Em aberto"
	case "paid":
		return "Pago"
	default:
		return status
	}
}
