import { describe, expect, it } from 'vitest'
import { firstPhotoIndex, mediaIds, moveMediaAt, moveMediaItem } from './productMediaOrder.js'

describe('productMediaOrder', () => {
  const list = [
    { id: 1, kind: 'photo' },
    { id: 2, kind: 'photo' },
    { id: 3, kind: 'video' },
  ]

  it('sobe o item e devolve ids na nova ordem', () => {
    const next = moveMediaItem(list, 1, -1)
    expect(mediaIds(next)).toEqual([2, 1, 3])
    expect(list[0].id).toBe(1)
  })

  it('não move além das pontas', () => {
    expect(mediaIds(moveMediaItem(list, 0, -1))).toEqual([1, 2, 3])
    expect(mediaIds(moveMediaItem(list, 2, 1))).toEqual([1, 2, 3])
  })

  it('arrasta um item para outro índice sem só trocar vizinhos', () => {
    const next = moveMediaAt(list, 0, 2)
    expect(mediaIds(next)).toEqual([2, 3, 1])
    expect(list[0].id).toBe(1)
    expect(mediaIds(moveMediaAt(list, 2, 0))).toEqual([3, 1, 2])
    expect(mediaIds(moveMediaAt(list, 1, 1))).toEqual([1, 2, 3])
    expect(mediaIds(moveMediaAt(list, -1, 0))).toEqual([1, 2, 3])
    expect(mediaIds(moveMediaAt(list, 0, 9))).toEqual([1, 2, 3])
  })

  it('marca a primeira foto como capa', () => {
    expect(firstPhotoIndex(list)).toBe(0)
    expect(firstPhotoIndex([{ id: 9, kind: 'video' }, { id: 8, kind: 'photo' }])).toBe(1)
  })
})
