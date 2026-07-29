import { beforeEach, describe, expect, it, vi } from 'vitest'
import {
  cartTotals,
  formatCartBRL,
  itemPixCents,
  productToCartSeed,
  shopCart,
} from './shopCart.js'
import { get } from 'svelte/store'

const store = new Map()

beforeEach(() => {
  store.clear()
  vi.stubGlobal('localStorage', {
    getItem: (k) => (store.has(k) ? store.get(k) : null),
    setItem: (k, v) => store.set(k, String(v)),
    removeItem: (k) => store.delete(k),
  })
  shopCart.clear()
})

describe('productToCartSeed', () => {
  it('maps product fields', () => {
    const seed = productToCartSeed({
      id: 7,
      name: 'Dell P2219H',
      price: 'R$ 300,00',
      priceCents: 30000,
      pixPrice: 'R$ 270,00',
      thumbUrl: '/x.jpg',
      qtyInStock: 3,
    })
    expect(seed).toMatchObject({
      id: 7,
      name: 'Dell P2219H',
      priceCents: 30000,
      maxQty: 3,
    })
  })

  it('returns null without id', () => {
    expect(productToCartSeed({})).toBeNull()
  })
})

describe('shopCart store', () => {
  it('adds and merges quantities up to max', () => {
    const p = {
      id: 1,
      name: 'Monitor A',
      priceCents: 10000,
      qtyInStock: 2,
    }
    expect(shopCart.add(p, 1)).toBe(true)
    expect(shopCart.add(p, 1)).toBe(true)
    expect(shopCart.add(p, 1)).toBe(false) // already at max 2
    const items = get(shopCart)
    expect(items).toHaveLength(1)
    expect(items[0].qty).toBe(2)
  })

  it('setQty removes at zero', () => {
    shopCart.add({ id: 2, name: 'B', priceCents: 5000, qtyInStock: 5 }, 2)
    shopCart.setQty(2, 0)
    expect(get(shopCart)).toHaveLength(0)
  })

  it('persists to localStorage', () => {
    shopCart.add({ id: 9, name: 'C', priceCents: 1000, qtyInStock: 1 }, 1)
    const raw = store.get('puppe-shop-cart-v1')
    expect(JSON.parse(raw)[0].id).toBe(9)
    expect(JSON.parse(raw)[0].name).toBe('C')
  })
})

describe('cartTotals', () => {
  it('sums units and PIX/card cents', () => {
    const items = [
      { id: 1, name: 'A', qty: 2, priceCents: 10000, maxQty: 5 },
      { id: 2, name: 'B', qty: 1, priceCents: 5000, maxQty: 5 },
    ]
    const t = cartTotals(items)
    expect(t.units).toBe(3)
    expect(t.cardCents).toBe(25000)
    // 2*9000 + 4500
    expect(t.pixCents).toBe(22500)
    expect(formatCartBRL(t.pixCents)).toContain('225')
  })

  it('itemPixCents applies 10% off', () => {
    expect(itemPixCents({ qty: 2, priceCents: 10000 })).toBe(18000)
  })
})
