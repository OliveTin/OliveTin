const fallbackNamedHtmlEntities = {
	amp: '&',
	apos: "'",
	darr: '\u2193',
	gt: '>',
	laquo: '\u00ab',
	larr: '\u2190',
	nbsp: '\u00a0',
	quot: '"',
	raquo: '\u00bb',
	rarr: '\u2192',
	uarr: '\u2191',
}

export function decodeHtmlEntities(text) {
	if (typeof document !== 'undefined') {
		const textarea = document.createElement('textarea')
		textarea.innerHTML = text

		return textarea.value
	}

	return text.replace(/&#x([0-9a-fA-F]+);?/g, (_, hex) => {
		const codePoint = Number.parseInt(hex, 16)
		return Number.isFinite(codePoint) ? String.fromCodePoint(codePoint) : ''
	}).replace(/&#(\d+);?/g, (_, decimal) => {
		const codePoint = Number.parseInt(decimal, 10)
		return Number.isFinite(codePoint) ? String.fromCodePoint(codePoint) : ''
	}).replace(/&([a-zA-Z][a-zA-Z0-9]+);?/g, (entity, name) => {
		return fallbackNamedHtmlEntities[name] ?? entity
	})
}

export function glyphLooksLikeHtml(text) {
	const trimmedText = text.trim()

	return trimmedText.startsWith('<') || /<img\b/i.test(text) || /\/custom-webui\//i.test(text)
}
