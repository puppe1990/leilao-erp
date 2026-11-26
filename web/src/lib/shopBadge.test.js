import { describe, expect, it } from 'vitest'
import { shopBadgeClass } from './shopBadge.js'

describe('classe do selo da loja', () => {
  it('pinta Pronta entrega numa cor própria', () => {
    expect(shopBadgeClass('PRONTA ENTREGA')).toBe('shop-badge-ready')
  })

  it('mantém Defeito no selo ruim e o resto no quieto', () => {
    expect(shopBadgeClass('DEFEITO')).toBe('shop-badge-bad')
    expect(shopBadgeClass('SEM BASE')).toBe('shop-badge-ok')
    expect(shopBadgeClass('Lote')).toBe('shop-badge-ok')
    expect(shopBadgeClass('')).toBe('shop-badge-ok')
  })
})
