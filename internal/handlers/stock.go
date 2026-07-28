package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/meta"
	inertia "github.com/romsar/gonertia/v3"

	"github.com/puppe1990/leilao-erp/internal/domain"
	"github.com/puppe1990/leilao-erp/internal/store"
)

type StockHandler struct {
	renderer *cais.Renderer
	store    store.Store
	site     meta.Site
	cfg      cais.Config
	inertia  *inertia.Inertia
}

func NewStockHandler(renderer *cais.Renderer, s store.Store, site meta.Site, cfg cais.Config, i *inertia.Inertia) *StockHandler {
	return &StockHandler{renderer: renderer, store: s, site: site, cfg: cfg, inertia: i}
}

// Index lists stock grouped by product (same name/product_id), with quantity.
func (h *StockHandler) Index(w http.ResponseWriter, r *http.Request) {
	groups, err := h.store.ListStockProductGroups()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Unit-level totals for summary cards (cost/margin exact)
	units, err := h.store.ListItemsInStock()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var totalCost, totalHint, potentialMargin int64
	var withHint int
	for _, it := range units {
		totalCost += it.UnitCostCents
		if it.SalePriceHintCents != nil {
			totalHint += *it.SalePriceHintCents
			potentialMargin += domain.Margin(*it.SalePriceHintCents, it.UnitCostCents)
			withHint++
		}
	}

	rows := make([]map[string]any, 0, len(groups))
	for _, g := range groups {
		isAcc := g.Kind == "accessory" || isAccessoryTitle(g.Name)
		saleHint := ""
		salePriceRaw := ""
		var salePriceCents any
		margin := ""
		marginTotal := ""
		if g.SalePriceHintCents != nil {
			saleHint = domain.FormatBRL(*g.SalePriceHintCents)
			salePriceRaw = formatCentsInput(*g.SalePriceHintCents)
			salePriceCents = *g.SalePriceHintCents
			m := domain.Margin(*g.SalePriceHintCents, g.UnitCostCents)
			margin = domain.FormatBRL(m)
			marginTotal = domain.FormatBRL(m * int64(g.QtyInStock))
		}
		// Stable row id for Svelte keys: product id, else sample item
		rowID := g.ID
		if rowID == 0 {
			rowID = g.SampleItemID
		}
		rows = append(rows, map[string]any{
			"id":             rowID,
			"productId":      g.ID,
			"sampleItemId":   g.SampleItemID,
			"lotId":          g.SampleLotID,
			"title":          g.Name,
			"qty":            g.QtyInStock,
			"sku":            "",
			"unitCost":       domain.FormatBRL(g.UnitCostCents),
			"unitCostRaw":    g.UnitCostCents,
			"salePriceHint":  saleHint,
			"salePriceRaw":   salePriceRaw,
			"salePriceCents": salePriceCents,
			"marginHint":     margin,
			"marginTotal":    marginTotal,
			"isAccessory":    isAcc,
			"canEdit":        true,
		})
	}

	sort.SliceStable(rows, func(i, j int) bool {
		ai := rows[i]["isAccessory"].(bool)
		aj := rows[j]["isAccessory"].(bool)
		if ai != aj {
			return !ai && aj
		}
		ti := rows[i]["title"].(string)
		tj := rows[j]["title"].(string)
		if ti != tj {
			return strings.ToLower(ti) < strings.ToLower(tj)
		}
		return rows[i]["id"].(int64) < rows[j]["id"].(int64)
	})

	// Compact group summary (same data, kept for optional UI)
	groupRows := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		groupRows = append(groupRows, map[string]any{
			"title":           row["title"],
			"count":           row["qty"],
			"unitCost":        row["unitCost"],
			"salePriceHint":   row["salePriceHint"],
			"potentialMargin": row["marginTotal"],
			"isAccessory":     row["isAccessory"],
		})
	}

	multiplierLabel := "—"
	if totalCost > 0 && totalHint > 0 {
		x10 := (totalHint*10 + totalCost/2) / totalCost
		multiplierLabel = fmt.Sprintf("%d,%d×", x10/10, x10%10)
	}

	_ = h.inertia.Render(w, r, "Stock/Index", withCompany(h.store, inertia.Props{
		"site":   meta.ForRequest(h.site, r),
		"items":  rows,
		"groups": groupRows,
		"summary": map[string]any{
			"count":           len(units),
			"productCount":    len(groups),
			"totalCost":       domain.FormatBRL(totalCost),
			"potentialGross":  domain.FormatBRL(totalHint),
			"potentialMargin": domain.FormatBRL(potentialMargin),
			"multiplier":      multiplierLabel,
			"pricedCount":     withHint,
		},
	}))
}

// ExportCSV downloads all in-stock units as UTF-8 CSV (Excel-friendly with BOM).
func (h *StockHandler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	items, err := h.store.ListItemsInStock()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("estoque-%s.csv", time.Now().UTC().Format("2006-01-02"))
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	if _, err := w.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
		return
	}

	cw := csv.NewWriter(w)
	cw.Comma = ';'
	_ = cw.Write([]string{
		"id",
		"lote_id",
		"product_id",
		"titulo",
		"sku",
		"tipo",
		"custo_centavos",
		"custo",
		"preco_venda_centavos",
		"preco_venda",
		"margem_centavos",
		"margem",
	})

	for _, it := range items {
		sku := ""
		if it.SKU != nil {
			sku = *it.SKU
		}
		tipo := "principal"
		if isAccessoryTitle(it.Title) {
			tipo = "acessorio"
		}
		pid := ""
		if it.ProductID != nil {
			pid = strconv.FormatInt(*it.ProductID, 10)
		}
		saleCents := ""
		saleFmt := ""
		marginCents := ""
		marginFmt := ""
		if it.SalePriceHintCents != nil {
			saleCents = strconv.FormatInt(*it.SalePriceHintCents, 10)
			saleFmt = domain.FormatBRL(*it.SalePriceHintCents)
			m := domain.Margin(*it.SalePriceHintCents, it.UnitCostCents)
			marginCents = strconv.FormatInt(m, 10)
			marginFmt = domain.FormatBRL(m)
		}
		_ = cw.Write([]string{
			strconv.FormatInt(it.ID, 10),
			strconv.FormatInt(it.LotID, 10),
			pid,
			it.Title,
			sku,
			tipo,
			strconv.FormatInt(it.UnitCostCents, 10),
			domain.FormatBRL(it.UnitCostCents),
			saleCents,
			saleFmt,
			marginCents,
			marginFmt,
		})
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		return
	}
}
