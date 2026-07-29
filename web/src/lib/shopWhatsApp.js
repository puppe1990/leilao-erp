/**
 * Build WhatsApp click-to-chat URL for a product order.
 * @param {string} phoneDigits E.164-ish digits (with country code preferred)
 * @param {{ name: string, price?: string }} product
 * @returns {string}
 */
export function buildWhatsAppOrderURL(phoneDigits, product) {
  const digits = String(phoneDigits || '').replace(/\D/g, '')
  if (!digits) return ''
  const name = product?.name || 'produto'
  const price = product?.price || ''
  const lines = [
    'Olá! Vi o anúncio e quero comprar:',
    `• ${name}`,
  ]
  if (price) lines.push(`• Preço: ${price}`)
  lines.push('', 'Pode me passar a disponibilidade e formas de pagamento?')
  const text = encodeURIComponent(lines.join('\n'))
  return `https://wa.me/${digits}?text=${text}`
}
