import { writable } from 'svelte/store'

/**
 * @typedef {object} ConfirmOptions
 * @property {string} [title]
 * @property {string} [message]
 * @property {string} [detail]
 * @property {string} [confirmLabel]
 * @property {string} [cancelLabel]
 * @property {'danger'|'warning'|'primary'} [tone]
 * @property {string} [icon] Material Symbols name
 */

/** @type {import('svelte/store').Writable<(ConfirmOptions & { resolve: (v: boolean) => void }) | null>} */
export const confirmState = writable(null)

/**
 * Opens the shared confirm modal. Resolves true if the user confirms.
 * @param {ConfirmOptions} opts
 * @returns {Promise<boolean>}
 */
export function askConfirm(opts = {}) {
  return new Promise((resolve) => {
    confirmState.set({
      title: opts.title || 'Tem certeza?',
      message: opts.message || 'Essa ação não pode ser desfeita.',
      detail: opts.detail || '',
      confirmLabel: opts.confirmLabel || 'Confirmar',
      cancelLabel: opts.cancelLabel || 'Cancelar',
      tone: opts.tone || 'danger',
      icon: opts.icon || defaultIcon(opts.tone || 'danger'),
      resolve,
    })
  })
}

/**
 * @param {boolean} result
 */
export function closeConfirm(result) {
  confirmState.update((s) => {
    if (s?.resolve) s.resolve(!!result)
    return null
  })
}

/** @param {'danger'|'warning'|'primary'} tone */
function defaultIcon(tone) {
  if (tone === 'warning') return 'warning'
  if (tone === 'primary') return 'help'
  return 'delete'
}
