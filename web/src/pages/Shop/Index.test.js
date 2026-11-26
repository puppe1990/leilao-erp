import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const dir = dirname(fileURLToPath(import.meta.url))
const page = readFileSync(resolve(dir, 'Index.svelte'), 'utf8')
const show = readFileSync(resolve(dir, 'Show.svelte'), 'utf8')
const lot = readFileSync(resolve(dir, 'Lot.svelte'), 'utf8')

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

  it('marca o catálogo como Pix Eletro', () => {
    for (const src of [page, show, lot]) {
      expect(src).toMatch(/companyName = 'Pix Eletro'/)
      expect(src).toMatch(/Pix Eletro<span>\.<\/span>/)
      expect(src).not.toMatch(/Puppe/)
    }
    expect(page).toMatch(/PIX ELETRO<span/)
  })

  it('filtra carregadores e cabos em ACESSÓRIOS, fora de COM BASE', () => {
    expect(page).toMatch(/id: 'acessorio'/)
    expect(page).toMatch(/ACESSÓRIOS/)
    expect(page).toMatch(/p\.category === 'acessorio'/)
  })

  it('og do catálogo fala de monitores e tech, com fotos reais', () => {
    expect(page).toMatch(/Monitores e tech/)
    expect(page).toMatch(/Fotos reais/)
    expect(page).not.toMatch(/Monitores usados testados/)
  })
})

