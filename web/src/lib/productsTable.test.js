import { describe, expect, it } from 'vitest'
import {
  brandsFromProducts,
  filterAndSortProducts,
  loadProductSorts,
  PRODUCT_SORTS_STORAGE_KEY,
  saveProductSorts,
  sanitizeProductSorts,
  stackProductSort,
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
    olxPublished: true,
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
    olxPublished: false,
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
    olxPublished: false,
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

  it('filters OLX published status', () => {
    expect(filterAndSortProducts(sample, { filterOlx: 'published' }).map((x) => x.id)).toEqual([1])
    expect(filterAndSortProducts(sample, { filterOlx: 'unpublished' }).map((x) => x.id)).toEqual([
      3, 2,
    ])
  })

  it('sorts catalog and frete OLX', () => {
    const withFlags = sample.map((p, i) => ({
      ...p,
      shopVisible: i === 1,
      olxFreeShipping: i === 2,
    }))
    const byCatalog = filterAndSortProducts(withFlags, { sortKey: 'catalog', sortDir: 'desc' })
    expect(byCatalog[0].shopVisible).toBe(true)
    const byFrete = filterAndSortProducts(withFlags, { sortKey: 'freteOlx', sortDir: 'desc' })
    expect(byFrete[0].olxFreeShipping).toBe(true)
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

  it('replaces default name so frete sort actually reorders', () => {
    let sorts = [{ key: 'name', dir: 'asc' }]
    sorts = stackProductSort(sorts, 'freteOlx')
    expect(sorts).toEqual([{ key: 'freteOlx', dir: 'desc' }])

    const rows = [
      { id: 1, name: 'AAA', olxFreeShipping: false },
      { id: 2, name: 'ZZZ', olxFreeShipping: true },
      { id: 3, name: 'MMM', olxFreeShipping: true },
    ]
    // If name stayed primary, order would be AAA,MMM,ZZZ — frete primary → Sim first
    const out = filterAndSortProducts(rows, { sorts })
    expect(out.map((x) => x.id)).toEqual([2, 3, 1])
  })

  it('stacks multi-column sorts: last click is primary, previous stay as tie-break', () => {
    let sorts = [{ key: 'name', dir: 'asc' }]
    sorts = stackProductSort(sorts, 'olx')
    sorts = stackProductSort(sorts, 'freteOlx')
    // frete is #1 (last click), olx is #2
    expect(sorts).toEqual([
      { key: 'freteOlx', dir: 'desc' },
      { key: 'olx', dir: 'desc' },
    ])
    sorts = stackProductSort(sorts, 'olx')
    // re-click olx → flip dir, move to #1
    expect(sorts).toEqual([
      { key: 'olx', dir: 'asc' },
      { key: 'freteOlx', dir: 'desc' },
    ])

    const rows = [
      { id: 1, name: 'A', olxPublished: true, olxFreeShipping: false },
      { id: 2, name: 'B', olxPublished: true, olxFreeShipping: true },
      { id: 3, name: 'C', olxPublished: false, olxFreeShipping: true },
    ]
    const out = filterAndSortProducts(rows, {
      sorts: [
        { key: 'olx', dir: 'desc' },
        { key: 'freteOlx', dir: 'desc' },
      ],
    })
    // published first, then frete among them
    expect(out.map((x) => x.id)).toEqual([2, 1, 3])
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

describe('product sorts localStorage', () => {
  it('sanitizes invalid keys and dirs', () => {
    expect(
      sanitizeProductSorts([
        { key: 'freteOlx', dir: 'desc' },
        { key: 'hack', dir: 'asc' },
        { key: 'olx', dir: 'nope' },
        { key: 'freteOlx', dir: 'asc' },
      ]),
    ).toEqual([
      { key: 'freteOlx', dir: 'desc' },
      { key: 'olx', dir: 'asc' },
    ])
  })

  it('round-trips through storage', () => {
    /** @type {Record<string, string>} */
    const mem = {}
    const storage = {
      getItem: (k) => (k in mem ? mem[k] : null),
      setItem: (k, v) => {
        mem[k] = String(v)
      },
    }
    const sorts = [
      { key: 'catalog', dir: 'desc' },
      { key: 'freteOlx', dir: 'asc' },
    ]
    saveProductSorts(sorts, storage)
    expect(mem[PRODUCT_SORTS_STORAGE_KEY]).toBeTruthy()
    expect(loadProductSorts(storage)).toEqual(sorts)
  })

  it('falls back when storage is empty or corrupt', () => {
    const storage = {
      getItem: () => '{not-json',
      setItem: () => {},
    }
    expect(loadProductSorts(storage)).toEqual([{ key: 'name', dir: 'asc' }])
    expect(loadProductSorts(null)).toEqual([{ key: 'name', dir: 'asc' }])
  })
})
