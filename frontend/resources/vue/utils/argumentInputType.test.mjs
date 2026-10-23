import test from 'node:test'
import assert from 'node:assert/strict'

import { argumentInputType } from './argumentInputType.js'

test('argumentInputType renders int arguments as number inputs', () => {
  assert.equal(argumentInputType({ type: 'int' }), 'number')
})

test('argumentInputType maps argument types to valid HTML input types', () => {
  assert.equal(argumentInputType({ type: 'confirmation' }), 'checkbox')
  assert.equal(argumentInputType({ type: 'checkbox' }), 'checkbox')
  assert.equal(argumentInputType({ type: 'datetime' }), 'datetime-local')
  assert.equal(argumentInputType({ type: 'email' }), 'email')
  assert.equal(argumentInputType({ type: 'password' }), 'password')
  assert.equal(argumentInputType({ type: 'url' }), 'url')
})

test('argumentInputType renders validation-only types as text inputs', () => {
  for (const type of ['', 'ascii', 'ascii_identifier', 'ascii_sentence', 'dnsname', 'shell_safe_identifier', 'unicode_identifier', 'very_dangerous_raw_string', 'regex:^[a-z]+$']) {
    assert.equal(argumentInputType({ type }), 'text', type)
  }
})

test('argumentInputType has no input type for non-input arguments', () => {
  for (const type of ['html', 'raw_string_multiline', 'select']) {
    assert.equal(argumentInputType({ type }), undefined, type)
  }
})
