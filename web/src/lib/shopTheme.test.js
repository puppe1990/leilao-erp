import { beforeEach, describe, expect, it, vi } from 'vitest'
import { getShopTheme, setShopTheme } from './shopTheme.js'

const store = new Map()

beforeEach(() => {
  store.clear()
  vi.stubGlobal('localStorage', {
    getItem: (k) => (store.has(k) ? store.get(k) : null),
    setItem: (k, v) => store.set(k, String(v)),
    removeItem: (k) => store.delete(k),
  })
})

describe('shop theme storage', () => {
  it('grava em pix-eletro-shop-theme', () => {
    setShopTheme('light')
    expect(store.get('pix-eletro-shop-theme')).toBe('light')
  })

  it('lê o tema legado puppe-shop-theme quando a chave nova está vazia', () => {
    store.set('puppe-shop-theme', 'light')
    expect(getShopTheme()).toBe('light')
  })
})
