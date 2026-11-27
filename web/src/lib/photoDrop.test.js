import { describe, expect, it } from 'vitest'
import { imageFilesFromDataTransfer } from './photoDrop.js'

describe('arrastar foto para o upload', () => {
  it('mantém só imagens aceitas pelo input', () => {
    const files = imageFilesFromDataTransfer({
      files: [
        { name: 'frente.jpg', type: 'image/jpeg' },
        { name: 'nota.txt', type: 'text/plain' },
        { name: 'selo.webp', type: 'IMAGE/WEBP' },
      ],
    })
    expect(files.map((f) => f.name)).toEqual(['frente.jpg', 'selo.webp'])
  })

  it('aceita múltiplas fotos e tolera drop vazio', () => {
    expect(
      imageFilesFromDataTransfer({
        files: [
          { name: 'a.png', type: 'image/png' },
          { name: 'b.gif', type: 'image/gif' },
        ],
      }),
    ).toHaveLength(2)
    expect(imageFilesFromDataTransfer(null)).toEqual([])
    expect(imageFilesFromDataTransfer({})).toEqual([])
  })
})
