import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const shell = readFileSync(resolve(dirname(fileURLToPath(import.meta.url)), 'AppShell.svelte'), 'utf8')

describe('AppShell', () => {
  it('abre o atalho Catálogo em nova aba', () => {
    const match = shell.match(/<a\s+href="\/"[\s\S]*?<\/a>/)
    expect(match).toBeTruthy()
    expect(match[0]).toMatch(/Catálogo/)
    expect(match[0]).toMatch(/target="_blank"/)
    expect(match[0]).toMatch(/rel="noopener noreferrer"/)
    expect(match[0]).not.toMatch(/use:inertia/)
  })
})
