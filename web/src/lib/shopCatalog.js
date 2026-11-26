function priceCents(product) {
  return Number(product?.priceCents) || 0
}

function compareFeatured(a, b) {
  if (Boolean(a.isLot) !== Boolean(b.isLot)) return a.isLot ? -1 : 1
  const accA = a.category === 'acessorio'
  const accB = b.category === 'acessorio'
  if (accA !== accB) return accA ? -1 : 1
  if ((a.badge === 'DEFEITO') !== (b.badge === 'DEFEITO')) {
    return a.badge === 'DEFEITO' ? 1 : -1
  }
  return priceCents(b) - priceCents(a)
}

/** Ordena o catálogo da loja. Preço ignora o agrupamento de acessório. */
export function sortShopCatalog(products, sortBy) {
  const list = Array.isArray(products) ? products.slice() : []
  return list.sort((a, b) => {
    if (sortBy === 'price-asc') return priceCents(a) - priceCents(b)
    if (sortBy === 'price-desc') return priceCents(b) - priceCents(a)
    return compareFeatured(a, b)
  })
}
