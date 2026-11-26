import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const page = readFileSync(resolve(dirname(fileURLToPath(import.meta.url)), 'Show.svelte'), 'utf8')

describe('Lots Show items table', () => {
  it('lista itens em tabela com busca, filtro de status e sort', () => {
    expect(page).toMatch(/filterAndSortLotItems/)
    expect(page).toMatch(/SearchableSelect/)
    expect(page).toMatch(/filterStatus/)
    expect(page).toMatch(/toggleSort/)
    expect(page).toMatch(/<table/)
    expect(page).toMatch(/Buscar item/)
    expect(page).toMatch(/Limpar filtros/)
    expect(page).not.toMatch(/Salvar item/)
    expect(page).not.toMatch(/SALVAR ITEM/)
  })

  it('mantém o POST de salvar item por unidade', () => {
    expect(page).toMatch(/\/lots\/\$\{lot\.id\}\/items\/\$\{/)
    expect(page).toMatch(/unit_cost/)
    expect(page).toMatch(/sale_price_hint/)
  })
})
