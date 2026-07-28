/**
 * Pure helpers for Stock inventory table: brand/base tags, filter, sort.
 * Kept out of Index.svelte so agents can load/test without the full page.
 */

/** @param {string} title */
export function brandOfTitle(title) {
  const t = String(title || '')
  const m = t.match(/^(?:Monitor\s+)?([A-Za-zÀ-ú]+)/i)
  if (!m) return 'Outros'
  const b = m[1]
  if (/^monitor$/i.test(b)) return 'Genérico'
  if (/^cabo$/i.test(b)) return 'Cabo'
  return b.charAt(0).toUpperCase() + b.slice(1)
}

/** @returns {'with'|'without'|'unset'} */
export function baseOfTitle(title) {
  const t = String(title || '').toLowerCase()
  if (t.includes('sem base')) return 'without'
  if (t.includes('com base')) return 'with'
  return 'unset'
}

/**
 * Normalize sale price to integer cents for sorting.
 * Accepts:
 * - item.salePriceCents (preferred, int)
 * - item.salePriceRaw as int cents (e.g. 39900)
 * - item.salePriceRaw as BRL input string (e.g. "399,00" or "1.399,00")
 */
export function salePriceCents(item) {
  if (item == null) return null

  if (item.salePriceCents != null && item.salePriceCents !== '') {
    const n = Number(item.salePriceCents)
    if (Number.isFinite(n)) return Math.round(n)
  }

  const raw = item.salePriceRaw
  if (raw == null || raw === '') return null

  if (typeof raw === 'number') {
    return Number.isFinite(raw) ? Math.round(raw) : null
  }

  const s = String(raw).trim()
  if (s === '') return null

  // Brazilian display/input first: "399,00" / "1.234,56"
  // Number("399,00") is NaN in JS — that was breaking price sort.
  if (s.includes(',')) {
    const reais = parseFloat(s.replace(/\./g, '').replace(',', '.'))
    return Number.isFinite(reais) ? Math.round(reais * 100) : null
  }

  // Plain digits: treat as cents (39900)
  if (/^\d+$/.test(s)) return parseInt(s, 10)

  const n = Number(s)
  return Number.isFinite(n) ? Math.round(n) : null
}

export function marginCents(item) {
  const sale = salePriceCents(item)
  if (sale == null) return null
  const cost = Number(item.unitCostRaw) || 0
  return sale - cost
}

/**
 * @param {object[]} items
 * @param {{
 *   query?: string,
 *   filterType?: 'all'|'main'|'accessory',
 *   filterPrice?: 'all'|'priced'|'unpriced',
 *   filterBase?: 'all'|'with'|'without'|'unset',
 *   filterBrand?: string,
 *   sortKey?: string,
 *   sortDir?: 'asc'|'desc',
 * }} opts
 */
export function filterAndSortStockItems(items, opts = {}) {
  const list = Array.isArray(items) ? items : []
  const query = String(opts.query || '')
    .trim()
    .toLowerCase()
  const filterType = opts.filterType || 'all'
  const filterPrice = opts.filterPrice || 'all'
  const filterBase = opts.filterBase || 'all'
  const filterBrand = opts.filterBrand || 'all'
  const sortKey = opts.sortKey || 'id'
  const sortDir = opts.sortDir === 'desc' ? 'desc' : 'asc'
  const dir = sortDir === 'asc' ? 1 : -1

  const filtered = list.filter((it) => {
    if (filterType === 'main' && it.isAccessory) return false
    if (filterType === 'accessory' && !it.isAccessory) return false
    if (filterPrice === 'priced' && !it.salePriceHint) return false
    if (filterPrice === 'unpriced' && it.salePriceHint) return false
    if (filterBase !== 'all' && baseOfTitle(it.title) !== filterBase) return false
    if (filterBrand !== 'all' && brandOfTitle(it.title) !== filterBrand) return false
    if (!query) return true
    const hay = [
      it.id,
      it.productId,
      it.lotId,
      it.title,
      it.sku,
      it.qty,
      it.unitCost,
      it.salePriceHint,
      brandOfTitle(it.title),
      it.isAccessory ? 'acessorio cabo' : 'principal monitor',
      baseOfTitle(it.title) === 'with' ? 'com base' : '',
      baseOfTitle(it.title) === 'without' ? 'sem base' : '',
    ]
      .filter(Boolean)
      .join(' ')
      .toLowerCase()
    return hay.includes(query)
  })

  return filtered.slice().sort((a, b) => {
    let av
    let bv
    switch (sortKey) {
      case 'title':
        return (a.title || '').localeCompare(b.title || '', 'pt-BR') * dir
      case 'type':
        return ((a.isAccessory ? 1 : 0) - (b.isAccessory ? 1 : 0)) * dir
      case 'qty':
        return ((Number(a.qty) || 1) - (Number(b.qty) || 1)) * dir
      case 'unitCost':
        return ((Number(a.unitCostRaw) || 0) - (Number(b.unitCostRaw) || 0)) * dir
      case 'salePrice': {
        av = salePriceCents(a)
        bv = salePriceCents(b)
        if (av == null && bv == null) return 0
        if (av == null) return 1
        if (bv == null) return -1
        return (av - bv) * dir
      }
      case 'margin': {
        av = marginCents(a)
        bv = marginCents(b)
        if (av == null && bv == null) return 0
        if (av == null) return 1
        if (bv == null) return -1
        return (av - bv) * dir
      }
      case 'lotId':
        return ((Number(a.lotId) || 0) - (Number(b.lotId) || 0)) * dir
      case 'id':
      default:
        return ((Number(a.id) || 0) - (Number(b.id) || 0)) * dir
    }
  })
}

/** Unique sorted brand labels from stock rows. */
export function brandsFromItems(items) {
  return [
    ...new Set((items || []).map((it) => brandOfTitle(it.title)).filter(Boolean)),
  ].sort((a, b) => a.localeCompare(b, 'pt-BR'))
}
