package handlers

import (
	"net/http"
	"sort"
	"strings"

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

func (h *StockHandler) Index(w http.ResponseWriter, r *http.Request) {
	items, err := h.store.ListItemsInStock()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Group by title for summary cards
	type groupAgg struct {
		Title     string
		Count     int
		UnitCost  int64 // avg-ish: first item cost (all same after rateio)
		SaleHint  *int64
		TotalCost int64
	}
	groups := map[string]*groupAgg{}
	var totalCost, totalHint int64
	var withHint int

	rows := make([]map[string]any, 0, len(items))
	for _, it := range items {
		sku := ""
		if it.SKU != nil {
			sku = *it.SKU
		}
		saleHint := ""
		salePriceRaw := ""
		margin := ""
		if it.SalePriceHintCents != nil {
			saleHint = domain.FormatBRL(*it.SalePriceHintCents)
			salePriceRaw = formatCentsInput(*it.SalePriceHintCents)
			margin = domain.FormatBRL(domain.Margin(*it.SalePriceHintCents, it.UnitCostCents))
			totalHint += *it.SalePriceHintCents
			withHint++
		}
		totalCost += it.UnitCostCents

		g := groups[it.Title]
		if g == nil {
			g = &groupAgg{Title: it.Title, UnitCost: it.UnitCostCents, SaleHint: it.SalePriceHintCents}
			groups[it.Title] = g
		}
		g.Count++
		g.TotalCost += it.UnitCostCents
		if g.SaleHint == nil && it.SalePriceHintCents != nil {
			g.SaleHint = it.SalePriceHintCents
		}

		rows = append(rows, map[string]any{
			"id":            it.ID,
			"lotId":         it.LotID,
			"title":         it.Title,
			"sku":           sku,
			"unitCost":      domain.FormatBRL(it.UnitCostCents),
			"unitCostRaw":   it.UnitCostCents,
			"salePriceHint": saleHint,
			"salePriceRaw":  salePriceRaw,
			"marginHint":    margin,
			"isAccessory":   isAccessoryTitle(it.Title),
			"canEdit":       true,
		})
	}

	// Stable sort: mains first, then accessories, by title then id
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

	groupRows := make([]map[string]any, 0, len(groups))
	for _, g := range groups {
		sale := "—"
		margin := "—"
		potential := int64(0)
		if g.SaleHint != nil {
			sale = domain.FormatBRL(*g.SaleHint)
			potential = *g.SaleHint * int64(g.Count)
			margin = domain.FormatBRL(domain.Margin(*g.SaleHint, g.UnitCost) * int64(g.Count))
		}
		groupRows = append(groupRows, map[string]any{
			"title":           g.Title,
			"count":           g.Count,
			"unitCost":        domain.FormatBRL(g.UnitCost),
			"salePriceHint":   sale,
			"potentialGross":  domain.FormatBRL(potential),
			"potentialMargin": margin,
			"isAccessory":     isAccessoryTitle(g.Title),
		})
	}
	sort.Slice(groupRows, func(i, j int) bool {
		ai := groupRows[i]["isAccessory"].(bool)
		aj := groupRows[j]["isAccessory"].(bool)
		if ai != aj {
			return !ai && aj
		}
		return groupRows[i]["title"].(string) < groupRows[j]["title"].(string)
	})

	potentialGross := totalHint
	potentialMargin := int64(0)
	if withHint > 0 {
		// only items with hint contribute
		for _, it := range items {
			if it.SalePriceHintCents != nil {
				potentialMargin += domain.Margin(*it.SalePriceHintCents, it.UnitCostCents)
			}
		}
	}

	_ = h.inertia.Render(w, r, "Stock/Index", withCompany(h.store, inertia.Props{
		"site":   meta.ForRequest(h.site, r),
		"items":  rows,
		"groups": groupRows,
		"summary": map[string]any{
			"count":           len(items),
			"totalCost":       domain.FormatBRL(totalCost),
			"potentialGross":  domain.FormatBRL(potentialGross),
			"potentialMargin": domain.FormatBRL(potentialMargin),
			"pricedCount":     withHint,
		},
	}))
}
