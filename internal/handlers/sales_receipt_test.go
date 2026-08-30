package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/puppe1990/leilao-erp/internal/store"
)

func seedSaleForReceipt(t *testing.T, s *store.SQLiteStore) (saleID, clientID int64) {
	t.Helper()
	accountID, itemIDs := seedLotWithItems(t, s, 1)
	var err error
	clientID, err = s.CreateClient(store.ClientInput{
		Name:     "Maria Silva",
		Document: "529.982.247-25",
		Phone:    "11 99999-0000",
		Email:    "maria@example.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	saleID, err = s.CreateSale(store.CreateSaleInput{
		ItemID:        itemIDs[0],
		ClientID:      clientID,
		SoldAt:        "2026-07-22T12:00:00Z",
		Channel:       "direct",
		GrossCents:    19900,
		PaymentStatus: "received",
		CashAccountID: accountID,
	})
	if err != nil {
		t.Fatal(err)
	}
	return saleID, clientID
}

func TestSalesHandler_Show_ReceiptProps(t *testing.T) {
	h, s := newSalesHandler(t)
	saleID, _ := seedSaleForReceipt(t, s)
	if err := s.SetCompanyName("Puppe Leiloes"); err != nil {
		t.Fatal(err)
	}
	if err := s.SetCompanyCNPJ("11.444.777/0001-61"); err != nil {
		t.Fatal(err)
	}

	req := inertiaRequest(http.MethodGet, fmt.Sprintf("/sales/%d", saleID), nil)
	rr := httptest.NewRecorder()
	h.Show(rr, req, saleID)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
	}
	assertInertiaComponent(t, rr, "Sales/Show")
	clients := assertInertiaProp(t, rr, "clients")
	list, ok := clients.([]any)
	if !ok || len(list) != 1 {
		t.Fatalf("clients = %#v", clients)
	}
	cnpj := assertInertiaProp(t, rr, "companyCnpj")
	if cnpj != "11.444.777/0001-61" {
		t.Fatalf("companyCnpj = %v", cnpj)
	}
	seller := assertInertiaProp(t, rr, "receiptSellerName")
	if seller != "Puppe Leiloes" {
		t.Fatalf("receiptSellerName = %v", seller)
	}
	saleProp := assertInertiaProp(t, rr, "sale").(map[string]any)
	if saleProp["clientName"] != "Maria Silva" {
		t.Fatalf("sale.clientName = %v", saleProp["clientName"])
	}
}

func TestSalesHandler_ReceiptPDF_OK(t *testing.T) {
	h, s := newSalesHandler(t)
	saleID, _ := seedSaleForReceipt(t, s)
	if err := s.SetCompanyName("Puppe Leiloes"); err != nil {
		t.Fatal(err)
	}
	if err := s.SetCompanyCNPJ("11.444.777/0001-61"); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/sales/%d/recibo.pdf", saleID), nil)
	rr := httptest.NewRecorder()
	h.ReceiptPDF(rr, req, saleID)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
	}
	ct := rr.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/pdf") {
		t.Fatalf("Content-Type = %q", ct)
	}
	body := rr.Body.Bytes()
	if !strings.HasPrefix(string(body), "%PDF-") {
		t.Fatalf("not a PDF: %q", body[:minInt(12, len(body))])
	}
	if !strings.Contains(string(body), "11.444.777/0001-61") {
		t.Fatal("PDF missing seller CNPJ")
	}
	if !strings.Contains(string(body), "Maria Silva") {
		t.Fatal("PDF missing buyer name")
	}
	cd := rr.Header().Get("Content-Disposition")
	if !strings.Contains(cd, "recibo-venda-") {
		t.Fatalf("Content-Disposition = %q", cd)
	}
}

func TestSalesHandler_ReceiptPDF_MissingClient(t *testing.T) {
	h, s := newSalesHandler(t)
	accountID, itemIDs := seedLotWithItems(t, s, 1)
	saleID, err := s.CreateSale(store.CreateSaleInput{
		ItemID:        itemIDs[0],
		SoldAt:        "2026-07-22T12:00:00Z",
		Channel:       "direct",
		GrossCents:    19900,
		PaymentStatus: "received",
		CashAccountID: accountID,
	})
	if err != nil {
		t.Fatal(err)
	}
	_ = s.SetCompanyName("Puppe Leiloes")
	_ = s.SetCompanyCNPJ("11.444.777/0001-61")

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/sales/%d/recibo.pdf", saleID), nil)
	rr := httptest.NewRecorder()
	h.ReceiptPDF(rr, req, saleID)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "cliente") {
		t.Fatalf("error should mention client: %s", rr.Body.String())
	}
}

func TestSalesHandler_AttachClient(t *testing.T) {
	h, s := newSalesHandler(t)
	accountID, itemIDs := seedLotWithItems(t, s, 1)
	saleID, err := s.CreateSale(store.CreateSaleInput{
		ItemID:        itemIDs[0],
		SoldAt:        "2026-07-22T12:00:00Z",
		Channel:       "direct",
		GrossCents:    19900,
		PaymentStatus: "received",
		CashAccountID: accountID,
	})
	if err != nil {
		t.Fatal(err)
	}
	clientID, err := s.CreateClient(store.ClientInput{
		Name:     "Joao Cliente",
		Document: "529.982.247-25",
	})
	if err != nil {
		t.Fatal(err)
	}

	form := fmt.Sprintf("client_id=%d", clientID)
	req := inertiaRequest(http.MethodPost, fmt.Sprintf("/sales/%d/client", saleID), strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.AttachClient(rr, req, saleID)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
	}

	sale, err := s.FindSaleByID(saleID)
	if err != nil {
		t.Fatal(err)
	}
	if sale.ClientID != clientID || sale.ClientName != "Joao Cliente" {
		t.Fatalf("sale client = %d %q", sale.ClientID, sale.ClientName)
	}
}

func TestSalesHandler_ReceiptPDF_NotFound(t *testing.T) {
	h, _ := newSalesHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/sales/999/recibo.pdf", nil)
	rr := httptest.NewRecorder()
	h.ReceiptPDF(rr, req, 999)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rr.Code)
	}
}

func TestSalesHandler_ReceiptPDF_Cancelled(t *testing.T) {
	h, s := newSalesHandler(t)
	accountID, itemIDs := seedLotWithItems(t, s, 1)
	saleID, err := s.CreateSale(store.CreateSaleInput{
		ItemID:        itemIDs[0],
		SoldAt:        "2026-07-22T12:00:00Z",
		Channel:       "direct",
		GrossCents:    19900,
		PaymentStatus: "pending",
		DueOn:         "2026-08-01",
		CashAccountID: accountID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.CancelPendingSale(saleID); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf(
		"/sales/%d/recibo.pdf?seller_name=X&seller_cnpj=11444777000161&buyer_name=Maria&buyer_document=52998224725",
		saleID,
	), nil)
	rr := httptest.NewRecorder()
	h.ReceiptPDF(rr, req, saleID)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 body=%s", rr.Code, rr.Body.String())
	}
}
