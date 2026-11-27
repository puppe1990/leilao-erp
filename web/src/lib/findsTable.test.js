import { describe, expect, it } from 'vitest'
import {
  brlCents,
  defaultFindSorts,
  loadFindSorts,
  sanitizeFindSorts,
  saveFindSorts,
  sortFinds,
  stackFindSort,
} from './findsTable.js'

const rows = [
  { id: 1, name: 'iPhone XR', source: 'facebook', buy: 'R$ 350,00', sell: 'R$ 700,00', margin: 'R$ 350,00' },
  { id: 2, name: 'Kwara iPhone 13 Pro', source: 'kwara', buy: 'R$ 1.204,25', sell: 'R$ 2.400,00', margin: 'R$ 1.195,75' },
  { id: 3, name: 'Bike aro 26', source: 'olx', buy: 'R$ 130,00', sell: 'R$ 300,00', margin: 'R$ 170,00' },
]

describe('sort da tabela de achados', () => {
  it('ordena por margem desc por padrão', () => {
    expect(sortFinds(rows, defaultFindSorts()).map((r) => r.id)).toEqual([2, 1, 3])
  })

  it('ordena por comprar ascendente e nome em pt-BR', () => {
    expect(sortFinds(rows, [{ key: 'buy', dir: 'asc' }]).map((r) => r.id)).toEqual([3, 1, 2])
    expect(sortFinds(rows, [{ key: 'name', dir: 'asc' }]).map((r) => r.name)).toEqual([
      'Bike aro 26',
      'iPhone XR',
      'Kwara iPhone 13 Pro',
    ])
  })

  it('empilha colunas: último clique vira principal', () => {
    let sorts = stackFindSort(defaultFindSorts(), 'source')
    sorts = stackFindSort(sorts, 'buy')
    expect(sorts).toEqual([
      { key: 'buy', dir: 'desc' },
      { key: 'source', dir: 'asc' },
    ])
    expect(sortFinds(rows, sorts).map((r) => r.id)).toEqual([2, 1, 3])
  })

  it('inverte a direção ao clicar de novo na mesma coluna', () => {
    const sorts = stackFindSort(stackFindSort([], 'margin'), 'margin')
    expect(sorts).toEqual([{ key: 'margin', dir: 'asc' }])
  })

  it('lê centavos de R$, de raw e de número', () => {
    expect(brlCents('R$ 1.204,25')).toBe(120425)
    expect(brlCents('1204,25')).toBe(120425)
    expect(brlCents('-R$ 50,00')).toBe(-5000)
    expect(brlCents(15000)).toBe(15000)
    expect(brlCents('')).toBe(0)
  })

  it('persiste e saneia a ordem no storage', () => {
    const mem = new Map()
    const storage = {
      getItem: (k) => (mem.has(k) ? mem.get(k) : null),
      setItem: (k, v) => mem.set(k, v),
    }
    saveFindSorts([{ key: 'margin', dir: 'desc' }, { key: 'hack', dir: 'asc' }], storage)
    expect(loadFindSorts(storage)).toEqual([{ key: 'margin', dir: 'desc' }])
    expect(sanitizeFindSorts('nope')).toEqual(defaultFindSorts())
    expect(sanitizeFindSorts([{ key: 'margin', dir: 'desc' }, { key: 'margin', dir: 'asc' }])).toEqual([
      { key: 'margin', dir: 'desc' },
    ])
  })
})
