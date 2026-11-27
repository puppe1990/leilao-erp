/**
 * Pure helpers for the Finds table: multi-column sort with localStorage.
 */

/**
 * @typedef {{ key: string, dir: 'asc'|'desc' }} FindSort
 */

export const FIND_SORTS_STORAGE_KEY = 'leilao-erp.finds.sorts.v1'

const FIND_SORT_KEYS = new Set(['name', 'source', 'buy', 'sell', 'margin', 'id'])

/** Default when nothing is stored: best margin first. */
export function defaultFindSorts() {
  return [{ key: 'margin', dir: 'desc' }]
}

/** Parse "R$ 1.204,25", "1204,25", "-R$ 50,00" or cents into integer cents. */
export function brlCents(value) {
  if (value == null || value === '') return 0
  if (typeof value === 'number') return Number.isFinite(value) ? Math.round(value) : 0
  const cleaned = String(value).replace(/[^\d,.-]/g, '')
  if (!cleaned || cleaned === '-') return 0
  const normalized = cleaned.includes(',')
    ? cleaned.replace(/\./g, '').replace(',', '.')
    : cleaned
  const n = Number(normalized)
  return Number.isFinite(n) ? Math.round(n * 100) : 0
}

/** @param {object} find */
export function findBuyCents(find) {
  if (find?.buyCents != null) return brlCents(find.buyCents)
  return brlCents(find?.buyRaw || find?.buy)
}

/** @param {object} find */
export function findSellCents(find) {
  if (find?.sellCents != null) return brlCents(find.sellCents)
  return brlCents(find?.sellRaw || find?.sell)
}

/** @param {object} find */
export function findMarginCents(find) {
  if (find?.margin != null) return brlCents(find.margin)
  return findSellCents(find) - findBuyCents(find)
}

/**
 * @param {object} a
 * @param {object} b
 * @param {string} key
 * @param {'asc'|'desc'} [dir]
 */
export function compareFindSort(a, b, key, dir = 'asc') {
  const d = dir === 'desc' ? -1 : 1
  switch (key) {
    case 'buy':
      return (findBuyCents(a) - findBuyCents(b)) * d
    case 'sell':
      return (findSellCents(a) - findSellCents(b)) * d
    case 'margin':
      return (findMarginCents(a) - findMarginCents(b)) * d
    case 'source':
      return String(a.source || '').localeCompare(String(b.source || ''), 'pt-BR') * d
    case 'id':
      return ((Number(a.id) || 0) - (Number(b.id) || 0)) * d
    case 'name':
    default:
      return String(a.name || '').localeCompare(String(b.name || ''), 'pt-BR') * d
  }
}

/** @param {string} key @returns {'asc'|'desc'} */
export function defaultFindSortDir(key) {
  return key === 'name' || key === 'source' ? 'asc' : 'desc'
}

/** @param {unknown} raw @returns {FindSort[]} */
export function sanitizeFindSorts(raw) {
  if (!Array.isArray(raw) || !raw.length) return defaultFindSorts()
  const out = []
  const seen = new Set()
  for (const item of raw) {
    if (!item || typeof item !== 'object') continue
    const key = String(/** @type {{ key?: string }} */ (item).key || '')
    if (!FIND_SORT_KEYS.has(key) || seen.has(key)) continue
    seen.add(key)
    out.push({ key, dir: /** @type {{ dir?: string }} */ (item).dir === 'desc' ? 'desc' : 'asc' })
    if (out.length >= 6) break
  }
  return out.length ? out : defaultFindSorts()
}

/** @param {Storage | null | undefined} storage @returns {FindSort[]} */
export function loadFindSorts(storage = typeof localStorage !== 'undefined' ? localStorage : null) {
  if (!storage) return defaultFindSorts()
  try {
    const raw = storage.getItem(FIND_SORTS_STORAGE_KEY)
    if (!raw) return defaultFindSorts()
    return sanitizeFindSorts(JSON.parse(raw))
  } catch {
    return defaultFindSorts()
  }
}

/** @param {FindSort[]} sorts @param {Storage | null | undefined} storage */
export function saveFindSorts(sorts, storage = typeof localStorage !== 'undefined' ? localStorage : null) {
  if (!storage) return
  try {
    storage.setItem(FIND_SORTS_STORAGE_KEY, JSON.stringify(sanitizeFindSorts(sorts)))
  } catch {
    // quota / private mode — ignore
  }
}

/**
 * @param {FindSort[]|null|undefined} sorts
 * @param {string} key
 * @returns {FindSort[]}
 */
export function stackFindSort(sorts, key) {
  const list = Array.isArray(sorts)
    ? sorts
        .filter((s) => s && s.key)
        .map((s) => ({ key: String(s.key), dir: s.dir === 'desc' ? 'desc' : 'asc' }))
    : []
  const i = list.findIndex((s) => s.key === key)
  if (i >= 0) {
    const dir = list[i].dir === 'asc' ? 'desc' : 'asc'
    list.splice(i, 1)
    list.unshift({ key, dir })
    return list
  }
  if (list.length === 1 && list[0].key === 'margin' && key !== 'margin') {
    return [{ key, dir: defaultFindSortDir(key) }]
  }
  list.unshift({ key, dir: defaultFindSortDir(key) })
  if (list.length > 6) list.length = 6
  return list
}

/**
 * @param {object[]} finds
 * @param {FindSort[]|null|undefined} sorts
 * @returns {object[]}
 */
export function sortFinds(finds, sorts) {
  const list = Array.isArray(finds) ? finds : []
  const order = sanitizeFindSorts(sorts)
  return list.slice().sort((a, b) => {
    for (const s of order) {
      const c = compareFindSort(a, b, s.key, s.dir)
      if (c !== 0) return c
    }
    return compareFindSort(a, b, 'id', 'desc')
  })
}
