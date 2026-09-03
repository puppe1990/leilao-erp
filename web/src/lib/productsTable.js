/**
 * Pure helpers for Products catalog table: filter + multi-column sort.
 * Reuses brand/base tags from stock helpers for consistent labels.
 */

import { baseOfTitle, brandOfTitle, salePriceCents } from './stockInventoryTable.js'

export { baseOfTitle, brandOfTitle, salePriceCents }

/**
 * @typedef {{ key: string, dir: 'asc'|'desc' }} ProductSort
 */

export const PRODUCT_SORTS_STORAGE_KEY = 'leilao-erp.products.sorts.v1'

const PRODUCT_SORT_KEYS = new Set([
  'name',
  'type',
  'qty',
  'media',
  'salePrice',
  'catalog',
  'olx',
  'freteOlx',
  'id',
])

/** Default when nothing is stored. */
export function defaultProductSorts() {
  return [{ key: 'name', dir: 'asc' }]
}

/**
 * Validate/normalize a sorts array from storage or UI.
 * @param {unknown} raw
 * @returns {ProductSort[]}
 */
export function sanitizeProductSorts(raw) {
  if (!Array.isArray(raw) || !raw.length) return defaultProductSorts()
  const out = []
  const seen = new Set()
  for (const item of raw) {
    if (!item || typeof item !== 'object') continue
    const key = String(/** @type {{ key?: string }} */ (item).key || '')
    if (!PRODUCT_SORT_KEYS.has(key) || seen.has(key)) continue
    seen.add(key)
    const dir = /** @type {{ dir?: string }} */ (item).dir === 'desc' ? 'desc' : 'asc'
    out.push({ key, dir })
    if (out.length >= 6) break
  }
  return out.length ? out : defaultProductSorts()
}

/**
 * @param {Storage | null | undefined} storage
 * @returns {ProductSort[]}
 */
export function loadProductSorts(storage = typeof localStorage !== 'undefined' ? localStorage : null) {
  if (!storage) return defaultProductSorts()
  try {
    const raw = storage.getItem(PRODUCT_SORTS_STORAGE_KEY)
    if (!raw) return defaultProductSorts()
    return sanitizeProductSorts(JSON.parse(raw))
  } catch {
    return defaultProductSorts()
  }
}

/**
 * @param {ProductSort[]} sorts
 * @param {Storage | null | undefined} storage
 */
export function saveProductSorts(
  sorts,
  storage = typeof localStorage !== 'undefined' ? localStorage : null,
) {
  if (!storage) return
  try {
    storage.setItem(PRODUCT_SORTS_STORAGE_KEY, JSON.stringify(sanitizeProductSorts(sorts)))
  } catch {
    // quota / private mode — ignore
  }
}

/**
 * Compare two products by a single sort key.
 * @param {object} a
 * @param {object} b
 * @param {string} key
 * @param {'asc'|'desc'} [dir]
 */
export function compareProductSort(a, b, key, dir = 'asc') {
  const d = dir === 'desc' ? -1 : 1
  switch (key) {
    case 'type': {
      const ai = a.kind === 'accessory' || a.isAccessory ? 1 : 0
      const bi = b.kind === 'accessory' || b.isAccessory ? 1 : 0
      return (ai - bi) * d
    }
    case 'qty':
      return ((Number(a.qtyInStock) || 0) - (Number(b.qtyInStock) || 0)) * d
    case 'media': {
      const am = (Number(a.photoCount) || 0) + (Number(a.videoCount) || 0)
      const bm = (Number(b.photoCount) || 0) + (Number(b.videoCount) || 0)
      return (am - bm) * d
    }
    case 'salePrice': {
      const av = salePriceCents(a)
      const bv = salePriceCents(b)
      if (av == null && bv == null) return 0
      if (av == null) return 1
      if (bv == null) return -1
      return (av - bv) * d
    }
    case 'olx': {
      const ai = a.olxPublished ? 1 : 0
      const bi = b.olxPublished ? 1 : 0
      return (ai - bi) * d
    }
    case 'catalog': {
      const ai = a.shopVisible ? 1 : 0
      const bi = b.shopVisible ? 1 : 0
      return (ai - bi) * d
    }
    case 'freteOlx': {
      const ai = a.olxFreeShipping ? 1 : 0
      const bi = b.olxFreeShipping ? 1 : 0
      return (ai - bi) * d
    }
    case 'id':
      return ((Number(a.id) || 0) - (Number(b.id) || 0)) * d
    case 'name':
    default:
      return (a.name || a.title || '').localeCompare(b.name || b.title || '', 'pt-BR') * d
  }
}

/**
 * Default direction when a column is first added to the sort stack.
 * @param {string} key
 * @returns {'asc'|'desc'}
 */
export function defaultProductSortDir(key) {
  return key === 'name' || key === 'type' ? 'asc' : 'desc'
}

/**
 * Stackable sort (last click wins as primary; previous keys stay as tie-breakers).
 * - New column → becomes rank 1 (others shift down, not removed).
 * - Same column again → toggles dir and stays / moves to rank 1.
 * @param {ProductSort[]|null|undefined} sorts
 * @param {string} key
 * @returns {ProductSort[]}
 */
