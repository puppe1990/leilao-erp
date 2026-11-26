import { describe, expect, it } from 'vitest'
import { sortShopCatalog } from './shopCatalog.js'

const cable = { id: 1, name: 'Cabo', category: 'acessorio', priceCents: 1000 }
const charger = { id: 2, name: 'C-43', category: 'acessorio', priceCents: 3500 }
const phone = { id: 3, name: 'S9+', category: 'com-base', priceCents: 34900 }
const monitor = { id: 4, name: 'Dell', category: 'com-base', priceCents: 27900 }
const broken = {
  id: 5,
  name: 'Peça',
  category: 'defeito',
  badge: 'DEFEITO',
  priceCents: 8000,
}

const catalog = [charger, phone, cable, monitor, broken]

describe('ordenação do catálogo da loja', () => {
  it('em maior preço coloca o item mais caro na frente, mesmo fora de acessório', () => {
    const names = sortShopCatalog(catalog, 'price-desc').map((p) => p.name)
    expect(names[0]).toBe('S9+')
    expect(names[1]).toBe('Dell')
    expect(names.at(-1)).toBe('Cabo')
  })

  it('em menor preço coloca o item mais barato na frente', () => {
    const names = sortShopCatalog(catalog, 'price-asc').map((p) => p.name)
    expect(names[0]).toBe('Cabo')
    expect(names.at(-1)).toBe('S9+')
  })

  it('em destaques mantém acessórios na frente e defeito no fim', () => {
    const names = sortShopCatalog(catalog, 'featured').map((p) => p.name)
    expect(names.slice(0, 2)).toEqual(['C-43', 'Cabo'])
    expect(names.at(-1)).toBe('Peça')
  })

  it('não altera a lista original', () => {
    const original = catalog.map((p) => p.id)
    sortShopCatalog(catalog, 'price-asc')
    expect(catalog.map((p) => p.id)).toEqual(original)
  })
})
