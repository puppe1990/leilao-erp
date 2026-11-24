package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/meta"
	inertia "github.com/romsar/gonertia/v3"

	"github.com/puppe1990/leilao-erp/internal/domain"
	"github.com/puppe1990/leilao-erp/internal/models"
	"github.com/puppe1990/leilao-erp/internal/store"
)

type SuppliersHandler struct {
	renderer *cais.Renderer
	store    store.Store
	site     meta.Site
	cfg      cais.Config
	inertia  *inertia.Inertia
}

func NewSuppliersHandler(renderer *cais.Renderer, s store.Store, site meta.Site, cfg cais.Config, i *inertia.Inertia) *SuppliersHandler {
	return &SuppliersHandler{renderer: renderer, store: s, site: site, cfg: cfg, inertia: i}
}

func supplierRow(item store.SupplierListItem) map[string]any {
	return map[string]any{
		"id":            item.ID,
		"name":          item.Name,
		"whatsapp":      item.WhatsApp,
		"whatsappLabel": formatSupplierWhatsApp(item.WhatsApp),
		"chatId":        item.ChatID,
		"notes":         item.Notes,
		"minOrder":      domain.FormatBRL(item.MinOrderCents),
		"minOrderRaw":   formatCashInput(item.MinOrderCents),
		"lastBody":      item.LastBody,
		"lastFromMe":    item.LastFromMe,
		"lastAt":        formatSupplierWhen(item.LastAt),
		"messageCount":  item.MessageCount,
	}
}

func supplierShow(s models.Supplier) map[string]any {
	return map[string]any{
		"id":            s.ID,
		"name":          s.Name,
		"whatsapp":      s.WhatsApp,
		"whatsappLabel": formatSupplierWhatsApp(s.WhatsApp),
		"chatId":        s.ChatID,
		"notes":         s.Notes,
		"minOrder":      domain.FormatBRL(s.MinOrderCents),
		"minOrderRaw":   formatCashInput(s.MinOrderCents),
	}
}

func supplierMessageRow(m models.SupplierMessage) map[string]any {
	return map[string]any{
		"id":     m.ID,
		"fromMe": m.FromMe,
		"body":   m.Body,
		"sentAt": formatSupplierWhen(m.SentAt),
	}
}

func formatSupplierWhatsApp(digits string) string {
	d := domain.NormalizeWhatsAppPhone(digits)
	if strings.HasPrefix(d, "55") && len(d) == 13 {
		return "(" + d[2:4] + ") " + d[4:9] + "-" + d[9:]
	}
	return digits
}

func formatSupplierWhen(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		loc = time.FixedZone("BRT", -3*3600)
	}
	return t.In(loc).Format("02/01 15:04")
}

func supplierInputFromRequest(r *http.Request) (store.SupplierInput, inertia.ValidationErrors) {
	errs := inertia.ValidationErrors{}
	in := store.SupplierInput{
		Name:     r.FormValue("name"),
		WhatsApp: r.FormValue("whatsapp"),
		Notes:    r.FormValue("notes"),
	}
	rawMin := strings.TrimSpace(r.FormValue("min_order"))
	if rawMin == "" {
		in.MinOrderCents = 0
	} else {
		cents, err := domain.ParseBRLToCents(rawMin)
		if err != nil {
			errs["min_order"] = "Pedido mínimo inválido"
		} else {
			in.MinOrderCents = cents
		}
	}
	if strings.TrimSpace(in.Name) == "" {
		errs["name"] = "Nome obrigatório"
	}
	if domain.WhatsAppChatID(in.WhatsApp) == "" {
		errs["whatsapp"] = "WhatsApp inválido"
	}
	return in, errs
}

// Index lists wholesale suppliers with last WhatsApp line.
func (h *SuppliersHandler) Index(w http.ResponseWriter, r *http.Request) {
	list, err := h.store.ListSuppliers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rows := make([]map[string]any, 0, len(list))
	for _, item := range list {
		rows = append(rows, supplierRow(item))
	}
	_ = h.inertia.Render(w, r, "Suppliers/Index", withCompany(h.store, inertia.Props{
		"site":      meta.ForRequest(h.site, r),
		"suppliers": rows,
	}))
}

// Show renders one supplier conversation.
func (h *SuppliersHandler) Show(w http.ResponseWriter, r *http.Request, id int64) {
	sup, err := h.store.GetSupplier(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	msgs, err := h.store.ListSupplierMessages(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rows := make([]map[string]any, 0, len(msgs))
	for _, m := range msgs {
		rows = append(rows, supplierMessageRow(m))
	}
	_ = h.inertia.Render(w, r, "Suppliers/Show", withCompany(h.store, inertia.Props{
		"site":     meta.ForRequest(h.site, r),
		"supplier": supplierShow(sup),
		"messages": rows,
	}))
}

// Create inserts a supplier.
func (h *SuppliersHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := parseFormOrJSON(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	in, errs := supplierInputFromRequest(r)
	if len(errs) > 0 {
		ctx := inertia.SetValidationErrors(r.Context(), errs)
		r = r.WithContext(ctx)
		h.Index(w, r)
		return
	}
	if _, err := h.store.CreateSupplier(in); err != nil {
		ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{"form": err.Error()})
		r = r.WithContext(ctx)
		h.Index(w, r)
		return
	}
	h.inertia.Redirect(w, r, "/suppliers", http.StatusSeeOther)
}

// Update updates a supplier.
func (h *SuppliersHandler) Update(w http.ResponseWriter, r *http.Request, id int64) {
	if err := parseFormOrJSON(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	in, errs := supplierInputFromRequest(r)
	if len(errs) > 0 {
		ctx := inertia.SetValidationErrors(r.Context(), errs)
		r = r.WithContext(ctx)
		h.Index(w, r)
		return
	}
	if err := h.store.UpdateSupplier(id, in); err != nil {
		ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{"form": err.Error()})
		r = r.WithContext(ctx)
		h.Index(w, r)
		return
	}
	h.inertia.Redirect(w, r, "/suppliers", http.StatusSeeOther)
}

// Destroy deletes a supplier.
func (h *SuppliersHandler) Destroy(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.store.DeleteSupplier(id); err != nil {
		ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{"form": err.Error()})
		r = r.WithContext(ctx)
		h.Index(w, r)
		return
	}
	h.inertia.Redirect(w, r, "/suppliers", http.StatusSeeOther)
}
