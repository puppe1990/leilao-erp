import { describe, expect, it } from 'vitest'
import {
  applyVisibleMedia,
  firstPhotoIndex,
  mediaIds,
  mediaLeaveUnsavedMessage,
  mediaSaveFailedMessage,
  moveMediaAt,
  moveMediaItem,
  planMediaSave,
  shouldClearMediaDraft,
  shouldWarnMediaLeave,
} from './productMediaOrder.js'

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

  it('mostra o rascunho na tela antes do servidor confirmar', () => {
    const saved = list
    const draft = moveMediaAt(list, 0, 2)
    expect(mediaIds(applyVisibleMedia(saved, draft))).toEqual([2, 3, 1])
    expect(mediaIds(applyVisibleMedia(saved, null))).toEqual([1, 2, 3])
    expect(applyVisibleMedia(null, null)).toEqual([])
  })

  it('avisa se o save falhar ou se sair com rascunho pendente', () => {
    expect(mediaSaveFailedMessage()).toMatch(/Não deu pra salvar a ordem das fotos/)
    expect(mediaLeaveUnsavedMessage()).toMatch(/ainda não salvou/)
    expect(shouldWarnMediaLeave([{ id: 1 }], '/products/1/edit')).toBe(true)
    expect(shouldWarnMediaLeave([{ id: 1 }], '/products/1/media/reorder')).toBe(false)
    expect(shouldWarnMediaLeave(null, '/products')).toBe(false)
  })

  it('só limpa o rascunho quando o servidor bate com a tela ou quando falha', () => {
    const draft = moveMediaAt(list, 0, 2)
    expect(shouldClearMediaDraft(draft, draft, '')).toBe(true)
    expect(shouldClearMediaDraft(draft, list, '')).toBe(false)
    expect(shouldClearMediaDraft(draft, list, 'erro')).toBe(true)
    expect(shouldClearMediaDraft(null, list, '')).toBe(true)
  })

  it('enfileira o save se outro ainda está no ar', () => {
    expect(planMediaSave(false, [2, 1, 3])).toEqual({ send: [2, 1, 3], queue: false })
    expect(planMediaSave(true, [3, 2, 1])).toEqual({ send: null, queue: true })
    expect(planMediaSave(false, [])).toEqual({ send: null, queue: false })
  })
})
