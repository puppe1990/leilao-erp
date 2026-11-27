export function moveMediaItem(list, index, delta) {
  const next = Array.isArray(list) ? list.slice() : []
  const j = index + delta
  if (index < 0 || j < 0 || index >= next.length || j >= next.length) {
    return next
  }
  const tmp = next[index]
  next[index] = next[j]
  next[j] = tmp
  return next
}

export function mediaIds(list) {
  return (Array.isArray(list) ? list : []).map((m) => m.id)
}

export function firstPhotoIndex(list) {
  if (!Array.isArray(list)) return -1
  return list.findIndex((m) => m?.kind === 'photo')
}
