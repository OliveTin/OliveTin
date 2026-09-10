import test from 'node:test'
import assert from 'node:assert/strict'

function installThemeDom () {
  const styleEl = { textContent: '' }
  const body = { attributes: {}, setAttribute (name, value) { this.attributes[name] = value } }

  globalThis.document = {
    getElementById: (id) => (id === 'theme-style' ? styleEl : null),
    createElement: () => styleEl,
    head: { appendChild: () => {} },
    body
  }

  return { styleEl, body }
}

test('applyThemeStyles times out when the response body stalls', async (t) => {
  const originalFetch = globalThis.fetch
  const originalDocument = globalThis.document
  const originalAbortSignal = globalThis.AbortSignal

  t.after(() => {
    globalThis.fetch = originalFetch
    globalThis.document = originalDocument
    globalThis.AbortSignal = originalAbortSignal
  })

  installThemeDom()

  globalThis.AbortSignal = {
    timeout () {
      const controller = new AbortController()
      setTimeout(() => controller.abort(), 50)
      return controller.signal
    }
  }

  globalThis.fetch = (_url, options) => Promise.resolve({
    ok: true,
    status: 200,
    text () {
      return new Promise((_resolve, reject) => {
        options.signal.addEventListener('abort', () => {
          reject(new DOMException('The operation was aborted.', 'AbortError'))
        })
      })
    }
  })

  const { applyThemeStyles } = await import('./themeLoader.js')

  await assert.rejects(
    () => applyThemeStyles(''),
    (err) => err.name === 'AbortError' || err.name === 'TimeoutError'
  )
})
