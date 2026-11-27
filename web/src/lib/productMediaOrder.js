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
