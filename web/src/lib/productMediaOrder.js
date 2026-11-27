export function moveMediaItem(list, index, delta) {
  return moveMediaAt(list, index, index + delta)
}

export function moveMediaAt(list, from, to) {
  const next = Array.isArray(list) ? list.slice() : []
  if (from === to) return next
  if (from < 0 || to < 0 || from >= next.length || to >= next.length) {
    return next
  }
  const [item] = next.splice(from, 1)
  next.splice(to, 0, item)
  return next
}

export function mediaIds(list) {
  return (Array.isArray(list) ? list : []).map((m) => m.id)
}

export function firstPhotoIndex(list) {
  if (!Array.isArray(list)) return -1
  return list.findIndex((m) => m?.kind === 'photo')
}

export function applyVisibleMedia(saved, draft) {
  if (Array.isArray(draft)) return draft
  return Array.isArray(saved) ? saved : []
}

export function mediaSaveFailedMessage() {
  return 'Não deu pra salvar a ordem das fotos. A lista voltou.'
}

export function mediaLeaveUnsavedMessage() {
  return 'A ordem das fotos ainda não salvou. Sair mesmo assim?'
}

export function shouldWarnMediaLeave(draft, visitUrl) {
  if (!Array.isArray(draft)) return false
  return !String(visitUrl || '').includes('/media/reorder')
}

export function shouldClearMediaDraft(draft, saved, formError) {
  if (formError) return true
  if (!Array.isArray(draft)) return true
  return mediaIds(draft).join() === mediaIds(saved).join()
}

export function planMediaSave(inFlight, latestIds) {
  if (!Array.isArray(latestIds) || latestIds.length === 0) {
    return { send: null, queue: false }
  }
  if (inFlight) return { send: null, queue: true }
  return { send: latestIds, queue: false }
}
