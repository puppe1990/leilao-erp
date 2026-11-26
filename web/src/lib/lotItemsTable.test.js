import { describe, expect, it } from 'vitest'
import {
  filterAndSortLotItems,
  loadLotItemSorts,
  LOT_ITEM_SORTS_STORAGE_KEY,
  saveLotItemSorts,
  sanitizeLotItemSorts,
  stackLotItemSort,
} from './lotItemsTable.js'

const sample = [
  {
    id: 97,
    title: 'Cabo turbo USB para Tipo C 66W',
    sku: 'CB-17-TC',
    unitCost: 'R$ 4,29',
    unitCostRaw: '4,29',
    salePriceHint: 'R$ 12,00',
    salePriceRaw: '12,00',
    marginHint: 'R$ 7,71',
    status: 'in_stock',
    statusLabel: 'Em estoque',
    canEdit: true,
  },
  {
    id: 98,
    title: 'Cabo turbo USB para Tipo C 66W',
    sku: 'CB-17-TC',
    unitCost: 'R$ 4,29',
    unitCostRaw: '4,29',
    salePriceHint: 'R$ 12,00',
    salePriceRaw: '12,00',
    marginHint: 'R$ 7,71',
    status: 'in_stock',
    statusLabel: 'Em estoque',
    canEdit: true,
  },
  {
    id: 110,
    title: 'Carregador Kaidi KD-27',
    sku: 'KD-27',
    unitCost: 'R$ 5,15',
    unitCostRaw: '5,15',
    salePriceHint: 'R$ 15,00',
    salePriceRaw: '15,00',
    marginHint: 'R$ 9,85',
    status: 'in_stock',
    statusLabel: 'Em estoque',
    canEdit: true,
  },
  {
    id: 120,
    title: 'Fonte HC-61',
    sku: 'HC-61',
    unitCost: 'R$ 15,45',
    unitCostRaw: '15,45',
    salePriceHint: 'R$ 35,00',
    salePriceRaw: '35,00',
    marginHint: 'R$ 19,55',
    status: 'sold',
    statusLabel: 'Vendido',
    canEdit: false,
  },
  {
    id: 121,
    title: 'Cabo CB-18 C-C',
    sku: 'CB-18-CC',
    unitCost: 'R$ 4,72',
    unitCostRaw: '4,72',
    salePriceHint: '',
    salePriceRaw: '',
    marginHint: '',
    status: 'reserved',
    statusLabel: 'Reservado',
    canEdit: true,
  },
]

describe('filterAndSortLotItems', () => {
  it('filters by status', () => {
    expect(filterAndSortLotItems(sample, { filterStatus: 'in_stock' }).map((x) => x.id)).toEqual([
      97, 98, 110,
    ])
    expect(filterAndSortLotItems(sample, { filterStatus: 'sold' }).map((x) => x.id)).toEqual([120])
    expect(filterAndSortLotItems(sample, { filterStatus: 'reserved' }).map((x) => x.id)).toEqual([
      121,
    ])
  })

  it('searches title, sku, id and status label', () => {
    expect(filterAndSortLotItems(sample, { query: 'kaidi' }).map((x) => x.id)).toEqual([110])
    expect(filterAndSortLotItems(sample, { query: 'cb-17' }).map((x) => x.id)).toEqual([97, 98])
    expect(filterAndSortLotItems(sample, { query: '120' }).map((x) => x.id)).toEqual([120])
    expect(filterAndSortLotItems(sample, { query: 'reservado' }).map((x) => x.id)).toEqual([121])
  })

  it('sorts by unit cost, sale price and margin', () => {
    const byCost = filterAndSortLotItems(sample, { sortKey: 'unitCost', sortDir: 'desc' })
    expect(byCost.map((x) => x.id)).toEqual([120, 110, 121, 97, 98])

    const bySale = filterAndSortLotItems(sample, { sortKey: 'salePrice', sortDir: 'desc' })
    expect(bySale.map((x) => x.id)[0]).toBe(120)

    const byMargin = filterAndSortLotItems(sample, { sortKey: 'margin', sortDir: 'desc' })
    expect(byMargin.map((x) => x.id)[0]).toBe(120)
  })

  it('defaults to id asc so kit units stay in sequence', () => {
    expect(filterAndSortLotItems(sample).map((x) => x.id)).toEqual([97, 98, 110, 120, 121])
  })

  it('replaces default id sort so sku sort actually reorders', () => {
    let sorts = [{ key: 'id', dir: 'asc' }]
    sorts = stackLotItemSort(sorts, 'sku')
    expect(sorts).toEqual([{ key: 'sku', dir: 'asc' }])

    const out = filterAndSortLotItems(sample, { sorts })
    expect(out.map((x) => x.sku)).toEqual(['CB-17-TC', 'CB-17-TC', 'CB-18-CC', 'HC-61', 'KD-27'])
  })

  it('stacks multi-column sorts: last click is primary', () => {
    let sorts = [{ key: 'id', dir: 'asc' }]
    sorts = stackLotItemSort(sorts, 'status')
    sorts = stackLotItemSort(sorts, 'unitCost')
    expect(sorts).toEqual([
      { key: 'unitCost', dir: 'desc' },
      { key: 'status', dir: 'asc' },
    ])
  })
})

describe('lot item sorts localStorage', () => {
  it('sanitizes invalid keys and dirs', () => {
    expect(
      sanitizeLotItemSorts([
        { key: 'unitCost', dir: 'desc' },
        { key: 'hack', dir: 'asc' },
        { key: 'title', dir: 'nope' },
        { key: 'unitCost', dir: 'asc' },
      ]),
    ).toEqual([
      { key: 'unitCost', dir: 'desc' },
      { key: 'title', dir: 'asc' },
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
      { key: 'sku', dir: 'asc' },
      { key: 'id', dir: 'desc' },
    ]
    saveLotItemSorts(sorts, storage)
    expect(mem[LOT_ITEM_SORTS_STORAGE_KEY]).toBeTruthy()
    expect(loadLotItemSorts(storage)).toEqual(sorts)
  })
})
