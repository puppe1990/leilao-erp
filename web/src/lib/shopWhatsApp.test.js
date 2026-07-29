import { describe, expect, it } from 'vitest'
import { buildWhatsAppOrderURL } from './shopWhatsApp.js'

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
