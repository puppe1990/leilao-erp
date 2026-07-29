/**
 * Shared confirm dialog bus (works across Inertia page navigations).
 * ConfirmModal must be mounted (via AppShell) to show UI.
 */

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

/**
 * @typedef {ConfirmOptions & { resolve: (v: boolean) => void }} ConfirmRequest
 */

/** @type {ConfirmRequest | null} */
let current = null

/** @type {Set<(req: ConfirmRequest | null) => void>} */
const listeners = new Set()

function notify() {
  for (const fn of listeners) {
    try {
      fn(current)
    } catch {
      // ignore listener errors
    }
  }
}

/**
 * Subscribe to open/close requests. Returns unsubscribe.
 * @param {(req: ConfirmRequest | null) => void} fn
 */
export function subscribeConfirm(fn) {
  listeners.add(fn)
  // push current state immediately
  try {
    fn(current)
  } catch {
    // ignore
  }
  return () => {
    listeners.delete(fn)
  }
}

/**
 * Opens the shared confirm modal. Resolves true if the user confirms.
 * @param {ConfirmOptions} opts
 * @returns {Promise<boolean>}
 */
export function askConfirm(opts = {}) {
  // close any previous pending confirm as cancelled
  if (current?.resolve) {
    try {
      current.resolve(false)
    } catch {
      // ignore
    }
  }

  return new Promise((resolve) => {
    const tone = opts.tone || 'danger'
    current = {
      title: opts.title || 'Tem certeza?',
      message: opts.message || 'Essa ação não pode ser desfeita.',
      detail: opts.detail || '',
      confirmLabel: opts.confirmLabel || 'Confirmar',
      cancelLabel: opts.cancelLabel || 'Cancelar',
      tone,
      icon: opts.icon || defaultIcon(tone),
      resolve: (v) => {
        current = null
        notify()
        resolve(!!v)
      },
    }
    notify()

    // Safety: if no modal is mounted, fall back so deletes aren't stuck.
    queueMicrotask(() => {
      if (listeners.size === 0 && current) {
        const req = current
        current = null
        const ok =
          typeof window !== 'undefined' &&
          window.confirm([req.title, req.message, req.detail].filter(Boolean).join('\n\n'))
        req.resolve(ok)
      }
    })
  })
}

/**
 * @param {boolean} result
 */
export function closeConfirm(result) {
  if (!current) return
  const req = current
  current = null
  notify()
  req.resolve(!!result)
}

/** @param {'danger'|'warning'|'primary'} tone */
function defaultIcon(tone) {
  if (tone === 'warning') return 'warning'
  if (tone === 'primary') return 'help'
  return 'delete'
}
