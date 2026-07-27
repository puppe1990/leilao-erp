package handlers

import (
	"fmt"
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
		title := sale.ItemTitle
		if sale.Composition != "" {
			title = sale.Composition
		}
		rows = append(rows, map[string]any{
			"id":            sale.ID,
			"itemId":        sale.ItemID,
			"itemTitle":     title,
			"lineCount":     sale.LineCount,
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
	accessoryIDs := parseAccessoryIDs(r)
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
		AccessoryIDs:  accessoryIDs,
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
			if strings.Contains(msg, "accessory") || strings.Contains(strings.ToLower(msg), "cabo") {
				ve["accessory_ids"] = msg
			} else {
				ve["item_id"] = msg
			}
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
		saleHint := ""
		saleHintRaw := int64(0)
		if it.SalePriceHintCents != nil {
			saleHint = domain.FormatBRL(*it.SalePriceHintCents)
			saleHintRaw = *it.SalePriceHintCents
		}
		out = append(out, map[string]any{
			"id":            it.ID,
			"title":         it.Title,
			"lotId":         it.LotID,
			"unitCost":      domain.FormatBRL(it.UnitCostCents),
			"unitCostRaw":   it.UnitCostCents,
			"salePriceHint": saleHint,
			"salePriceRaw":  saleHintRaw,
			"isAccessory":   isAccessoryTitle(it.Title),
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

func (h *SalesHandler) Show(w http.ResponseWriter, r *http.Request, id int64) {
	sale, err := h.store.FindSaleByID(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	lines, _ := h.store.ListSaleLines(id)
	lineProps := make([]map[string]any, 0, len(lines))
	for _, ln := range lines {
		roleLabel := "Principal"
		if ln.Role == "accessory" {
			roleLabel = "Acessório"
		}
		lineProps = append(lineProps, map[string]any{
			"id":        ln.ID,
			"itemId":    ln.ItemID,
			"title":     ln.ItemTitle,
			"role":      ln.Role,
			"roleLabel": roleLabel,
			"unitCost":  domain.FormatBRL(ln.UnitCostCentsAtSale),
		})
	}
	title := sale.ItemTitle
	if sale.LineCount > 1 || len(lines) > 1 {
		n := sale.LineCount
		if n == 0 {
			n = len(lines)
		}
		if n > 1 {
			title = fmt.Sprintf("%s + %d acessório(s)", sale.ItemTitle, n-1)
		}
	}
	margin := domain.Margin(sale.NetCents, sale.UnitCostCentsAtSale)
	_ = h.inertia.Render(w, r, "Sales/Show", withCompany(h.store, inertia.Props{
		"site": meta.ForRequest(h.site, r),
		"sale": map[string]any{
			"id":            sale.ID,
			"itemId":        sale.ItemID,
			"itemTitle":     title,
			"soldAt":        sale.SoldAt,
			"channel":       sale.Channel,
			"channelLabel":  channelLabel(sale.Channel),
			"gross":         domain.FormatBRL(sale.GrossCents),
			"fee":           domain.FormatBRL(sale.FeeCents),
			"shipping":      domain.FormatBRL(sale.ShippingCents),
			"net":           domain.FormatBRL(sale.NetCents),
			"grossRaw":      sale.GrossCents,
			"feeRaw":        sale.FeeCents,
			"shippingRaw":   sale.ShippingCents,
			"paymentStatus": sale.PaymentStatus,
			"paymentLabel":  paymentStatusLabel(sale.PaymentStatus),
			"canEdit":       sale.PaymentStatus == "pending",
			"canDelete":     sale.PaymentStatus == "pending",
			"unitCost":      domain.FormatBRL(sale.UnitCostCentsAtSale),
			"margin":        domain.FormatBRL(margin),
			"lines":         lineProps,
		},
	}))
}

func (h *SalesHandler) Edit(w http.ResponseWriter, r *http.Request, id int64) {
	sale, err := h.store.FindSaleByID(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if sale.PaymentStatus != "pending" {
		h.inertia.Redirect(w, r, fmt.Sprintf("/sales/%d", id), http.StatusSeeOther)
		return
	}
	dueOn := ""
	recs, _ := h.store.ListReceivables()
	for _, rec := range recs {
		if rec.SaleID != nil && *rec.SaleID == id && rec.Status == "open" {
			dueOn = rec.DueOn
			break
		}
	}
	_ = h.inertia.Render(w, r, "Sales/Edit", withCompany(h.store, inertia.Props{
		"site":     meta.ForRequest(h.site, r),
		"channels": channelOptions(),
		"sale": map[string]any{
			"id":        sale.ID,
			"itemTitle": sale.ItemTitle,
			"soldAt":    sale.SoldAt[:minInt(10, len(sale.SoldAt))],
			"channel":   sale.Channel,
			"gross":     formatCentsInput(sale.GrossCents),
			"fee":       formatCentsInput(sale.FeeCents),
			"shipping":  formatCentsInput(sale.ShippingCents),
			"dueOn":     dueOn,
		},
	}))
}

func (h *SalesHandler) Update(w http.ResponseWriter, r *http.Request, id int64) {
	if err := parseFormOrJSON(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	gross, err1 := domain.ParseBRLToCents(r.FormValue("gross"))
	fee, errFee := domain.ParseBRLToCents(r.FormValue("fee"))
	if errFee != nil {
		fee = 0
	}
	ship, errShip := domain.ParseBRLToCents(r.FormValue("shipping"))
	if errShip != nil {
		ship = 0
	}
	if err1 != nil {
		ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{"gross": "Valor inválido"})
		r = r.WithContext(ctx)
		h.Edit(w, r, id)
		return
	}
	err := h.store.UpdateSale(id, store.UpdateSaleInput{
		SoldAt:        strings.TrimSpace(r.FormValue("sold_at")),
		Channel:       strings.TrimSpace(r.FormValue("channel")),
		GrossCents:    gross,
		FeeCents:      fee,
		ShippingCents: ship,
		DueOn:         strings.TrimSpace(r.FormValue("due_on")),
	})
	if err != nil {
		ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{"form": err.Error()})
		r = r.WithContext(ctx)
		h.Edit(w, r, id)
		return
	}
	h.inertia.Redirect(w, r, fmt.Sprintf("/sales/%d", id), http.StatusSeeOther)
}

func (h *SalesHandler) Destroy(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.store.DeleteSale(id); err != nil {
		ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{"form": err.Error()})
		r = r.WithContext(ctx)
		h.Index(w, r)
		return
	}
	h.inertia.Redirect(w, r, "/sales", http.StatusSeeOther)
}
