import { describe, expect, it } from 'vitest'
import {
  baseOfTitle,
  brandOfTitle,
  filterAndSortStockItems,
  marginCents,
  salePriceCents,
} from './stockInventoryTable.js'

describe('brandOfTitle', () => {
  it('extracts brand after Monitor', () => {
    expect(brandOfTitle('Monitor Dell P2219H 22"')).toBe('Dell')
    expect(brandOfTitle('Monitor Samsung 733NW 17" (sem base)')).toBe('Samsung')
  })
  it('maps bare Monitor to Genérico', () => {
    expect(brandOfTitle('Monitor')).toBe('Genérico')
  })
})

describe('baseOfTitle', () => {
  it('detects com/sem base', () => {
    expect(baseOfTitle('X (sem base)')).toBe('without')
    expect(baseOfTitle('X (com base)')).toBe('with')
    expect(baseOfTitle('X')).toBe('unset')
  })
})

describe('salePriceCents / marginCents', () => {
  it('reads integer cents', () => {
    expect(salePriceCents({ salePriceRaw: 39900 })).toBe(39900)
    expect(salePriceCents({ salePriceCents: 14900 })).toBe(14900)
    expect(marginCents({ salePriceRaw: 39900, unitCostRaw: 3121 })).toBe(39900 - 3121)
  })

  it('parses Brazilian input strings used by the stock form', () => {
    // Stock handler historically sent formatCentsInput → "399,00"
    expect(salePriceCents({ salePriceRaw: '399,00' })).toBe(39900)
    expect(salePriceCents({ salePriceRaw: '1.399,50' })).toBe(139950)
    expect(salePriceCents({ salePriceRaw: '99,00' })).toBe(9900)
  })

  it('sorts prices numerically not as strings', () => {
    const items = [
      { id: 1, salePriceRaw: '99,00', title: 'A' },
      { id: 2, salePriceRaw: '399,00', title: 'B' },
      { id: 3, salePriceRaw: '149,00', title: 'C' },
      { id: 4, salePriceRaw: '', title: 'D' },
    ]
    const asc = filterAndSortStockItems(items, { sortKey: 'salePrice', sortDir: 'asc' })
    expect(asc.map((x) => x.id)).toEqual([1, 3, 2, 4]) // nulls last
    const desc = filterAndSortStockItems(items, { sortKey: 'salePrice', sortDir: 'desc' })
    expect(desc.map((x) => x.id)).toEqual([2, 3, 1, 4])
  })
})

describe('filterAndSortStockItems', () => {
  const items = [
    {
      id: 2,
      title: 'Monitor Dell A',
      isAccessory: false,
      unitCostRaw: 100,
      salePriceRaw: 200,
      salePriceHint: 'R$ 2,00',
      lotId: 1,
      qty: 3,
    },
    {
      id: 1,
      title: 'Cabo VGA',
      isAccessory: true,
      unitCostRaw: 50,
      salePriceRaw: null,
      salePriceHint: '',
      lotId: 2,
      qty: 1,
    },
  ]

  it('sorts by quantity', () => {
    const desc = filterAndSortStockItems(items, { sortKey: 'qty', sortDir: 'desc' })
    expect(desc.map((x) => x.id)).toEqual([2, 1])
    const asc = filterAndSortStockItems(items, { sortKey: 'qty', sortDir: 'asc' })
    expect(asc.map((x) => x.id)).toEqual([1, 2])
  })

  it('filters accessories and sorts by id', () => {
    const out = filterAndSortStockItems(items, {
      filterType: 'accessory',
      sortKey: 'id',
      sortDir: 'asc',
    })
    expect(out).toHaveLength(1)
    expect(out[0].id).toBe(1)
  })

  it('filters by brand and query', () => {
    const out = filterAndSortStockItems(items, { filterBrand: 'Dell', query: 'dell' })
    expect(out).toHaveLength(1)
    expect(out[0].title).toContain('Dell')
  })
})
