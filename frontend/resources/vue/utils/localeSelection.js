export function selectBrowserLocale (availableLocales, browserLanguages) {
  for (const candidate of browserLanguages || []) {
    const lowerCandidate = candidate.toLowerCase()
    const exact = availableLocales.find(locale => locale.toLowerCase() === lowerCandidate)

    if (exact) {
      return exact
    }

    const language = lowerCandidate.split('-')[0]
    const prefix = availableLocales.find(locale => locale.toLowerCase().startsWith(`${language}-`))

    if (prefix) {
      return prefix
    }
  }

  return 'en'
}
