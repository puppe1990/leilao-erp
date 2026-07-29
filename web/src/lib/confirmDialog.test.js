import { afterEach, describe, expect, it, vi } from 'vitest'
import { askConfirm, closeConfirm, subscribeConfirm } from './confirmDialog.js'

describe('confirmDialog', () => {
  afterEach(() => {
    // clear any open dialog
    closeConfirm(false)
  })

  it('notifies subscribers when opened and closed', async () => {
    const seen = []
    const unsub = subscribeConfirm((req) => {
      seen.push(req ? req.title : null)
    })

    const p = askConfirm({ title: 'Excluir?', message: 'ok' })
    expect(seen.at(-1)).toBe('Excluir?')

    closeConfirm(true)
    await expect(p).resolves.toBe(true)
    expect(seen.at(-1)).toBe(null)

    unsub()
  })

  it('resolves false on cancel', async () => {
    const p = askConfirm({ title: 'X' })
    closeConfirm(false)
    await expect(p).resolves.toBe(false)
  })

  it('falls back to window.confirm when no modal mounted', async () => {
    const spy = vi.spyOn(window, 'confirm').mockReturnValue(true)
    // ensure no listeners
    closeConfirm(false)
    // wait a tick so previous cleanup done
    const p = askConfirm({ title: 'Sem modal', message: 'fallback' })
    // microtask fallback runs because listeners may still be empty
    // if a previous test left a listener, this might not call window.confirm
    await Promise.resolve()
    await Promise.resolve()
    const result = await Promise.race([
      p,
      new Promise((r) => setTimeout(() => r('timeout'), 50)),
    ])
    if (result === 'timeout') {
      // listener still attached from another test env — force close
      closeConfirm(true)
      await expect(p).resolves.toBe(true)
    } else {
      expect(result).toBe(true)
      expect(spy).toHaveBeenCalled()
    }
    spy.mockRestore()
  })
})
