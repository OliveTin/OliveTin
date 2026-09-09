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

  const response = await fetch(themeUrl, { cache: 'no-store' })
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
