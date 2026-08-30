package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	inertia "github.com/romsar/gonertia/v3"

	"github.com/puppe1990/leilao-erp/internal/domain"
	"github.com/puppe1990/leilao-erp/internal/models"
	"github.com/puppe1990/leilao-erp/internal/store"
)

// ReceiptPDF downloads a non-fiscal sale receipt PDF (seller CNPJ + client data).
func (h *SalesHandler) ReceiptPDF(w http.ResponseWriter, r *http.Request, id int64) {
	sale, err := h.store.FindSaleByID(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if sale.PaymentStatus == "cancelled" {
		http.Error(w, "Venda cancelada não emite recibo.", http.StatusBadRequest)
		return
	}

	if sale.ClientID <= 0 {
		http.Error(w, "Esta venda não tem cliente. Vincule um cliente na venda para emitir o recibo.", http.StatusBadRequest)
		return
	}
	buyer := domain.ReceiptParty{
		Name:     sale.ClientName,
		Document: sale.ClientDocument,
		Phone:    sale.ClientPhone,
		Email:    sale.ClientEmail,
	}

	sellerName := companyName(h.store)
	sellerCNPJ, _ := h.store.CompanyCNPJ()

	items, err := receiptItemsForSale(h.store, sale)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pdf, err := domain.BuildReceiptPDF(domain.Receipt{
		Number:     strconv.FormatInt(sale.ID, 10),
		IssuedAt:   sale.SoldAt,
		Channel:    channelLabel(sale.Channel),
		Seller:     domain.ReceiptParty{Name: sellerName, Document: sellerCNPJ},
		Buyer:      buyer,
		Items:      items,
		TotalCents: sale.GrossCents,
	})
	if err != nil {
		http.Error(w, receiptUserError(err), http.StatusBadRequest)
		return
	}

	filename := fmt.Sprintf("recibo-venda-%d.pdf", sale.ID)
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	_, _ = w.Write(pdf)
}

// AttachClient saves the buyer on the sale so receipts use that client.
func (h *SalesHandler) AttachClient(w http.ResponseWriter, r *http.Request, id int64) {
	if err := parseFormOrJSON(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	clientID, _ := strconv.ParseInt(strings.TrimSpace(r.FormValue("client_id")), 10, 64)
	if clientID <= 0 {
		ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{
			"client_id": "Selecione o cliente",
		})
		h.Show(w, r.WithContext(ctx), id)
		return
	}
	if err := h.store.SetSaleClient(id, clientID); err != nil {
		ctx := inertia.SetValidationErrors(r.Context(), inertia.ValidationErrors{"form": err.Error()})
		h.Show(w, r.WithContext(ctx), id)
		return
	}
	h.inertia.Redirect(w, r, fmt.Sprintf("/sales/%d", id), http.StatusSeeOther)
}

func receiptItemsForSale(s store.Store, sale models.Sale) ([]domain.ReceiptItem, error) {
	lines, err := s.ListSaleLines(sale.ID)
	if err != nil {
		return nil, fmt.Errorf("list sale lines: %w", err)
	}
	items := make([]domain.ReceiptItem, 0, len(lines)+1)
	for _, ln := range lines {
		title := strings.TrimSpace(ln.ItemTitle)
		if title == "" {
			continue
		}
		items = append(items, domain.ReceiptItem{Description: title})
	}
	if len(items) == 0 && strings.TrimSpace(sale.ItemTitle) != "" {
		items = append(items, domain.ReceiptItem{Description: sale.ItemTitle})
	}
	return items, nil
}

func clientReceiptProps(s store.Store) ([]map[string]any, error) {
	list, err := s.ListClients()
	if err != nil {
		return nil, err
	}
	rows := make([]map[string]any, 0, len(list))
	for _, c := range list {
		rows = append(rows, map[string]any{
			"id":       c.ID,
			"name":     c.Name,
			"phone":    c.Phone,
			"email":    c.Email,
			"document": c.Document,
		})
	}
	return rows, nil
}

func receiptUserError(err error) string {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "seller CNPJ"):
		return "Informe um CNPJ válido da empresa."
	case strings.Contains(msg, "seller name"):
		return "Informe o nome da empresa."
	case strings.Contains(msg, "buyer name"):
		return "Informe o nome do cliente."
	case strings.Contains(msg, "buyer document"):
		return "Informe um CPF ou CNPJ válido do cliente."
	case strings.Contains(msg, "total"):
		return "A venda precisa de um valor bruto maior que zero."
	default:
		return msg
	}
}
