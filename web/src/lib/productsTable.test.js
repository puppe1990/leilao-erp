import { describe, expect, it } from 'vitest'
import {
  brandsFromProducts,
  filterAndSortProducts,
} from './productsTable.js'

const sample = [
  {
    id: 1,
    name: 'Monitor Dell P2219H 22"',
    kind: 'principal',
    qtyInStock: 1,
    salePriceRaw: '279,00',
    salePriceHint: 'R$ 279,00',
    photoCount: 3,
    videoCount: 1,
  },
  {
    id: 2,
    name: 'Monitor Samsung 733NW 17" (sem base)',
    kind: 'principal',
    qtyInStock: 0,
    salePriceRaw: '',
    salePriceHint: '',
    photoCount: 0,
    videoCount: 0,
  },
  {
    id: 3,
    name: 'Cabo HDMI 1,5m',
    kind: 'accessory',
    qtyInStock: 5,
    salePriceRaw: '19,00',
    salePriceHint: 'R$ 19,00',
    photoCount: 0,
    videoCount: 0,
  },
]

describe('filterAndSortProducts', () => {
  it('filters accessories', () => {
    const out = filterAndSortProducts(sample, { filterType: 'accessory' })
    expect(out.map((x) => x.id)).toEqual([3])
  })

  it('filters priced / unpriced', () => {
    expect(filterAndSortProducts(sample, { filterPrice: 'unpriced' }).map((x) => x.id)).toEqual([
      2,
    ])
    expect(filterAndSortProducts(sample, { filterPrice: 'priced' }).map((x) => x.id)).toEqual([
      3, 1,
    ]) // sorted by name default: Cabo, Dell
  })

  it('filters stock and media', () => {
    expect(filterAndSortProducts(sample, { filterStock: 'out' }).map((x) => x.id)).toEqual([2])
    expect(filterAndSortProducts(sample, { filterMedia: 'photo' }).map((x) => x.id)).toEqual([1])
    expect(filterAndSortProducts(sample, { filterMedia: 'none' }).map((x) => x.id)).toEqual([
      3, 2,
    ])
  })

  it('filters brand and base', () => {
    expect(filterAndSortProducts(sample, { filterBrand: 'Dell' }).map((x) => x.id)).toEqual([1])
    expect(filterAndSortProducts(sample, { filterBase: 'without' }).map((x) => x.id)).toEqual([2])
  })

  it('sorts by price and qty', () => {
    const byPrice = filterAndSortProducts(sample, {
      filterType: 'main',
      sortKey: 'salePrice',
      sortDir: 'asc',
    })
    // Dell 279, Samsung null last
    expect(byPrice.map((x) => x.id)).toEqual([1, 2])

    const byQty = filterAndSortProducts(sample, { sortKey: 'qty', sortDir: 'desc' })
    expect(byQty.map((x) => x.id)).toEqual([3, 1, 2])
  })

  it('searches description fields via query', () => {
    const withDesc = [
      ...sample,
      {
        id: 9,
        name: 'Monitor X',
        kind: 'principal',
        description: 'painel IPS Full HD',
        qtyInStock: 1,
      },
    ]
    const out = filterAndSortProducts(withDesc, { query: 'ips' })
    expect(out.map((x) => x.id)).toEqual([9])
  })
})

describe('brandsFromProducts', () => {
  it('lists unique brands', () => {
    expect(brandsFromProducts(sample)).toEqual(['Cabo', 'Dell', 'Samsung'])
  })
})
