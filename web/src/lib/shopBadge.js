/** CSS class for the shop overlay badge. */
export function shopBadgeClass(badge) {
  if (badge === 'DEFEITO') return 'shop-badge-bad'
  if (badge === 'PRONTA ENTREGA') return 'shop-badge-ready'
  return 'shop-badge-ok'
}
