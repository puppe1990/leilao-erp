/**
 * Pure helpers for Products catalog table: filter + sort.
 * Reuses brand/base tags from stock helpers for consistent labels.
 */

import { baseOfTitle, brandOfTitle, salePriceCents } from './stockInventoryTable.js'

export { baseOfTitle, brandOfTitle, salePriceCents }

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
  const sortKey = opts.sortKey || 'name'
  const sortDir = opts.sortDir === 'desc' ? 'desc' : 'asc'
  const dir = sortDir === 'asc' ? 1 : -1

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
      case 'type': {
        const ai = a.kind === 'accessory' || a.isAccessory ? 1 : 0
        const bi = b.kind === 'accessory' || b.isAccessory ? 1 : 0
        return (ai - bi) * dir
      }
      case 'qty':
        return ((Number(a.qtyInStock) || 0) - (Number(b.qtyInStock) || 0)) * dir
      case 'media': {
        const am = (Number(a.photoCount) || 0) + (Number(a.videoCount) || 0)
        const bm = (Number(b.photoCount) || 0) + (Number(b.videoCount) || 0)
        return (am - bm) * dir
      }
      case 'salePrice': {
        av = salePriceCents(a)
        bv = salePriceCents(b)
        if (av == null && bv == null) return 0
        if (av == null) return 1
        if (bv == null) return -1
        return (av - bv) * dir
      }
      case 'id':
        return ((Number(a.id) || 0) - (Number(b.id) || 0)) * dir
      case 'name':
      default:
        return (a.name || a.title || '').localeCompare(b.name || b.title || '', 'pt-BR') * dir
    }
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
