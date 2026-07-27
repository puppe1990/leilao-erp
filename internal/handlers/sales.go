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

type SalesHandler struct {
	renderer *cais.Renderer
	store    store.Store
	site     meta.Site
	cfg      cais.Config
	inertia  *inertia.Inertia
}

func NewSalesHandler(renderer *cais.Renderer, s store.Store, site meta.Site, cfg cais.Config, i *inertia.Inertia) *SalesHandler {
	return &SalesHandler{renderer: renderer, store: s, site: site, cfg: cfg, inertia: i}
}

func (h *SalesHandler) Index(w http.ResponseWriter, r *http.Request) {
	sales, err := h.store.ListSales()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows := make([]map[string]any, 0, len(sales))
	for _, sale := range sales {
		rows = append(rows, map[string]any{
			"id":            sale.ID,
			"itemId":        sale.ItemID,
			"itemTitle":     sale.ItemTitle,
			"soldAt":        sale.SoldAt,
			"channel":       sale.Channel,
			"channelLabel":  channelLabel(sale.Channel),
			"gross":         domain.FormatBRL(sale.GrossCents),
			"fee":           domain.FormatBRL(sale.FeeCents),
			"shipping":      domain.FormatBRL(sale.ShippingCents),
			"net":           domain.FormatBRL(sale.NetCents),
			"paymentStatus": sale.PaymentStatus,
			"paymentLabel":  paymentStatusLabel(sale.PaymentStatus),
			"canCancel":     sale.PaymentStatus == "pending",
		})
	}

	_ = h.inertia.Render(w, r, "Sales/Index", withCompany(h.store, inertia.Props{
		"site":  meta.ForRequest(h.site, r),
		"sales": rows,
	}))
}

func (h *SalesHandler) New(w http.ResponseWriter, r *http.Request) {
	items, err := h.inStockItemProps()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	accounts, err := h.cashAccountProps()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = h.inertia.Render(w, r, "Sales/New", withCompany(h.store, inertia.Props{
		"site":         meta.ForRequest(h.site, r),
		"items":        items,
		"cashAccounts": accounts,
		"channels":     channelOptions(),
	}))
}

