import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const page = readFileSync(resolve(dirname(fileURLToPath(import.meta.url)), 'Index.svelte'), 'utf8')

describe('Catálogo da loja', () => {
  it('não anuncia o lote fechado de R$ 1.500', () => {
    expect(page).not.toMatch(/lote fechado por R\$ 1\.500/i)
    expect(page).not.toMatch(/lote inteiro por R\$ 1\.500/i)
  })

  it('não renderiza o hero de monitores testados', () => {
    expect(page).not.toMatch(/class="shop-hero"/)
    expect(page).not.toMatch(/Monitores <em>testados<\/em>/)
    expect(page).not.toMatch(/Ver catálogo/)
  })
})
