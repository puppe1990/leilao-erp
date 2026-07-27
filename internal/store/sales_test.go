package store

import (
	"testing"

	"github.com/puppe1990/leilao-erp/internal/domain"
)

func TestCreateSale_DirectReceived(t *testing.T) {
	st := newTestStore(t)

	accountID, err := st.InsertCashAccount("PIX principal", "pix", 0)
	if err != nil {
		t.Fatal(err)
	}

	const purchaseCost int64 = 2741
	lotID, err := st.CreateLotPurchase(CreateLotInput{
		Name:        "Monitores — leilão Jul/2026",
		PurchasedAt: "2026-07-20",
		ItemTitle:   "Monitor",
		ItemQty:     1,
		Costs: []CostInput{
			{Label: "Arremate", AmountCents: purchaseCost, AlreadyPaid: true},
		},
		CashAccountID: accountID,
		PaidAt:        "2026-07-20T12:00:00Z",
	})
	if err != nil {
		t.Fatal(err)
	}

	items, err := st.ListItemsByLot(lotID)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("items=%d", len(items))
	}
	item := items[0]
	if item.UnitCostCents != purchaseCost {
		t.Fatalf("unit cost=%d want %d", item.UnitCostCents, purchaseCost)
	}

	const gross int64 = 15000
	saleID, err := st.CreateSale(CreateSaleInput{
		ItemID:        item.ID,
		SoldAt:        "2026-07-22T15:00:00Z",
		Channel:       "direct",
		GrossCents:    gross,
		FeeCents:      0,
		ShippingCents: 0,
		PaymentStatus: "received",
		CashAccountID: accountID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if saleID == 0 {
		t.Fatal("saleID = 0")
	}

	items, err = st.ListItemsByLot(lotID)
	if err != nil {
		t.Fatal(err)
	}
	if items[0].Status != "sold" {
		t.Fatalf("item status=%s want sold", items[0].Status)
	}

	sales, err := st.ListSales()
	if err != nil {
		t.Fatal(err)
	}
	if len(sales) != 1 {
		t.Fatalf("sales=%d", len(sales))
	}
	sale := sales[0]
	if sale.NetCents != gross {
		t.Fatalf("net=%d want %d", sale.NetCents, gross)
	}
	if sale.PaymentStatus != "received" {
		t.Fatalf("payment_status=%s", sale.PaymentStatus)
	}
	if sale.UnitCostCentsAtSale != purchaseCost {
		t.Fatalf("unit_cost_at_sale=%d want %d", sale.UnitCostCentsAtSale, purchaseCost)
	}
	if sale.Channel != "direct" {
		t.Fatalf("channel=%s", sale.Channel)
	}

	margin := domain.Margin(sale.NetCents, sale.UnitCostCentsAtSale)
	if margin != gross-purchaseCost {
		t.Fatalf("margin=%d want %d", margin, gross-purchaseCost)
	}

	// cash: -purchase + sale net
	bal, err := st.CashBalance(accountID)
	if err != nil {
		t.Fatal(err)
	}
	wantBal := -purchaseCost + gross
	if bal != wantBal {
		t.Fatalf("balance=%d want %d (-purchase + sale)", bal, wantBal)
	}

	lot, err := st.FindLot(lotID)
	if err != nil {
		t.Fatal(err)
	}
	if lot.Status != "sold" {
		t.Fatalf("lot status=%s want sold", lot.Status)
	}
}

func TestCreateSale_MarketplacePending(t *testing.T) {
	st := newTestStore(t)

	accountID, err := st.InsertCashAccount("PIX principal", "pix", 0)
	if err != nil {
		t.Fatal(err)
	}

	const purchaseCost int64 = 2741
	lotID, err := st.CreateLotPurchase(CreateLotInput{
		Name:        "Monitores ML",
		PurchasedAt: "2026-07-20",
		ItemTitle:   "Monitor LG",
		ItemQty:     2,
		Costs: []CostInput{
			{Label: "Arremate", AmountCents: purchaseCost * 2, AlreadyPaid: true},
		},
		CashAccountID: accountID,
		PaidAt:        "2026-07-20T12:00:00Z",
	})
	if err != nil {
		t.Fatal(err)
	}

	items, err := st.ListItemsByLot(lotID)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("items=%d", len(items))
	}
	item := items[0]
	unitCost := item.UnitCostCents

	balBefore, err := st.CashBalance(accountID)
	if err != nil {
		t.Fatal(err)
	}

	const gross, fee, shipping int64 = 18000, 3000, 2000
	wantNet := domain.SaleNet(gross, fee, shipping) // 13000
	if wantNet != 13000 {
		t.Fatalf("SaleNet=%d want 13000", wantNet)
	}

	saleID, err := st.CreateSale(CreateSaleInput{
		ItemID:        item.ID,
		SoldAt:        "2026-07-23T10:00:00Z",
		Channel:       "mercadolivre",
		GrossCents:    gross,
		FeeCents:      fee,
		ShippingCents: shipping,
		PaymentStatus: "pending",
		DueOn:         "2026-08-06",
	})
	if err != nil {
		t.Fatal(err)
	}
	if saleID == 0 {
		t.Fatal("saleID = 0")
	}

	items, err = st.ListItemsByLot(lotID)
	if err != nil {
		t.Fatal(err)
	}
	var sold, inStock int
	for _, it := range items {
		switch it.Status {
		case "sold":
			sold++
		case "in_stock":
			inStock++
		}
	}
	if sold != 1 || inStock != 1 {
		t.Fatalf("sold=%d in_stock=%d want 1/1", sold, inStock)
	}

	sales, err := st.ListSales()
	if err != nil {
		t.Fatal(err)
	}
	if len(sales) != 1 {
		t.Fatalf("sales=%d", len(sales))
	}
	sale := sales[0]
	if sale.NetCents != wantNet {
		t.Fatalf("net=%d want %d", sale.NetCents, wantNet)
	}
	if sale.PaymentStatus != "pending" {
		t.Fatalf("payment_status=%s", sale.PaymentStatus)
	}
	if sale.UnitCostCentsAtSale != unitCost {
		t.Fatalf("unit_cost_at_sale=%d want %d", sale.UnitCostCentsAtSale, unitCost)
	}
	if sale.Channel != "mercadolivre" {
		t.Fatalf("channel=%s", sale.Channel)
	}
	if sale.FeeCents != fee || sale.ShippingCents != shipping {
		t.Fatalf("fee=%d shipping=%d", sale.FeeCents, sale.ShippingCents)
	}

	recs, err := st.ListReceivables()
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 1 {
		t.Fatalf("receivables=%d", len(recs))
	}
	rec := recs[0]
	if rec.Status != "open" {
		t.Fatalf("receivable status=%s want open", rec.Status)
	}
	if rec.AmountCents != wantNet {
		t.Fatalf("receivable amount=%d want %d", rec.AmountCents, wantNet)
	}
	if rec.DueOn != "2026-08-06" {
		t.Fatalf("due_on=%s", rec.DueOn)
	}
	if rec.SaleID == nil || *rec.SaleID != saleID {
		t.Fatalf("receivable sale_id=%v want %d", rec.SaleID, saleID)
	}
	if rec.Description == "" {
		t.Fatal("expected receivable description with item title")
	}

	// no cash in yet — balance unchanged after pending sale
	balAfter, err := st.CashBalance(accountID)
	if err != nil {
		t.Fatal(err)
	}
	if balAfter != balBefore {
		t.Fatalf("balance changed on pending sale: before=%d after=%d", balBefore, balAfter)
	}

	lot, err := st.FindLot(lotID)
	if err != nil {
		t.Fatal(err)
	}
	if lot.Status != "partial" {
		t.Fatalf("lot status=%s want partial", lot.Status)
	}
}

