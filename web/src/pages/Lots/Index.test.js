import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const page = readFileSync(resolve(dirname(fileURLToPath(import.meta.url)), 'Index.svelte'), 'utf8')

describe('Lots Index table', () => {
  it('usa tabela com busca, filtro de status e sort empilhado', () => {
    expect(page).toMatch(/filterAndSortLots/)
    expect(page).toMatch(/SearchableSelect/)
    expect(page).toMatch(/filterStatus/)
    expect(page).toMatch(/toggleSort/)
    expect(page).toMatch(/<table/)
    expect(page).toMatch(/Buscar lote/)
    expect(page).toMatch(/Limpar filtros/)
    expect(page).not.toMatch(/lot-card/)
  })
})
