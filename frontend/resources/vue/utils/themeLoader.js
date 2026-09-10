const DEFAULT_THEME_FETCH_TIMEOUT_MS = 5000

export async function applyThemeStyles (themePreference = '') {
  let themeStyle = document.getElementById('theme-style')

  if (!themeStyle) {
    themeStyle = document.createElement('style')
    themeStyle.id = 'theme-style'
    themeStyle.type = 'text/css'
    document.head.appendChild(themeStyle)
  }

  const themeUrl = themePreference && themePreference !== ''
    ? `/custom-webui/themes/${encodeURIComponent(themePreference)}/theme.css`
    : '/theme.css'

  const response = await fetch(themeUrl, {
    cache: 'no-store',
    signal: AbortSignal.timeout(DEFAULT_THEME_FETCH_TIMEOUT_MS)
  })
  if (!response.ok) {
    throw new Error(`theme fetch failed: ${response.status}`)
  }

  const css = await response.text()
  themeStyle.textContent = `@layer theme { ${css} }`
  document.body.setAttribute('loaded-theme', themeUrl)
}

export function getStoredThemePreference () {
  return localStorage.getItem('olivetin-theme') || ''
}
