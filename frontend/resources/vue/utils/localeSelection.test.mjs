import assert from 'node:assert/strict'
import test from 'node:test'

import { selectBrowserLocale } from './localeSelection.js'

const availableLocales = ['de-DE', 'en', 'es-ES', 'zh-Hans-CN', 'zh-Hant-TW']

test('selectBrowserLocale prefers an exact locale match', () => {
  assert.equal(selectBrowserLocale(availableLocales, ['es-ES']), 'es-ES')
})

test('selectBrowserLocale falls back to a matching language', () => {
  assert.equal(selectBrowserLocale(availableLocales, ['de-AT']), 'de-DE')
})

test('selectBrowserLocale preserves script subtags during fallback', () => {
  assert.equal(selectBrowserLocale(availableLocales, ['zh-Hant-HK']), 'zh-Hant-TW')
})

for (const language of ['zh-TW', 'zh-HK', 'zh-MO', 'zh-Hant', 'zh-Hant-CN', 'ZH-tw']) {
  test(`selectBrowserLocale selects Traditional Chinese for ${language}`, () => {
    assert.equal(selectBrowserLocale(availableLocales, [language]), 'zh-Hant-TW')
  })
}

for (const language of ['zh-CN', 'zh-SG', 'zh-Hans', 'zh-Hans-TW', 'zh']) {
  test(`selectBrowserLocale selects Simplified Chinese for ${language} regardless of locale order`, () => {
    assert.equal(selectBrowserLocale([...availableLocales].reverse(), [language]), 'zh-Hans-CN')
  })
}

test('selectBrowserLocale skips unavailable Chinese scripts and tries the next browser language', () => {
  assert.equal(selectBrowserLocale(['en', 'de-DE', 'zh-Hans-CN'], ['zh-TW', 'de-AT']), 'de-DE')
  assert.equal(selectBrowserLocale(['en', 'zh-Hant-TW'], ['zh-CN']), 'en')
})

test('selectBrowserLocale uses English when no locale matches', () => {
  assert.equal(selectBrowserLocale(availableLocales, ['fr-FR']), 'en')
})
