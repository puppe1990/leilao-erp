/**
 * Pure helpers for Lots Show items table: filter + multi-column sort.
 */

/**
 * @typedef {{ key: string, dir: 'asc'|'desc' }} LotItemSort
 */

export const LOT_ITEM_SORTS_STORAGE_KEY = 'leilao-erp.lot-items.sorts.v1'

const LOT_ITEM_SORT_KEYS = new Set([
  'id',
  'title',
  'sku',
  'unitCost',
  'salePrice',
  'margin',
  'status',
])

const STATUS_RANK = {
  in_stock: 0,
  'em estoque': 0,
  reserved: 1,
  reservado: 1,
  sold: 2,
  vendido: 2,
}

/** Default when nothing is stored — kit units stay in sequence. */
export function defaultLotItemSorts() {
  return [{ key: 'id', dir: 'asc' }]
}

/**
 * @param {unknown} raw
 * @returns {LotItemSort[]}
 */
export function sanitizeLotItemSorts(raw) {
  if (!Array.isArray(raw) || !raw.length) return defaultLotItemSorts()
  const out = []
  const seen = new Set()
  for (const item of raw) {
    if (!item || typeof item !== 'object') continue
    const key = String(/** @type {{ key?: string }} */ (item).key || '')
    if (!LOT_ITEM_SORT_KEYS.has(key) || seen.has(key)) continue
    seen.add(key)
    const dir = /** @type {{ dir?: string }} */ (item).dir === 'desc' ? 'desc' : 'asc'
    out.push({ key, dir })
    if (out.length >= 6) break
  }
  return out.length ? out : defaultLotItemSorts()
}

/**
 * @param {Storage | null | undefined} storage
 * @returns {LotItemSort[]}
 */
export function loadLotItemSorts(storage = typeof localStorage !== 'undefined' ? localStorage : null) {
  if (!storage) return defaultLotItemSorts()
  try {
    const raw = storage.getItem(LOT_ITEM_SORTS_STORAGE_KEY)
    if (!raw) return defaultLotItemSorts()
    return sanitizeLotItemSorts(JSON.parse(raw))
  } catch {
    return defaultLotItemSorts()
  }
}

/**
 * @param {LotItemSort[]} sorts
 * @param {Storage | null | undefined} storage
 */
export function saveLotItemSorts(
  sorts,
  storage = typeof localStorage !== 'undefined' ? localStorage : null,
) {
  if (!storage) return
  try {
    storage.setItem(LOT_ITEM_SORTS_STORAGE_KEY, JSON.stringify(sanitizeLotItemSorts(sorts)))
  } catch {
    // quota / private mode — ignore
  }
}

/** Parse BRL display/input ("4,29" / "R$ 4,29") to integer cents. */
export function brlToCents(value) {
  if (value == null || value === '') return 0
  if (typeof value === 'number') {
    return Number.isFinite(value) ? Math.round(value) : 0
  }
  const s = String(value)
    .replace(/[^\d,.-]/g, '')
    .trim()
  if (!s) return 0
  if (s.includes(',')) {
    const reais = parseFloat(s.replace(/\./g, '').replace(',', '.'))
    return Number.isFinite(reais) ? Math.round(reais * 100) : 0
  }
  const n = Number(s)
  return Number.isFinite(n) ? Math.round(n) : 0
}

/** @param {object} item */
export function itemCostCents(item) {
  if (item == null) return 0
  if (item.unitCostCents != null && item.unitCostCents !== '') {
    const n = Number(item.unitCostCents)
    if (Number.isFinite(n)) return Math.round(n)
  }
  return brlToCents(item.unitCostRaw || item.unitCost)
}

/** @param {object} item */
export function itemSaleCents(item) {
  if (item == null) return 0
  if (item.salePriceCents != null && item.salePriceCents !== '') {
    const n = Number(item.salePriceCents)
    if (Number.isFinite(n)) return Math.round(n)
  }
  const raw = item.salePriceRaw || item.salePriceHint
  if (raw == null || raw === '') return 0
  return brlToCents(raw)
}

/** @param {object} item */
export function itemMarginCents(item) {
  const sale = itemSaleCents(item)
  if (!sale) return 0
  return sale - itemCostCents(item)
}

/** @param {object} item */
export function lotItemStatusKey(item) {
  const raw = String(item?.status || item?.statusLabel || '')
    .trim()
    .toLowerCase()
  if (raw === 'em estoque') return 'in_stock'
  if (raw === 'reservado') return 'reserved'
  if (raw === 'vendido') return 'sold'
  return raw
}

