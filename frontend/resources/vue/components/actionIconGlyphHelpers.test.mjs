import test from 'node:test'
import assert from 'node:assert/strict'
import { decodeHtmlEntities, glyphLooksLikeHtml } from './actionIconGlyphHelpers.mjs'

test('decodeHtmlEntities decodes named entity icons as plain glyph text', () => {
	assert.equal(decodeHtmlEntities('&laquo;'), '\u00ab')
	assert.equal(decodeHtmlEntities('&rarr;'), '\u2192')
	assert.equal(decodeHtmlEntities('&laquo; next &rarr;'), '\u00ab next \u2192')
})

test('decoded named entity icons are not treated as HTML markup', () => {
	const decodedGlyph = decodeHtmlEntities('&rarr;')

	assert.equal(glyphLooksLikeHtml(decodedGlyph), false)
})

test('decodeHtmlEntities keeps existing numeric entity icon support', () => {
	assert.equal(decodeHtmlEntities('&#x1f4a9;'), '\ud83d\udca9')
	assert.equal(decodeHtmlEntities('&#128190;'), '\ud83d\udcbe')
})
