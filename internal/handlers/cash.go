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

type CashHandler struct {
	renderer *cais.Renderer
	store    store.Store
	site     meta.Site
	cfg      cais.Config
	inertia  *inertia.Inertia
}

func NewCashHandler(renderer *cais.Renderer, s store.Store, site meta.Site, cfg cais.Config, i *inertia.Inertia) *CashHandler {
	return &CashHandler{renderer: renderer, store: s, site: site, cfg: cfg, inertia: i}
}

func (h *CashHandler) Index(w http.ResponseWriter, r *http.Request) {
	accountFilter, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("account_id")), 10, 64)

	accounts, err := h.store.ListCashAccounts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	accountNameByID := make(map[int64]string, len(accounts))
	balances := make([]map[string]any, 0, len(accounts))
	for _, a := range accounts {
		accountNameByID[a.ID] = a.Name
		bal, err := h.store.CashBalance(a.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		balances = append(balances, map[string]any{
			"id":           a.ID,
			"name":         a.Name,
			"kind":         a.Kind,
			"opening":      domain.FormatBRL(a.OpeningBalanceCents),
			"openingRaw":   formatCashInput(a.OpeningBalanceCents),
			"balance":      domain.FormatBRL(bal),
			"balanceCents": bal,
		})
	}

	entries, err := h.store.ListCashEntries(accountFilter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	entryRows := make([]map[string]any, 0, len(entries))
	for _, e := range entries {
		memo := ""
		if e.Memo != nil {
			memo = *e.Memo
		}
		entryRows = append(entryRows, map[string]any{
			"id":             e.ID,
			"accountId":      e.AccountID,
			"accountName":    accountNameByID[e.AccountID],
			"direction":      e.Direction,
			"directionLabel": cashDirectionLabel(e.Direction),
			"amount":         domain.FormatBRL(e.AmountCents),
			"occurredAt":     e.OccurredAt,
			"category":       e.Category,
			"canDelete":      e.Category == "ajuste",
			"categoryLabel":  cashCategoryLabel(e.Category),
			"memo":           memo,
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

	_ = h.inertia.Render(w, r, "Cash/Index", withCompany(h.store, inertia.Props{
		"site":            meta.ForRequest(h.site, r),
		"balances":        balances,
		"entries":         entryRows,
		"cashAccounts":    accountOptions,
		"filterAccountId": accountFilter,
	}))
}

func (h *CashHandler) CreateManual(w http.ResponseWriter, r *http.Request) {
	if err := parseFormOrJSON(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	accountID, _ := strconv.ParseInt(strings.TrimSpace(r.FormValue("account_id")), 10, 64)
	direction := strings.TrimSpace(r.FormValue("direction"))
	amountStr := strings.TrimSpace(r.FormValue("amount"))
	memo := strings.TrimSpace(r.FormValue("memo"))
	occurredAt := strings.TrimSpace(r.FormValue("occurred_at"))
	if occurredAt == "" {
		occurredAt = time.Now().UTC().Format("2006-01-02")
	}

	ve := make(inertia.ValidationErrors)
	if accountID <= 0 {
		ve["account_id"] = "Selecione a conta de caixa"
	}
	if direction != "in" && direction != "out" {
		ve["direction"] = "Selecione entrada ou saída"
	}
	amountCents, err := domain.ParseBRLToCents(amountStr)
	if err != nil || amountCents <= 0 {
		ve["amount"] = "Valor inválido"
	}

	if len(ve) > 0 {
		h.renderIndexWithErrors(w, r, ve)
		return
	}

	occurredAtValue := occurredAt
	if len(occurredAt) == 10 {
		occurredAtValue = occurredAt + "T12:00:00Z"
	}

	// category is fixed to "ajuste" inside the store method
	if _, err := h.store.InsertManualCashEntry(accountID, direction, amountCents, occurredAtValue, memo); err != nil {
		ve["form"] = err.Error()
		h.renderIndexWithErrors(w, r, ve)
		return
	}

	h.inertia.Redirect(w, r, "/cash", http.StatusSeeOther)
}

func (h *CashHandler) renderIndexWithErrors(w http.ResponseWriter, r *http.Request, ve inertia.ValidationErrors) {
	ctx := inertia.SetValidationErrors(r.Context(), ve)
	// Re-render Index with same props as Index (without filter for simplicity on error)
	accounts, err := h.store.ListCashAccounts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	accountNameByID := make(map[int64]string, len(accounts))
	balances := make([]map[string]any, 0, len(accounts))
	for _, a := range accounts {
		accountNameByID[a.ID] = a.Name
		bal, err := h.store.CashBalance(a.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		balances = append(balances, map[string]any{
			"id":           a.ID,
			"name":         a.Name,
			"kind":         a.Kind,
			"opening":      domain.FormatBRL(a.OpeningBalanceCents),
			"openingRaw":   formatCashInput(a.OpeningBalanceCents),
			"balance":      domain.FormatBRL(bal),
			"balanceCents": bal,
		})
	}
	entries, err := h.store.ListCashEntries(0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	entryRows := make([]map[string]any, 0, len(entries))
	for _, e := range entries {
		memo := ""
		if e.Memo != nil {
			memo = *e.Memo
		}
		entryRows = append(entryRows, map[string]any{
			"id":             e.ID,
			"accountId":      e.AccountID,
			"accountName":    accountNameByID[e.AccountID],
			"direction":      e.Direction,
			"directionLabel": cashDirectionLabel(e.Direction),
			"amount":         domain.FormatBRL(e.AmountCents),
			"occurredAt":     e.OccurredAt,
			"category":       e.Category,
			"canDelete":      e.Category == "ajuste",
			"categoryLabel":  cashCategoryLabel(e.Category),
			"memo":           memo,
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
	_ = h.inertia.Render(w, r.WithContext(ctx), "Cash/Index", withCompany(h.store, inertia.Props{
		"site":            meta.ForRequest(h.site, r),
		"balances":        balances,
		"entries":         entryRows,
		"cashAccounts":    accountOptions,
		"filterAccountId": int64(0),
	}))
}

func cashDirectionLabel(direction string) string {
	switch direction {
	case "in":
		return "Entrada"
	case "out":
		return "Saída"
	default:
		return direction
	}
}

func cashCategoryLabel(category string) string {
	switch category {
	case "ajuste":
		return "Ajuste"
	case "pagamento":
		return "Pagamento"
	case "recebimento":
		return "Recebimento"
	case "compra_lote":
		return "Compra de lote"
	case "venda":
		return "Venda"
	case "taxa":
		return "Taxa"
	case "frete":
		return "Frete"
	default:
		return category
	}
}

func (h *CashHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	if err := parseFormOrJSON(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	kind := strings.TrimSpace(r.FormValue("kind"))
	if kind == "" {
		kind = "pix"
	}
	opening, _ := domain.ParseBRLToCents(r.FormValue("opening_balance"))
	if name == "" {
		ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{"name": "Nome obrigatório"})
		r = r.WithContext(ctx)
		h.Index(w, r)
		return
	}
	if _, err := h.store.InsertCashAccount(name, kind, opening); err != nil {
		ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{"form": err.Error()})
		r = r.WithContext(ctx)
		h.Index(w, r)
		return
	}
	h.inertia.Redirect(w, r, "/cash", http.StatusSeeOther)
}

func (h *CashHandler) UpdateAccount(w http.ResponseWriter, r *http.Request, id int64) {
	if err := parseFormOrJSON(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	kind := strings.TrimSpace(r.FormValue("kind"))
	opening, _ := domain.ParseBRLToCents(r.FormValue("opening_balance"))
	if err := h.store.UpdateCashAccount(id, name, kind, opening); err != nil {
		ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{"form": err.Error()})
		r = r.WithContext(ctx)
		h.Index(w, r)
		return
	}
	h.inertia.Redirect(w, r, "/cash", http.StatusSeeOther)
}

func (h *CashHandler) DestroyAccount(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.store.DeleteCashAccount(id); err != nil {
		ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{"form": err.Error()})
		r = r.WithContext(ctx)
		h.Index(w, r)
		return
	}
	h.inertia.Redirect(w, r, "/cash", http.StatusSeeOther)
}

func (h *CashHandler) DestroyEntry(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.store.DeleteCashEntry(id); err != nil {
		ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{"form": err.Error()})
		r = r.WithContext(ctx)
		h.Index(w, r)
		return
	}
	h.inertia.Redirect(w, r, "/cash", http.StatusSeeOther)
}

func formatCashInput(cents int64) string {
	neg := cents < 0
	if neg {
		cents = -cents
	}
	s := fmt.Sprintf("%d,%02d", cents/100, cents%100)
	if neg {
		return "-" + s
	}
	return s
}
