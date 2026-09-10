import { selectBrowserLocale } from './utils/localeSelection.js'

const localePathPrefix = '../../../lang/generated/'
const localeModules = import.meta.glob('../../../lang/generated/*.json', {
  import: 'default'
})

export const availableLocales = Object.keys(localeModules)
  .map(path => path.slice(localePathPrefix.length, -'.json'.length))
  .sort()

export function resolveBrowserLocale (browserLanguages = navigator.languages) {
  return selectBrowserLocale(availableLocales, browserLanguages)
}

export function getSelectedLocale () {
  const storedLanguage = localStorage.getItem('olivetin-language')

  if (storedLanguage && storedLanguage !== 'auto' && availableLocales.includes(storedLanguage)) {
    return storedLanguage
  }

  if (storedLanguage === 'auto') {
    localStorage.removeItem('olivetin-language')
  }

  return resolveBrowserLocale()
}

export async function loadLocaleMessages (locale) {
  const loadLocale = localeModules[`${localePathPrefix}${locale}.json`]

  if (!loadLocale) {
    throw new Error(`Unsupported locale: ${locale}`)
  }

  return loadLocale()
}

export async function loadInitialMessages (locale) {
  const locales = locale === 'en' ? ['en'] : ['en', locale]
  const loadedMessages = await Promise.all(locales.map(async currentLocale => {
    return [currentLocale, await loadLocaleMessages(currentLocale)]
  }))

  return Object.fromEntries(loadedMessages)
}

export async function ensureLocaleMessages (i18n, locale) {
  if (!i18n.availableLocales.includes(locale)) {
    i18n.setLocaleMessage(locale, await loadLocaleMessages(locale))
  }
}
