import { describe, expect, it } from 'vitest'
import {
  buildCartOrderMessage,
  buildWhatsAppCartURL,
  buildWhatsAppOrderURL,
} from './shopWhatsApp.js'

describe('buildWhatsAppOrderURL', () => {
  it('returns empty without phone', () => {
    expect(buildWhatsAppOrderURL('', { name: 'X' })).toBe('')
  })

  it('builds wa.me URL with product and price', () => {
    const url = buildWhatsAppOrderURL('5511999998888', {
      name: 'Monitor Dell P2016t',
      price: 'R$ 279,00',
    })
    expect(url.startsWith('https://wa.me/5511999998888?text=')).toBe(true)
    const text = decodeURIComponent(url.split('text=')[1])
    expect(text).toContain('Monitor Dell P2016t')
    expect(text).toContain('279')
    expect(text).toContain('quero comprar')
  })

  it('strips non-digits from phone', () => {
    const url = buildWhatsAppOrderURL('(11) 99999-8888', { name: 'Y' })
    expect(url.startsWith('https://wa.me/11999998888?text=')).toBe(true)
  })
})

describe('buildCartOrderMessage', () => {
  it('returns empty for empty cart', () => {
    expect(buildCartOrderMessage([])).toBe('')
  })

  it('lists quantities and PIX total', () => {
    const msg = buildCartOrderMessage(
      [
        { name: 'Dell P2219H', qty: 2, priceCents: 30000 },
        { name: 'LG 19', qty: 1, priceCents: 20000 },
      ],
      { companyName: 'Puppe Eletrônicos' },
    )
    expect(msg).toContain('Olá Puppe Eletrônicos!')
    expect(msg).toContain('2x Dell P2219H')
    expect(msg).toContain('1x LG 19')
    expect(msg).toContain('Total PIX')
    expect(msg).toContain('Prefiro pagar no PIX')
    // 2*27000 + 18000 = 72000 → R$ 720,00
    expect(msg).toMatch(/720/)
  })
})

describe('buildWhatsAppCartURL', () => {
  it('encodes cart message', () => {
    const url = buildWhatsAppCartURL('5511999990000', [
      { name: 'Monitor X', qty: 1, priceCents: 10000 },
    ])
    expect(url.startsWith('https://wa.me/5511999990000?text=')).toBe(true)
    const text = decodeURIComponent(url.split('text=')[1])
    expect(text).toContain('1x Monitor X')
    expect(text).toContain('pedido')
  })
})
