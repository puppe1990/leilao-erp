import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const page = readFileSync(resolve(dirname(fileURLToPath(import.meta.url)), 'Index.svelte'), 'utf8')

describe('Finds Ações', () => {
  it('mostra o botão Descartar na coluna de ações', () => {
    expect(page).toMatch(/Ações/)
    expect(page).toMatch(/on:click=\{\(\) => destroy\(row\)\}/)
    expect(page).toMatch(/>\s*Descartar\s*</)
    expect(page).not.toMatch(/>\s*Excluir\s*</)
    expect(page).toMatch(/confirmLabel: 'Descartar'/)
  })
})