func TestCreateSale_WithAccessories(t *testing.T) {
	st := newTestStore(t)

	accountID, err := st.InsertCashAccount("PIX principal", "pix", 0)
	if err != nil {
		t.Fatal(err)
	}

	monitorLot, err := st.CreateLotPurchase(CreateLotInput{
		Name: "Monitores kit", PurchasedAt: "2026-07-20", ItemTitle: "Monitor", ItemQty: 2,
		Costs:         []CostInput{{Label: "Arremate", AmountCents: 5000, AlreadyPaid: true}},
		CashAccountID: accountID, PaidAt: "2026-07-20T12:00:00Z",
	})
	if err != nil {
		t.Fatal(err)
	}
	powerLot, err := st.CreateLotPurchase(CreateLotInput{
		Name: "Cabos força", PurchasedAt: "2026-07-21", ItemTitle: "Cabo de força", ItemQty: 2,
		Costs:         []CostInput{{Label: "Compra", AmountCents: 400, AlreadyPaid: true}},
		CashAccountID: accountID, PaidAt: "2026-07-21T12:00:00Z",
	})
	if err != nil {
		t.Fatal(err)
	}
	hdmiLot, err := st.CreateLotPurchase(CreateLotInput{
		Name: "Cabos HDMI", PurchasedAt: "2026-07-21", ItemTitle: "Cabo HDMI", ItemQty: 2,
		Costs:         []CostInput{{Label: "Compra", AmountCents: 600, AlreadyPaid: true}},
		CashAccountID: accountID, PaidAt: "2026-07-21T12:00:00Z",
	})
	if err != nil {
		t.Fatal(err)
	}

	monitors, _ := st.ListItemsByLot(monitorLot)
	powers, _ := st.ListItemsByLot(powerLot)
	hdmis, _ := st.ListItemsByLot(hdmiLot)

	wantCost := monitors[0].UnitCostCents + powers[0].UnitCostCents + hdmis[0].UnitCostCents
	const gross int64 = 20000

	saleID, err := st.CreateSale(CreateSaleInput{
		ItemID:        monitors[0].ID,
		AccessoryIDs:  []int64{powers[0].ID, hdmis[0].ID},
		SoldAt:        "2026-07-25T12:00:00Z",
		Channel:       "direct",
		GrossCents:    gross,
		PaymentStatus: "received",
		CashAccountID: accountID,
	})
	if err != nil {
		t.Fatal(err)
	}

	sale, err := st.FindSaleByID(saleID)
	if err != nil {
		t.Fatal(err)
	}
	if sale.UnitCostCentsAtSale != wantCost {
		t.Fatalf("total cost at sale=%d want %d", sale.UnitCostCentsAtSale, wantCost)
	}

	lines, err := st.ListSaleLines(saleID)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 3 {
		t.Fatalf("lines=%d want 3", len(lines))
	}
	roles := map[string]int{}
	for _, ln := range lines {
		roles[ln.Role]++
	}
	if roles["main"] != 1 || roles["accessory"] != 2 {
		t.Fatalf("roles=%v", roles)
	}

	// All three items sold
	for _, lotID := range []int64{monitorLot, powerLot, hdmiLot} {
		items, _ := st.ListItemsByLot(lotID)
		var sold, stock int
		for _, it := range items {
			if it.Status == "sold" {
				sold++
			}
			if it.Status == "in_stock" {
				stock++
			}
		}
		if sold != 1 || stock != 1 {
			t.Fatalf("lot %d: sold=%d stock=%d want 1/1", lotID, sold, stock)
		}
	}

	sales, err := st.ListSales()
	if err != nil {
		t.Fatal(err)
	}
	if sales[0].LineCount != 3 {
		t.Fatalf("line count=%d", sales[0].LineCount)
	}
	if sales[0].Composition == "" || sales[0].Composition == sales[0].ItemTitle {
		t.Fatalf("composition=%q", sales[0].Composition)
	}

	// Cancel path: pending multi-line restores all
	sale2, err := st.CreateSale(CreateSaleInput{
		ItemID:        monitors[1].ID,
		AccessoryIDs:  []int64{powers[1].ID, hdmis[1].ID},
		SoldAt:        "2026-07-26T12:00:00Z",
		Channel:       "mercadolivre",
		GrossCents:    18000,
		FeeCents:      1000,
		PaymentStatus: "pending",
		DueOn:         "2026-08-10",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := st.CancelPendingSale(sale2); err != nil {
		t.Fatal(err)
	}
	for _, lotID := range []int64{monitorLot, powerLot, hdmiLot} {
		items, _ := st.ListItemsByLot(lotID)
		var stock int
		for _, it := range items {
			if it.Status == "in_stock" {
				stock++
			}
		}
		// one still sold from first sale, one restored
		if stock != 1 {
			t.Fatalf("after cancel lot %d stock=%d want 1", lotID, stock)
		}
	}
}

func TestCreateSale_RejectsSoldItem(t *testing.T) {
	st := newTestStore(t)

	accountID, err := st.InsertCashAccount("PIX principal", "pix", 0)
	if err != nil {
		t.Fatal(err)
	}

	lotID, err := st.CreateLotPurchase(CreateLotInput{
		Name:        "Lote único",
		PurchasedAt: "2026-07-20",
		ItemTitle:   "Item",
		ItemQty:     1,
		Costs: []CostInput{
			{Label: "Arremate", AmountCents: 1000, AlreadyPaid: true},
		},
		CashAccountID: accountID,
		PaidAt:        "2026-07-20T12:00:00Z",
	})
	if err != nil {
		t.Fatal(err)
	}

	items, err := st.ListItemsByLot(lotID)
	if err != nil {
		t.Fatal(err)
	}
	itemID := items[0].ID

	_, err = st.CreateSale(CreateSaleInput{
		ItemID:        itemID,
		SoldAt:        "2026-07-22T15:00:00Z",
		Channel:       "direct",
		GrossCents:    5000,
		PaymentStatus: "received",
		CashAccountID: accountID,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = st.CreateSale(CreateSaleInput{
		ItemID:        itemID,
		SoldAt:        "2026-07-23T15:00:00Z",
		Channel:       "direct",
		GrossCents:    6000,
		PaymentStatus: "received",
		CashAccountID: accountID,
	})
	if err == nil {
		t.Fatal("expected error on second sale of sold item")
	}
}
