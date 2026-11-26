import { readdirSync, readFileSync } from 'node:fs'
import { dirname, join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const here = dirname(fileURLToPath(import.meta.url))
const srcRoot = resolve(here, '..')
const repoRoot = resolve(srcRoot, '../..')
const css = readFileSync(resolve(repoRoot, 'input.css'), 'utf8')
const tailwind = readFileSync(resolve(repoRoot, 'tailwind.config.js'), 'utf8')
const dashboard = readFileSync(resolve(srcRoot, 'pages/Dashboard.svelte'), 'utf8')
const lotsIndex = readFileSync(resolve(srcRoot, 'pages/Lots/Index.svelte'), 'utf8')

function walkSvelte(dir, acc = []) {
  for (const name of readdirSync(dir, { withFileTypes: true })) {
    const path = join(dir, name.name)
    if (name.isDirectory()) walkSvelte(path, acc)
    else if (name.name.endsWith('.svelte')) acc.push(path)
  }
  return acc
}

describe('Admin chrome', () => {
  it('ahq-* fica quieto: sentence case, sem scale, marca sem círculo', () => {
    const label = css.match(/\.ahq-label \{[^}]*\}/)?.[0] || ''
    const primary = css.match(/\.ahq-btn-primary \{[^}]*\}/)?.[0] || ''
    const dark = css.match(/\.ahq-btn-dark \{[^}]*\}/)?.[0] || ''
    const ghost = css.match(/\.ahq-btn-ghost \{[^}]*\}/)?.[0] || ''
    const badge = css.match(/\.ahq-badge \{[^}]*\}/)?.[0] || ''
    const mark = css.match(/\.ahq-brand-mark \{[^}]*\}/)?.[0] || ''
    const themeBtn = css.match(/\.ahq-theme-btn \{[^}]*\}/)?.[0] || ''

    for (const block of [label, primary, dark, ghost, badge]) {
      expect(block).not.toMatch(/uppercase/)
      expect(block).not.toMatch(/tracking-wide/)
      expect(block).not.toMatch(/tracking-wider/)
    }
    for (const block of [primary, dark, ghost, themeBtn]) {
      expect(block).not.toMatch(/scale/)
    }
    expect(primary).not.toMatch(/rounded-full/)
    expect(mark).not.toMatch(/rounded-full/)
    expect(css).not.toMatch(/pulse-live/)
  })

  it('Inter do admin não passa de peso 600', () => {
    expect(tailwind).not.toMatch(/fontWeight:\s*["']?(700|800|900)/)
  })

  it('páginas Svelte do admin não caem em AuctionHQ', () => {
    const pages = walkSvelte(resolve(srcRoot, 'pages')).filter((p) => !p.includes('/Shop/'))
    const shell = resolve(srcRoot, 'components/AppShell.svelte')
    for (const file of [...pages, shell]) {
      const src = readFileSync(file, 'utf8')
      expect(src, file).not.toMatch(/AuctionHQ/)
    }
  })

  it('dashboard e lotes falam da bancada, não de ERP de leilão', () => {
    expect(dashboard).toMatch(/Caixa, lotes e o que está na bancada/)
    expect(dashboard).not.toMatch(/negócio de leilão/)
    expect(dashboard).toMatch(/Registrar primeiro lote/)
    expect(lotsIndex).toMatch(/Compras e rateio de custos/)
    expect(lotsIndex).not.toMatch(/Compras de leilão/)
  })
})
