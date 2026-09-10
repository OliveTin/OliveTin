const DEFAULT_THEME_FETCH_TIMEOUT_MS = 5000

async function fetchWithTimeout (url, options, timeoutMs) {
  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), timeoutMs)
  try {
    return await fetch(url, { ...options, signal: controller.signal })
  } finally {
    clearTimeout(timeoutId)
  }
}

export async function applyThemeStyles (themePreference = '', timeoutMs = DEFAULT_THEME_FETCH_TIMEOUT_MS) {
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

  const response = await fetchWithTimeout(themeUrl, { cache: 'no-store' }, timeoutMs)
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
