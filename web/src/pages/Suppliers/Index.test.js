import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const dir = dirname(fileURLToPath(import.meta.url))
const indexPage = readFileSync(resolve(dir, 'Index.svelte'), 'utf8')
const showPage = readFileSync(resolve(dir, 'Show.svelte'), 'utf8')
const shell = readFileSync(resolve(dir, '../../components/AppShell.svelte'), 'utf8')

describe('Fornecedores', () => {
  it('lista nome, whatsapp e última mensagem', () => {
    expect(indexPage).toMatch(/Fornecedores/)
    expect(indexPage).toMatch(/Novo fornecedor/)
    expect(indexPage).toMatch(/lastBody/)
    expect(indexPage).toMatch(/confirmLabel: 'Descartar'/)
    expect(indexPage).not.toMatch(/>\s*Excluir\s*</)
  })

  it('mostra a conversa do zap no detalhe', () => {
    expect(showPage).toMatch(/Conversa/)
    expect(showPage).toMatch(/fromMe/)
    expect(showPage).toMatch(/messages/)
  })

  it('entra na navegação do admin depois de Achados', () => {
    expect(shell).toMatch(/key: 'suppliers'/)
    expect(shell).toMatch(/label: 'Fornecedores'/)
    expect(shell).toMatch(/href: '\/suppliers'/)
  })
})
