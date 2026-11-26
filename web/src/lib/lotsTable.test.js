import { describe, expect, it } from 'vitest'
import {
  filterAndSortLots,
  loadLotSorts,
  LOT_SORTS_STORAGE_KEY,
  saveLotSorts,
  sanitizeLotSorts,
  stackLotSort,
} from './lotsTable.js'

const sample = [
  {
    id: 12,
    name: 'PMCELL Kit Turbo 26/11',
    purchasedAt: '2026-11-26',
    status: 'open',
    statusLabel: 'Aberto',
    totalCost: 'R$ 300,00',
    totalCostCents: 30000,
    itemCount: 43,
  },
  {
    id: 11,
    name: 'Pessoal — Galaxy S9+',
    purchasedAt: '2026-11-23',
    status: 'open',
    statusLabel: 'Aberto',
    totalCost: 'R$ 170,00',
    totalCostCents: 17000,
    itemCount: 1,
  },
  {
    id: 1,
    name: 'Monitores — leilão Jul/2026',
    purchasedAt: '2026-07-20',
    status: 'partial',
    statusLabel: 'Parcial',
    totalCost: 'R$ 1.500,00',
    totalCostCents: 150000,
    itemCount: 22,
  },
  {
    id: 10,
    name: 'Pessoal — não leilão',
    purchasedAt: '2026-10-01',
    status: 'sold',
    statusLabel: 'Vendido',
    totalCost: 'R$ 0,00',
    totalCostCents: 0,
    itemCount: 1,
  },
]

describe('filterAndSortLots', () => {
  it('filters by status', () => {
    expect(filterAndSortLots(sample, { filterStatus: 'open' }).map((x) => x.id)).toEqual([12, 11])
    expect(filterAndSortLots(sample, { filterStatus: 'partial' }).map((x) => x.id)).toEqual([1])
    expect(filterAndSortLots(sample, { filterStatus: 'sold' }).map((x) => x.id)).toEqual([10])
  })

  it('searches name, date and status label', () => {
    expect(filterAndSortLots(sample, { query: 'pmcell' }).map((x) => x.id)).toEqual([12])
    expect(filterAndSortLots(sample, { query: '2026-07' }).map((x) => x.id)).toEqual([1])
    expect(filterAndSortLots(sample, { query: 'parcial' }).map((x) => x.id)).toEqual([1])
  })

  it('sorts by cost and item count', () => {
    const byCost = filterAndSortLots(sample, { sortKey: 'totalCost', sortDir: 'desc' })
    expect(byCost.map((x) => x.id)).toEqual([1, 12, 11, 10])

    const byQty = filterAndSortLots(sample, { sortKey: 'itemCount', sortDir: 'desc' })
    expect(byQty.map((x) => x.id)).toEqual([12, 1, 11, 10])
  })

  it('defaults to purchasedAt desc (newest first)', () => {
    expect(filterAndSortLots(sample).map((x) => x.id)).toEqual([12, 11, 10, 1])
  })

  it('replaces default date so cost sort actually reorders', () => {
    let sorts = [{ key: 'purchasedAt', dir: 'desc' }]
    sorts = stackLotSort(sorts, 'totalCost')
    expect(sorts).toEqual([{ key: 'totalCost', dir: 'desc' }])

    const out = filterAndSortLots(sample, { sorts })
    expect(out.map((x) => x.id)).toEqual([1, 12, 11, 10])
  })

  it('stacks multi-column sorts: last click is primary', () => {
    let sorts = [{ key: 'purchasedAt', dir: 'desc' }]
    sorts = stackLotSort(sorts, 'status')
    sorts = stackLotSort(sorts, 'itemCount')
    expect(sorts).toEqual([
      { key: 'itemCount', dir: 'desc' },
      { key: 'status', dir: 'asc' },
    ])
  })
})

describe('lot sorts localStorage', () => {
  it('sanitizes invalid keys and dirs', () => {
    expect(
      sanitizeLotSorts([
        { key: 'totalCost', dir: 'desc' },
        { key: 'hack', dir: 'asc' },
        { key: 'name', dir: 'nope' },
        { key: 'totalCost', dir: 'asc' },
      ]),
    ).toEqual([
      { key: 'totalCost', dir: 'desc' },
      { key: 'name', dir: 'asc' },
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
      { key: 'status', dir: 'asc' },
      { key: 'purchasedAt', dir: 'desc' },
    ]
    saveLotSorts(sorts, storage)
    expect(mem[LOT_SORTS_STORAGE_KEY]).toBeTruthy()
    expect(loadLotSorts(storage)).toEqual(sorts)
  })
})
