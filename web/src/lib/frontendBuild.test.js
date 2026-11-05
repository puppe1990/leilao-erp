import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '../../..')
const pkg = JSON.parse(readFileSync(resolve(root, 'package.json'), 'utf8'))

describe('npm run build (Cleat golang frontend step)', () => {
  it('generates the gitignored Tailwind stylesheet before Vite', () => {
    const build = pkg.scripts.build || ''
    expect(build).toMatch(/tailwindcss/)
    expect(build).toMatch(/web\/static\/css\/styles\.css/)
    expect(build).toMatch(/vite build/)
    const tailwindAt = build.indexOf('tailwindcss')
    const viteAt = build.indexOf('vite build')
    expect(tailwindAt).toBeGreaterThanOrEqual(0)
    expect(viteAt).toBeGreaterThan(tailwindAt)
  })
})
