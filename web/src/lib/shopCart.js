/**
 * Client-side shop cart (localStorage). Shared across catalog and product pages.
 */
import { derived, get, writable } from 'svelte/store'

const STORAGE_KEY = 'puppe-shop-cart-v1'

/**
 * @typedef {{
 *   id: number|string,
 *   name: string,
 *   price?: string,
 *   priceCents?: number,
 *   pixPrice?: string,
 *   thumbUrl?: string,
 *   qty: number,
 *   maxQty: number,
 * }} CartItem
 */

/** @returns {CartItem[]} */
function load() {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return []
    const parsed = JSON.parse(raw)
    if (!Array.isArray(parsed)) return []
    return parsed
      .filter((it) => it && it.id != null && it.name && Number(it.qty) > 0)
      .map((it) => ({
        id: it.id,
        name: String(it.name),
        price: it.price || '',
        priceCents: Number(it.priceCents) || 0,
        pixPrice: it.pixPrice || '',
        thumbUrl: it.thumbUrl || '',
        qty: Math.max(1, Math.floor(Number(it.qty) || 1)),
        maxQty: Math.max(1, Math.floor(Number(it.maxQty) || 99)),
      }))
  } catch {
    return []
  }
}

/** @param {CartItem[]} items */
function persist(items) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(items))
  } catch {
    // ignore quota / private mode
  }
}

/**
 * Format cents as BRL (pt-BR).
 * @param {number} cents
 */
export function formatCartBRL(cents) {
  const n = Number(cents) || 0
  return (n / 100).toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })
}

/**
 * PIX line cents for an item (10% off when price known).
 * @param {CartItem} item
 */
export function itemPixCents(item) {
  const unit = Number(item.priceCents) || 0
  if (unit <= 0) return 0
  return Math.floor((unit * 90) / 100) * (Number(item.qty) || 1)
}

/**
 * Card/list price cents for an item.
 * @param {CartItem} item
 */
export function itemCardCents(item) {
  return (Number(item.priceCents) || 0) * (Number(item.qty) || 1)
}

/**
 * @param {CartItem[]} items
 */
export function cartTotals(items) {
  let units = 0
  let cardCents = 0
  let pixCents = 0
  let priced = 0
  for (const it of items || []) {
    const q = Number(it.qty) || 0
    units += q
    if ((Number(it.priceCents) || 0) > 0) {
      priced += q
      cardCents += itemCardCents(it)
      pixCents += itemPixCents(it)
    }
  }
  return { units, cardCents, pixCents, priced }
}

/**
 * Normalize a catalog/detail product into a cart line seed.
 * @param {Record<string, any>} product
 * @returns {Omit<CartItem, 'qty'> | null}
 */
export function productToCartSeed(product) {
  if (!product || product.id == null) return null
  const maxQty = Math.max(1, Math.floor(Number(product.qtyInStock) || 1))
  return {
    id: product.id,
    name: String(product.name || 'Produto'),
    price: product.price || '',
    priceCents: Number(product.priceCents) || 0,
    pixPrice: product.pixPrice || '',
    thumbUrl: product.thumbUrl || '',
    maxQty,
  }
}

function createCartStore() {
  const store = writable(/** @type {CartItem[]} */ ([]))
  const { subscribe, set, update } = store
  let hydrated = false

  function hydrate() {
    if (hydrated || typeof localStorage === 'undefined') return
    hydrated = true
    set(load())
  }

  function write(items) {
    persist(items)
    set(items)
  }

  return {
    subscribe,
    hydrate,
    /** @param {Record<string, any>} product @param {number} [qty] */
    add(product, qty = 1) {
      hydrate()
      const seed = productToCartSeed(product)
      if (!seed) return false
      const addQty = Math.max(1, Math.floor(Number(qty) || 1))
      let ok = false
      update((items) => {
        const next = items.slice()
        const i = next.findIndex((x) => String(x.id) === String(seed.id))
        if (i >= 0) {
          const cur = next[i]
          const max = Math.max(1, Number(cur.maxQty) || seed.maxQty || 1)
          const newQty = Math.min(max, cur.qty + addQty)
          if (newQty === cur.qty) {
            ok = false
            return items
          }
          next[i] = {
            ...cur,
            ...seed,
            maxQty: max,
            qty: newQty,
          }
          ok = true
        } else {
          next.push({
            ...seed,
            qty: Math.min(seed.maxQty, addQty),
          })
          ok = true
        }
        persist(next)
        return next
      })
      return ok
    },
    /** @param {number|string} id @param {number} qty */
    setQty(id, qty) {
      hydrate()
      update((items) => {
        const next = items
          .map((it) => {
            if (String(it.id) !== String(id)) return it
            const max = Math.max(1, Number(it.maxQty) || 1)
            const q = Math.floor(Number(qty) || 0)
            if (q <= 0) return null
            return { ...it, qty: Math.min(max, q) }
          })
          .filter(Boolean)
        persist(next)
        return next
      })
    },
    /** @param {number|string} id */
    remove(id) {
      hydrate()
      update((items) => {
        const next = items.filter((it) => String(it.id) !== String(id))
        persist(next)
        return next
      })
    },
    clear() {
      hydrate()
      write([])
    },
    /** @param {number|string} id */
    getQty(id) {
      hydrate()
      const it = get(store).find((x) => String(x.id) === String(id))
      return it ? it.qty : 0
    },
  }
}

export const shopCart = createCartStore()

export const cartCount = derived(shopCart, ($items) =>
  ($items || []).reduce((n, it) => n + (Number(it.qty) || 0), 0),
)

export const cartSummary = derived(shopCart, ($items) => cartTotals($items || []))
