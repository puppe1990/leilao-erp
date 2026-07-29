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
  const lines = ['Olá! Vi o anúncio e quero comprar:', `• ${name}`]
  if (price) lines.push(`• Preço: ${price}`)
  lines.push('', 'Pode me passar a disponibilidade e formas de pagamento?')
  const text = encodeURIComponent(lines.join('\n'))
  return `https://wa.me/${digits}?text=${text}`
}

/**
 * Format cents as BRL for order messages.
 * @param {number} cents
 */
function formatBRL(cents) {
  const n = Number(cents) || 0
  return (n / 100).toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })
}

/**
 * Build multi-item cart order message body.
 * @param {Array<{ name: string, qty: number, priceCents?: number, price?: string, pixPrice?: string }>} items
 * @param {{ companyName?: string }} [opts]
 * @returns {string}
 */
export function buildCartOrderMessage(items, opts = {}) {
  const list = Array.isArray(items) ? items.filter((it) => it && (Number(it.qty) || 0) > 0) : []
  if (list.length === 0) return ''

  const company = (opts.companyName || '').trim()
  const lines = []
  if (company) {
    lines.push(`Olá ${company}! Quero fechar o seguinte pedido:`)
  } else {
    lines.push('Olá! Quero fechar o seguinte pedido:')
  }
  lines.push('')

  let cardTotal = 0
  let pixTotal = 0
  let hasPriced = false

  for (const it of list) {
    const qty = Math.max(1, Math.floor(Number(it.qty) || 1))
    const unit = Number(it.priceCents) || 0
    let priceBit = ''
    if (unit > 0) {
      hasPriced = true
      const unitPix = Math.floor((unit * 90) / 100)
      cardTotal += unit * qty
      pixTotal += unitPix * qty
      priceBit = ` — ${formatBRL(unitPix * qty)} no PIX`
    } else if (it.pixPrice) {
      priceBit = ` — ${it.pixPrice} no PIX`
    } else if (it.price && it.price !== 'Consulte') {
      priceBit = ` — ${it.price}`
    }
    lines.push(`${qty}x ${it.name || 'Produto'}${priceBit}`)
  }

  lines.push('')
  if (hasPriced) {
    lines.push(`Total PIX (10% OFF): ${formatBRL(pixTotal)}`)
    lines.push(`Total cartão: ${formatBRL(cardTotal)}`)
    lines.push('')
    lines.push('Prefiro pagar no PIX.')
  }
  lines.push('Pode confirmar disponibilidade e combinar entrega/retirada?')
  return lines.join('\n')
}

/**
 * Build WhatsApp URL for a full cart order.
 * @param {string} phoneDigits
 * @param {Array<{ name: string, qty: number, priceCents?: number, price?: string, pixPrice?: string }>} items
 * @param {{ companyName?: string }} [opts]
 * @returns {string}
 */
export function buildWhatsAppCartURL(phoneDigits, items, opts = {}) {
  const digits = String(phoneDigits || '').replace(/\D/g, '')
  if (!digits) return ''
  const body = buildCartOrderMessage(items, opts)
  if (!body) return ''
  return `https://wa.me/${digits}?text=${encodeURIComponent(body)}`
}
