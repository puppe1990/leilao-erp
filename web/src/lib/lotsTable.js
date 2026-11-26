/**
 * Pure helpers for Lots index table: filter + multi-column sort.
 */

/**
 * @typedef {{ key: string, dir: 'asc'|'desc' }} LotSort
 */

export const LOT_SORTS_STORAGE_KEY = 'leilao-erp.lots.sorts.v1'

const LOT_SORT_KEYS = new Set(['name', 'purchasedAt', 'status', 'itemCount', 'totalCost', 'id'])

const STATUS_RANK = {
  open: 0,
  aberto: 0,
  partial: 1,
  parcial: 1,
  sold: 2,
  vendido: 2,
  closed: 3,
  fechado: 3,
}

/** Default when nothing is stored. */
export function defaultLotSorts() {
  return [{ key: 'purchasedAt', dir: 'desc' }]
}

/**
 * @param {unknown} raw
 * @returns {LotSort[]}
 */
export function sanitizeLotSorts(raw) {
  if (!Array.isArray(raw) || !raw.length) return defaultLotSorts()
  const out = []
  const seen = new Set()
  for (const item of raw) {
    if (!item || typeof item !== 'object') continue
    const key = String(/** @type {{ key?: string }} */ (item).key || '')
    if (!LOT_SORT_KEYS.has(key) || seen.has(key)) continue
    seen.add(key)
    const dir = /** @type {{ dir?: string }} */ (item).dir === 'desc' ? 'desc' : 'asc'
    out.push({ key, dir })
    if (out.length >= 6) break
  }
  return out.length ? out : defaultLotSorts()
}

/**
 * @param {Storage | null | undefined} storage
 * @returns {LotSort[]}
 */
export function loadLotSorts(storage = typeof localStorage !== 'undefined' ? localStorage : null) {
  if (!storage) return defaultLotSorts()
  try {
    const raw = storage.getItem(LOT_SORTS_STORAGE_KEY)
    if (!raw) return defaultLotSorts()
    return sanitizeLotSorts(JSON.parse(raw))
  } catch {
    return defaultLotSorts()
  }
}

/**
 * @param {LotSort[]} sorts
 * @param {Storage | null | undefined} storage
 */
export function saveLotSorts(
  sorts,
  storage = typeof localStorage !== 'undefined' ? localStorage : null,
) {
  if (!storage) return
  try {
    storage.setItem(LOT_SORTS_STORAGE_KEY, JSON.stringify(sanitizeLotSorts(sorts)))
  } catch {
    // quota / private mode — ignore
  }
}

/** @param {object} lot */
export function lotCostCents(lot) {
  if (lot == null) return 0
  if (lot.totalCostCents != null && lot.totalCostCents !== '') {
    const n = Number(lot.totalCostCents)
    if (Number.isFinite(n)) return Math.round(n)
  }
  const s = String(lot.totalCost || '')
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

/** @param {object} lot */
export function lotStatusKey(lot) {
  const raw = String(lot?.status || lot?.statusLabel || '')
    .trim()
    .toLowerCase()
  if (raw in STATUS_RANK) {
    if (raw === 'aberto') return 'open'
    if (raw === 'parcial') return 'partial'
    if (raw === 'vendido') return 'sold'
    if (raw === 'fechado') return 'closed'
    return raw
  }
  return raw
}

/**
 * @param {object} a
 * @param {object} b
 * @param {string} key
 * @param {'asc'|'desc'} [dir]
 */
export function compareLotSort(a, b, key, dir = 'asc') {
  const d = dir === 'desc' ? -1 : 1
  switch (key) {
    case 'itemCount':
      return ((Number(a.itemCount) || 0) - (Number(b.itemCount) || 0)) * d
    case 'totalCost':
      return (lotCostCents(a) - lotCostCents(b)) * d
    case 'purchasedAt':
      return String(a.purchasedAt || '').localeCompare(String(b.purchasedAt || '')) * d
    case 'status': {
      const ar = STATUS_RANK[lotStatusKey(a)] ?? 99
      const br = STATUS_RANK[lotStatusKey(b)] ?? 99
      return (ar - br) * d
    }
    case 'id':
      return ((Number(a.id) || 0) - (Number(b.id) || 0)) * d
    case 'name':
    default:
      return (a.name || '').localeCompare(b.name || '', 'pt-BR') * d
  }
}

/**
 * @param {string} key
 * @returns {'asc'|'desc'}
 */
export function defaultLotSortDir(key) {
  return key === 'name' || key === 'status' ? 'asc' : 'desc'
}

/**
 * @param {LotSort[]|null|undefined} sorts
 * @param {string} key
 * @returns {LotSort[]}
 */
export function stackLotSort(sorts, key) {
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
  if (list.length === 1 && list[0].key === 'purchasedAt' && key !== 'purchasedAt') {
    return [{ key, dir: defaultLotSortDir(key) }]
  }
  list.unshift({ key, dir: defaultLotSortDir(key) })
  if (list.length > 6) list.length = 6
  return list
}

/**
 * @param {{ sorts?: LotSort[], sortKey?: string, sortDir?: string }} opts
 * @returns {LotSort[]}
 */
export function normalizeLotSorts(opts = {}) {
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
        dir: opts.sortDir === 'asc' ? 'asc' : opts.sortDir === 'desc' ? 'desc' : defaultLotSortDir(opts.sortKey),
      },
    ]
  }
  return defaultLotSorts()
}

/**
 * @param {object[]} lots
 * @param {{
 *   query?: string,
 *   filterStatus?: 'all'|'open'|'partial'|'sold'|'closed',
 *   sorts?: LotSort[],
 *   sortKey?: string,
 *   sortDir?: 'asc'|'desc',
 * }} opts
 */
export function filterAndSortLots(lots, opts = {}) {
  const list = Array.isArray(lots) ? lots : []
  const query = String(opts.query || '')
    .trim()
    .toLowerCase()
  const filterStatus = opts.filterStatus || 'all'
  const sorts = normalizeLotSorts(opts)

  const filtered = list.filter((lot) => {
    if (filterStatus !== 'all' && lotStatusKey(lot) !== filterStatus) return false
    if (!query) return true
    const hay = [
      lot.id,
      lot.name,
      lot.purchasedAt,
      lot.status,
      lot.statusLabel,
      lot.totalCost,
      lot.itemCount,
    ]
      .filter((v) => v != null && v !== '')
      .join(' ')
      .toLowerCase()
    return hay.includes(query)
  })

  return filtered.slice().sort((a, b) => {
    for (const s of sorts) {
      const c = compareLotSort(a, b, s.key, s.dir)
      if (c !== 0) return c
    }
    const byDate = compareLotSort(a, b, 'purchasedAt', 'desc')
    if (byDate !== 0) return byDate
    return compareLotSort(a, b, 'id', 'desc')
  })
}
