import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const page = readFileSync(resolve(dirname(fileURLToPath(import.meta.url)), 'Index.svelte'), 'utf8')

describe('Config catálogo', () => {
  it('tem a opção de mostrar quantidade no catálogo', () => {
    expect(page).toMatch(/export let shopShowQty = true/)
    expect(page).toMatch(/\/config\/shop-qty/)
    expect(page).toMatch(/shop_show_qty/)
    expect(page).toMatch(/Mostrar quantidade no catálogo/)
  })
})
