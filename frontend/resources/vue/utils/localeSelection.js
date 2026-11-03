export function selectBrowserLocale (availableLocales, browserLanguages) {
  for (const candidate of browserLanguages || []) {
    const lowerCandidate = candidate.toLowerCase()
    const exact = availableLocales.find(locale => locale.toLowerCase() === lowerCandidate)

    if (exact) {
      return exact
    }

    const parts = lowerCandidate.split('-')

    if (parts[0] === 'zh') {
      const script = new Intl.Locale(candidate).maximize().script
      const match = availableLocales.find(locale =>
        locale.toLowerCase().startsWith('zh-') &&
        new Intl.Locale(locale).maximize().script === script
      )

      if (match) {
        return match
      }

      continue
    }

    for (let length = parts.length - 1; length > 0; length--) {
      const prefix = `${parts.slice(0, length).join('-')}-`
      const match = availableLocales.find(locale => locale.toLowerCase().startsWith(prefix))

      if (match) {
        return match
      }
    }
  }

  return 'en'
}