export function stackProductSort(sorts, key) {
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
  // First intentional sort: drop default "name" so it doesn't swallow every order.
  if (list.length === 1 && list[0].key === 'name' && key !== 'name') {
    return [{ key, dir: defaultProductSortDir(key) }]
  }
  list.unshift({ key, dir: defaultProductSortDir(key) })
  // Keep stack bounded (still multi, not infinite).
  if (list.length > 6) list.length = 6
  return list
}

/**
 * Normalize sort list from opts (sorts[] or legacy sortKey/sortDir).
 * @param {{ sorts?: ProductSort[], sortKey?: string, sortDir?: string }} opts
 * @returns {ProductSort[]}
 */
export function normalizeProductSorts(opts = {}) {
  if (Array.isArray(opts.sorts) && opts.sorts.length) {
    return opts.sorts
      .filter((s) => s && s.key)
      .map((s) => ({
        key: String(s.key),
        dir: s.dir === 'desc' ? 'desc' : 'asc',
      }))
  }
  return [
    {
      key: opts.sortKey || 'name',
      dir: opts.sortDir === 'desc' ? 'desc' : 'asc',
    },
  ]
}

/**
 * @param {object[]} products
 * @param {{
 *   query?: string,
 *   filterType?: 'all'|'main'|'accessory',
 *   filterPrice?: 'all'|'priced'|'unpriced',
 *   filterStock?: 'all'|'in_stock'|'out',
 *   filterMedia?: 'all'|'photo'|'video'|'any'|'none',
 *   filterBase?: 'all'|'with'|'without'|'unset',
 *   filterBrand?: string,
 *   filterOlx?: 'all'|'published'|'unpublished',
 *   sorts?: ProductSort[],
 *   sortKey?: string,
 *   sortDir?: 'asc'|'desc',
 * }} opts
 */
export function filterAndSortProducts(products, opts = {}) {
  const list = Array.isArray(products) ? products : []
  const query = String(opts.query || '')
    .trim()
    .toLowerCase()
  const filterType = opts.filterType || 'all'
  const filterPrice = opts.filterPrice || 'all'
  const filterStock = opts.filterStock || 'all'
  const filterMedia = opts.filterMedia || 'all'
  const filterBase = opts.filterBase || 'all'
  const filterBrand = opts.filterBrand || 'all'
  const filterOlx = opts.filterOlx || 'all'
  const sorts = normalizeProductSorts(opts)

  const filtered = list.filter((p) => {
    const isAcc = p.kind === 'accessory' || p.isAccessory === true
    if (filterType === 'main' && isAcc) return false
    if (filterType === 'accessory' && !isAcc) return false

    const priced = !!(p.salePriceHint || (salePriceCents(p) != null && salePriceCents(p) > 0))
    if (filterPrice === 'priced' && !priced) return false
    if (filterPrice === 'unpriced' && priced) return false

    const qty = Number(p.qtyInStock) || 0
    if (filterStock === 'in_stock' && qty <= 0) return false
    if (filterStock === 'out' && qty > 0) return false

    const photos = Number(p.photoCount) || 0
    const videos = Number(p.videoCount) || 0
    if (filterMedia === 'photo' && photos <= 0) return false
    if (filterMedia === 'video' && videos <= 0) return false
    if (filterMedia === 'any' && photos + videos <= 0) return false
    if (filterMedia === 'none' && photos + videos > 0) return false

    const olxPub = !!p.olxPublished
    if (filterOlx === 'published' && !olxPub) return false
    if (filterOlx === 'unpublished' && olxPub) return false

    const title = p.name || p.title || ''
    if (filterBase !== 'all' && baseOfTitle(title) !== filterBase) return false
    if (filterBrand !== 'all' && brandOfTitle(title) !== filterBrand) return false

    if (!query) return true
    const hay = [
      p.id,
      p.name,
      p.title,
      p.kind,
      p.kindLabel,
      p.description,
      p.listingText,
      p.salePriceHint,
      p.qtyInStock,
      brandOfTitle(title),
      isAcc ? 'acessorio cabo' : 'principal monitor',
      baseOfTitle(title) === 'with' ? 'com base' : '',
      baseOfTitle(title) === 'without' ? 'sem base' : '',
      photos ? 'foto' : '',
      videos ? 'video' : '',
      olxPub ? 'publicado olx no ar' : 'falta publicar olx',
    ]
      .filter(Boolean)
      .join(' ')
      .toLowerCase()
    return hay.includes(query)
  })

  return filtered.slice().sort((a, b) => {
    for (const s of sorts) {
      const c = compareProductSort(a, b, s.key, s.dir)
      if (c !== 0) return c
    }
    // Stable-ish fallback
    return compareProductSort(a, b, 'id', 'asc')
  })
}

/** Unique sorted brand labels from product names. */
export function brandsFromProducts(products) {
  return [
    ...new Set(
      (products || []).map((p) => brandOfTitle(p.name || p.title)).filter(Boolean),
    ),
  ].sort((a, b) => a.localeCompare(b, 'pt-BR'))
}