func (h *SalesHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := parseFormOrJSON(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	itemID, _ := strconv.ParseInt(strings.TrimSpace(r.FormValue("item_id")), 10, 64)
	channel := strings.TrimSpace(r.FormValue("channel"))
	grossStr := strings.TrimSpace(r.FormValue("gross"))
	feeStr := strings.TrimSpace(r.FormValue("fee"))
	shippingStr := strings.TrimSpace(r.FormValue("shipping"))
	paymentStatus := strings.TrimSpace(r.FormValue("payment_status"))
	cashAccountID, _ := strconv.ParseInt(strings.TrimSpace(r.FormValue("cash_account_id")), 10, 64)
	dueOn := strings.TrimSpace(r.FormValue("due_on"))
	soldAt := strings.TrimSpace(r.FormValue("sold_at"))
	if soldAt == "" {
		soldAt = time.Now().UTC().Format("2006-01-02")
	}

	ve := make(inertia.ValidationErrors)
	if itemID <= 0 {
		ve["item_id"] = "Selecione um item"
	}
	if !validChannel(channel) {
		ve["channel"] = "Canal inválido"
	}

	grossCents, err := domain.ParseBRLToCents(grossStr)
	if err != nil || grossCents < 0 {
		ve["gross"] = "Valor bruto inválido"
	}
	feeCents := int64(0)
	if feeStr != "" {
		feeCents, err = domain.ParseBRLToCents(feeStr)
		if err != nil || feeCents < 0 {
			ve["fee"] = "Taxa inválida"
		}
	}
	shippingCents := int64(0)
	if shippingStr != "" {
		shippingCents, err = domain.ParseBRLToCents(shippingStr)
		if err != nil || shippingCents < 0 {
			ve["shipping"] = "Frete inválido"
		}
	}

	switch paymentStatus {
	case "received":
		if cashAccountID <= 0 {
			ve["cash_account_id"] = "Selecione a conta de caixa"
		}
	case "pending":
		if dueOn == "" {
			ve["due_on"] = "Informe a data de vencimento"
		}
	default:
		ve["payment_status"] = "Selecione o status do pagamento"
	}

	if len(ve) > 0 {
		h.renderNewWithErrors(w, r, ve)
		return
	}

	// sold_at: accept date or datetime; store as datetime for cash_entries
	soldAtValue := soldAt
	if len(soldAt) == 10 {
		soldAtValue = soldAt + "T12:00:00Z"
	}

	_, err = h.store.CreateSale(store.CreateSaleInput{
		ItemID:        itemID,
		SoldAt:        soldAtValue,
		Channel:       channel,
		GrossCents:    grossCents,
		FeeCents:      feeCents,
		ShippingCents: shippingCents,
		PaymentStatus: paymentStatus,
		CashAccountID: cashAccountID,
		DueOn:         dueOn,
	})
	if err != nil {
		// Map common store errors to field errors
		msg := err.Error()
		if strings.Contains(msg, "not in stock") || strings.Contains(msg, "not found") {
			ve["item_id"] = msg
		} else {
			ve["form"] = msg
		}
		h.renderNewWithErrors(w, r, ve)
		return
	}

	h.inertia.Redirect(w, r, "/sales", http.StatusSeeOther)
}

func (h *SalesHandler) Cancel(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.store.CancelPendingSale(id); err != nil {
		// Re-render index with flash-like form error
		sales, listErr := h.store.ListSales()
		if listErr != nil {
			http.Error(w, listErr.Error(), http.StatusInternalServerError)
			return
		}
		rows := make([]map[string]any, 0, len(sales))
		for _, sale := range sales {
			rows = append(rows, map[string]any{
				"id":            sale.ID,
				"itemId":        sale.ItemID,
				"itemTitle":     sale.ItemTitle,
				"soldAt":        sale.SoldAt,
				"channel":       sale.Channel,
				"channelLabel":  channelLabel(sale.Channel),
				"gross":         domain.FormatBRL(sale.GrossCents),
				"fee":           domain.FormatBRL(sale.FeeCents),
				"shipping":      domain.FormatBRL(sale.ShippingCents),
				"net":           domain.FormatBRL(sale.NetCents),
				"paymentStatus": sale.PaymentStatus,
				"paymentLabel":  paymentStatusLabel(sale.PaymentStatus),
				"canCancel":     sale.PaymentStatus == "pending",
			})
		}
		ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{"form": err.Error()})
		_ = h.inertia.Render(w, r.WithContext(ctx), "Sales/Index", withCompany(h.store, inertia.Props{
			"site":  meta.ForRequest(h.site, r),
			"sales": rows,
		}))
		return
	}
	h.inertia.Redirect(w, r, "/sales", http.StatusSeeOther)
}

func (h *SalesHandler) renderNewWithErrors(w http.ResponseWriter, r *http.Request, ve inertia.ValidationErrors) {
	items, err := h.inStockItemProps()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	accounts, err := h.cashAccountProps()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ctx := inertia.SetValidationErrors(r.Context(), ve)
	_ = h.inertia.Render(w, r.WithContext(ctx), "Sales/New", withCompany(h.store, inertia.Props{
		"site":         meta.ForRequest(h.site, r),
		"items":        items,
		"cashAccounts": accounts,
		"channels":     channelOptions(),
	}))
}

func (h *SalesHandler) inStockItemProps() ([]map[string]any, error) {
	items, err := h.store.ListItemsInStock()
	if err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(items))
	for _, it := range items {
		out = append(out, map[string]any{
			"id":       it.ID,
			"title":    it.Title,
			"lotId":    it.LotID,
			"unitCost": domain.FormatBRL(it.UnitCostCents),
		})
	}
	return out, nil
}

func (h *SalesHandler) cashAccountProps() ([]map[string]any, error) {
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

func channelLabel(channel string) string {
	switch channel {
	case "direct":
		return "Direto"
	case "mercadolivre":
		return "Mercado Livre"
	case "shopee":
		return "Shopee"
	case "olx":
		return "OLX"
	case "other":
		return "Outro"
	default:
		return channel
	}
}

func channelOptions() []map[string]string {
	return []map[string]string{
		{"value": "direct", "label": "Direto"},
		{"value": "mercadolivre", "label": "Mercado Livre"},
		{"value": "shopee", "label": "Shopee"},
		{"value": "olx", "label": "OLX"},
		{"value": "other", "label": "Outro"},
	}
}

func validChannel(channel string) bool {
	switch channel {
	case "direct", "mercadolivre", "shopee", "olx", "other":
		return true
	default:
		return false
	}
}

func paymentStatusLabel(status string) string {
	switch status {
	case "received":
		return "Recebido"
	case "pending":
		return "A receber"
	case "cancelled":
		return "Cancelado"
	default:
		return status
	}
}
