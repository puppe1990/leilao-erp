import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const dir = dirname(fileURLToPath(import.meta.url))
const page = readFileSync(resolve(dir, 'Index.svelte'), 'utf8')
const show = readFileSync(resolve(dir, 'Show.svelte'), 'utf8')
const lot = readFileSync(resolve(dir, 'Lot.svelte'), 'utf8')
const css = readFileSync(resolve(dir, '../../../static/css/shop.css'), 'utf8')

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
    expect(page).toMatch(/class="shop-wordmark"/)
    expect(page).not.toMatch(/font-weight:900/)
  })

  it('filtra carregadores e cabos em Acessórios, fora de Com base', () => {
    expect(page).toMatch(/id: 'acessorio'/)
    expect(page).toMatch(/label: 'Acessórios'/)
    expect(page).not.toMatch(/ACESSÓRIOS/)
    expect(page).toMatch(/p\.category === 'acessorio'/)
  })

  it('og do catálogo fala de eletro e tech, com fotos reais', () => {
    expect(page).toMatch(/Eletro e tech/)
    expect(page).toMatch(/Fotos reais/)
    expect(page).not.toMatch(/Eletro & Tech/)
    expect(page).not.toMatch(/Monitores & Tech/)
    expect(page).not.toMatch(/Monitores e tech/)
    expect(page).not.toMatch(/Monitores usados testados/)
  })

  it('no mobile a busca ocupa a linha inteira', () => {
    expect(page).toMatch(/class="shop-header-actions"/)
    expect(css).toMatch(/\.shop-search-mobile \{[^}]*flex:\s*1 0 100%/s)
    expect(css).toMatch(/\.shop-search-mobile \{[^}]*max-width:\s*none/)
    expect(css).toMatch(/\.shop-header-actions \{[^}]*flex-shrink:\s*0/)
    expect(css).toMatch(/\.shop-sub \{[^}]*display:\s*none/)
  })

  it('chrome quieto: sentence case, sem pílula Catálogo nem raio no banner', () => {
    expect(page).toMatch(/label: 'Todos'/)
    expect(page).toMatch(/placeholder="Buscar"/)
    expect(page).not.toMatch(/class="shop-pill"/)
    expect(page).not.toMatch(/>bolt</)
  })

  it('tema da loja sem peso 900, glow verde, scale no hover ou gradiente no chat', () => {
    expect(css).not.toMatch(/font-weight:\s*900/)
    expect(css).not.toMatch(/box-shadow:[^;{]*211,\s*102/)
    expect(css).not.toMatch(/transform:\s*scale\(/)
    expect(css).not.toMatch(/linear-gradient\(90deg,\s*var\(--shop-float-head/)
  })

  it('banner, chips, WhatsApp e CTA não gritam em caixa alta', () => {
    expect(css).not.toMatch(/\.shop-banner \{[^}]*text-transform:\s*uppercase/s)
    expect(css).not.toMatch(/\.shop-btn-wa \{[^}]*text-transform:\s*uppercase/s)
    expect(css).not.toMatch(/\.shop-cat \{[^}]*text-transform:\s*uppercase/s)
    expect(css).not.toMatch(/\.shop-cta \{[^}]*text-transform:\s*uppercase/s)
  })

  it('pinta a scrollbar no verde do sistema', () => {
    expect(css).toMatch(/scrollbar-color:\s*var\(--shop-green\)/)
    expect(css).toMatch(/::-webkit-scrollbar-thumb/)
    expect(css).toMatch(/background:\s*var\(--shop-green\)/)
  })

  it('pinta Pronta entrega de âmbar, fora do verde do PIX', () => {
    expect(page).toMatch(/shopBadgeClass\(p\.badge\)/)
    expect(show).toMatch(/shopBadgeClass\(product\.badge\)/)
    const ready = css.match(/\.shop-badge-ready \{[^}]*\}/)
    expect(ready?.[0]).toMatch(/background:\s*#e4c15a/)
    expect(ready?.[0]).not.toMatch(/25d366/)
  })

  it('usa a marca pixel 3×3 no logo da loja', () => {
    expect(page).toMatch(/import PixEletroMark from '@\/components\/PixEletroMark\.svelte'/)
    expect(page).toMatch(/<div class="shop-logo-mark">\s*<PixEletroMark /)
    expect(show).toMatch(/<PixEletroMark /)
    expect(lot).toMatch(/<PixEletroMark /)
  })

  it('ordena menor e maior preço pelo helper, sem prender acessório na frente', () => {
    expect(page).toMatch(/import \{ sortShopCatalog \} from '@\/lib\/shopCatalog\.js'/)
    expect(page).toMatch(/sortShopCatalog\(/)
    expect(page).not.toMatch(/if \(accA !== accB\) return accA \? -1 : 1/)
  })
})



