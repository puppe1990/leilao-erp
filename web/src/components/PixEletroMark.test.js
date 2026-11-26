import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const src = readFileSync(resolve(dirname(fileURLToPath(import.meta.url)), 'PixEletroMark.svelte'), 'utf8')

describe('marca Pix Eletro', () => {
  it('é uma grade 3×3 com o pixel de baixo à direita no verde do PIX', () => {
    const cells = src.match(/class="cell"/g) || []
    expect(cells).toHaveLength(8)
    expect(src).toMatch(/class="cell-pix"/)
    expect(src).toMatch(/viewBox="0 0 64 64"/)
  })
})
