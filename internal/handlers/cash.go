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
	"github.com/puppe1990/leilao-erp/internal/models"
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
	h.renderIndex(w, r, accountFilter, nil)
}

func (h *CashHandler) renderIndex(w http.ResponseWriter, r *http.Request, accountFilter int64, ve inertia.ValidationErrors) {
	if ve != nil {
		r = r.WithContext(inertia.SetValidationErrors(r.Context(), ve))
	}

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
		entryRows = append(entryRows, cashEntryRow(e, accountNameByID[e.AccountID]))
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
		"categories":      manualCashCategoryOptions(),
	}))
}

func cashEntryRow(e models.CashEntry, accountName string) map[string]any {
	memo := ""
	if e.Memo != nil {
		memo = *e.Memo
	}
	manual := store.CashEntryIsManual(e)
	occurredDate := e.OccurredAt
	if len(occurredDate) >= 10 {
		occurredDate = occurredDate[:10]
	}
	return map[string]any{
		"id":             e.ID,
		"accountId":      e.AccountID,
		"accountName":    accountName,
		"direction":      e.Direction,
		"directionLabel": cashDirectionLabel(e.Direction),
		"amount":         domain.FormatBRL(e.AmountCents),
		"amountRaw":      formatCashInput(e.AmountCents),
		"occurredAt":     e.OccurredAt,
		"occurredDate":   occurredDate,
		"category":       e.Category,
		"canEdit":        manual,
		"canDelete":      manual,
		"categoryLabel":  cashCategoryLabel(e.Category),
		"memo":           memo,
	}
}

func manualCashCategoryOptions() []map[string]string {
	return []map[string]string{
		{"value": "despesa", "label": "Despesa"},
		{"value": "ajuste", "label": "Ajuste"},
		{"value": "frete", "label": "Frete"},
		{"value": "taxa", "label": "Taxa"},
	}
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
	category := strings.TrimSpace(r.FormValue("category"))
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
		h.renderIndex(w, r, 0, ve)
		return
	}

	occurredAtValue := occurredAt
	if len(occurredAt) == 10 {
		occurredAtValue = occurredAt + "T12:00:00Z"
	}

	if _, err := h.store.InsertManualCashEntry(accountID, direction, amountCents, occurredAtValue, category, memo); err != nil {
		ve["form"] = err.Error()
		h.renderIndex(w, r, 0, ve)
		return
	}

	h.inertia.Redirect(w, r, "/cash", http.StatusSeeOther)
}

func (h *CashHandler) UpdateEntry(w http.ResponseWriter, r *http.Request, id int64) {
	if err := parseFormOrJSON(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	accountID, _ := strconv.ParseInt(strings.TrimSpace(r.FormValue("account_id")), 10, 64)
	direction := strings.TrimSpace(r.FormValue("direction"))
	amountStr := strings.TrimSpace(r.FormValue("amount"))
	memo := strings.TrimSpace(r.FormValue("memo"))
	category := strings.TrimSpace(r.FormValue("category"))
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
		h.renderIndex(w, r, 0, ve)
		return
	}

	occurredAtValue := occurredAt
	if len(occurredAt) == 10 {
		occurredAtValue = occurredAt + "T12:00:00Z"
	}

	if err := h.store.UpdateCashEntry(id, accountID, direction, amountCents, occurredAtValue, category, memo); err != nil {
		ve["form"] = err.Error()
		h.renderIndex(w, r, 0, ve)
		return
	}

	h.inertia.Redirect(w, r, "/cash", http.StatusSeeOther)
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
	case "despesa":
		return "Despesa"
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
		h.renderIndex(w, r, 0, inertia.ValidationErrors{"name": "Nome obrigatório"})
		return
	}
	if _, err := h.store.InsertCashAccount(name, kind, opening); err != nil {
		h.renderIndex(w, r, 0, inertia.ValidationErrors{"form": err.Error()})
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
		h.renderIndex(w, r, 0, inertia.ValidationErrors{"form": err.Error()})
		return
	}
	h.inertia.Redirect(w, r, "/cash", http.StatusSeeOther)
}

func (h *CashHandler) DestroyAccount(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.store.DeleteCashAccount(id); err != nil {
		h.renderIndex(w, r, 0, inertia.ValidationErrors{"form": err.Error()})
		return
	}
	h.inertia.Redirect(w, r, "/cash", http.StatusSeeOther)
}

func (h *CashHandler) DestroyEntry(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.store.DeleteCashEntry(id); err != nil {
		h.renderIndex(w, r, 0, inertia.ValidationErrors{"form": err.Error()})
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
