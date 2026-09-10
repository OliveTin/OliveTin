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

test('selectBrowserLocale uses English when no locale matches', () => {
  assert.equal(selectBrowserLocale(availableLocales, ['fr-FR']), 'en')
})
