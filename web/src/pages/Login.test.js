import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const dir = dirname(fileURLToPath(import.meta.url))
const login = readFileSync(resolve(dir, 'Login.svelte'), 'utf8')
const forgot = readFileSync(resolve(dir, 'ForgotPassword.svelte'), 'utf8')
const reset = readFileSync(resolve(dir, 'ResetPassword.svelte'), 'utf8')
const home = readFileSync(resolve(dir, 'Home.svelte'), 'utf8')

describe('Auth chrome', () => {
  it('login, esqueci e reset usam a marca C e PIX ELETRO.', () => {
    for (const src of [login, forgot, reset]) {
      expect(src).toMatch(/import PixEletroMark from '@\/components\/PixEletroMark\.svelte'/)
      expect(src).toMatch(/<PixEletroMark /)
      expect(src).toMatch(/Pix Eletro<span class="text-secondary">\.<\/span>/)
      expect(src).not.toMatch(/MonitorIcon/)
      expect(src).not.toMatch(/font-black/)
      expect(src).not.toMatch(/ERP · leilões/)
      expect(src).not.toMatch(/Acesso restrito/)
    }
  })

  it('login fala de estoque e vendas, não de ERP de leilões', () => {
    expect(login).toMatch(/estoque e as vendas/)
    expect(login).not.toMatch(/gestão de leilões/)
  })

  it('home de ponte não anuncia AuctionHQ', () => {
    expect(home).toMatch(/Pix Eletro/)
    expect(home).toMatch(/<PixEletroMark /)
    expect(home).not.toMatch(/AuctionHQ/)
    expect(home).not.toMatch(/>AQ</)
    expect(home).not.toMatch(/gestão de leilões/)
  })
})
