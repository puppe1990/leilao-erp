import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const page = readFileSync(resolve(dirname(fileURLToPath(import.meta.url)), 'Show.svelte'), 'utf8')

describe('Products Show media order', () => {
  it('tem setas para reordenar e posta a lista de ids', () => {
    expect(page).toMatch(/moveMediaItem/)
    expect(page).toMatch(/mediaIds/)
    expect(page).toMatch(/\/products\/\$\{product\.id\}\/media\/reorder/)
    expect(page).toMatch(/media_ids/)
    expect(page).toMatch(/arrow_upward/)
    expect(page).toMatch(/arrow_downward/)
    expect(page).toMatch(/firstPhotoIndex/)
    expect(page).toMatch(/>Capa</)
    expect(page).toMatch(/A primeira foto vira a capa da loja/)
  })

  it('arrasta o card para reordenar a mídia', () => {
    expect(page).toMatch(/moveMediaAt/)
    expect(page).toMatch(/draggable=\{true\}/)
    expect(page).toMatch(/on:drop/)
    expect(page).toMatch(/on:dragstart/)
    expect(page).toMatch(/drag_indicator/)
    expect(page).toMatch(/Arraste para reordenar/)
  })
})
