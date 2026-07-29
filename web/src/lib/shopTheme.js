const STORAGE_KEY = 'puppe-shop-theme'

/** @returns {'dark'|'light'} */
export function getShopTheme() {
  try {
    const v = localStorage.getItem(STORAGE_KEY)
    if (v === 'light' || v === 'dark') return v
  } catch {
    // ignore
  }
  return 'dark'
}

/** @param {'dark'|'light'} theme */
export function setShopTheme(theme) {
  try {
    localStorage.setItem(STORAGE_KEY, theme)
  } catch {
    // ignore
  }
  applyShopThemeToDocument(theme)
}

/** @param {'dark'|'light'} theme */
export function applyShopThemeToDocument(theme) {
  const bg = theme === 'light' ? '#f4f6f8' : '#000000'
  const fg = theme === 'light' ? '#111827' : '#ffffff'
  document.documentElement.style.background = bg
  document.body.style.background = bg
  document.body.style.color = fg
  document.documentElement.setAttribute('data-shop-theme', theme)
  try {
    const meta = document.querySelector('meta[name="theme-color"]')
    if (meta) meta.setAttribute('content', bg)
  } catch {
    // ignore
  }
}

export function clearShopThemeFromDocument() {
  document.documentElement.style.background = ''
  document.body.style.background = ''
  document.body.style.color = ''
  document.documentElement.removeAttribute('data-shop-theme')
}

/** @param {'dark'|'light'} theme */
export function shopRootClass(theme) {
  return theme === 'light' ? 'shop shop-light' : 'shop'
}
