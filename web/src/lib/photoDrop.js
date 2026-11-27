const PHOTO_TYPES = ['image/jpeg', 'image/png', 'image/webp', 'image/gif']

/** Files dropped on the photo box, keeping only the formats the input accepts. */
export function imageFilesFromDataTransfer(dataTransfer) {
  const files = Array.from(dataTransfer?.files || [])
  return files.filter((file) => PHOTO_TYPES.includes(String(file?.type || '').toLowerCase()))
}