/**
 * @param {object} a
 * @param {object} b
 * @param {string} key
 * @param {'asc'|'desc'} [dir]
 */
export function compareLotItemSort(a, b, key, dir = 'asc') {
  const d = dir === 'desc' ? -1 : 1
  switch (key) {
    case 'id':
      return ((Number(a.id) || 0) - (Number(b.id) || 0)) * d
    case 'unitCost':
      return (itemCostCents(a) - itemCostCents(b)) * d
    case 'salePrice':
      return (itemSaleCents(a) - itemSaleCents(b)) * d
    case 'margin':
      return (itemMarginCents(a) - itemMarginCents(b)) * d
    case 'status': {
      const ar = STATUS_RANK[lotItemStatusKey(a)] ?? 99
      const br = STATUS_RANK[lotItemStatusKey(b)] ?? 99
      return (ar - br) * d
    }
    case 'sku':
      return String(a.sku || '').localeCompare(String(b.sku || ''), 'pt-BR') * d
    case 'title':
    default:
      return String(a.title || '').localeCompare(String(b.title || ''), 'pt-BR') * d
  }
}

/**
 * @param {string} key
 * @returns {'asc'|'desc'}
 */
export function defaultLotItemSortDir(key) {
  if (key === 'title' || key === 'sku' || key === 'status' || key === 'id') return 'asc'
  return 'desc'
}

/**
 * @param {LotItemSort[]|null|undefined} sorts
 * @param {string} key
 * @returns {LotItemSort[]}
 */
export function stackLotItemSort(sorts, key) {
  const list = Array.isArray(sorts)
    ? sorts
        .filter((s) => s && s.key)
        .map((s) => ({
          key: String(s.key),
          dir: s.dir === 'desc' ? 'desc' : 'asc',
        }))
    : []
  const i = list.findIndex((s) => s.key === key)
  if (i >= 0) {
    const dir = list[i].dir === 'asc' ? 'desc' : 'asc'
    list.splice(i, 1)
    list.unshift({ key, dir })
    return list
  }
  if (list.length === 1 && list[0].key === 'id' && key !== 'id') {
    return [{ key, dir: defaultLotItemSortDir(key) }]
  }
  list.unshift({ key, dir: defaultLotItemSortDir(key) })
  if (list.length > 6) list.length = 6
  return list
}

/**
 * @param {{ sorts?: LotItemSort[], sortKey?: string, sortDir?: string }} opts
 * @returns {LotItemSort[]}
 */
export function normalizeLotItemSorts(opts = {}) {
  if (Array.isArray(opts.sorts) && opts.sorts.length) {
    return opts.sorts
      .filter((s) => s && s.key)
      .map((s) => ({
        key: String(s.key),
        dir: s.dir === 'desc' ? 'desc' : 'asc',
      }))
  }
  if (opts.sortKey) {
    return [
      {
        key: opts.sortKey,
        dir:
          opts.sortDir === 'asc'
            ? 'asc'
            : opts.sortDir === 'desc'
              ? 'desc'
              : defaultLotItemSortDir(opts.sortKey),
      },
    ]
  }
  return defaultLotItemSorts()
}

/**
 * @param {object[]} items
 * @param {{
 *   query?: string,
 *   filterStatus?: 'all'|'in_stock'|'reserved'|'sold',
 *   sorts?: LotItemSort[],
 *   sortKey?: string,
 *   sortDir?: 'asc'|'desc',
 * }} opts
 */
export function filterAndSortLotItems(items, opts = {}) {
  const list = Array.isArray(items) ? items : []
  const query = String(opts.query || '')
    .trim()
    .toLowerCase()
  const filterStatus = opts.filterStatus || 'all'
  const sorts = normalizeLotItemSorts(opts)

  const filtered = list.filter((item) => {
    if (filterStatus !== 'all' && lotItemStatusKey(item) !== filterStatus) return false
    if (!query) return true
    const hay = [item.id, item.title, item.sku, item.status, item.statusLabel, item.unitCost, item.salePriceHint]
      .filter((v) => v != null && v !== '')
      .join(' ')
      .toLowerCase()
    return hay.includes(query)
  })

  return filtered.slice().sort((a, b) => {
    for (const s of sorts) {
      const c = compareLotItemSort(a, b, s.key, s.dir)
      if (c !== 0) return c
    }
    return compareLotItemSort(a, b, 'id', 'asc')
  })
}
