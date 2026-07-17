import test from 'node:test'
import assert from 'node:assert/strict'

import {
  ARGUMENT_FIELD_ID_PREFIX,
  argumentFieldChoicesId,
  argumentFieldId,
  argumentFieldListboxId,
  argumentFieldListboxOptionId,
  argumentFieldOptionId,
  argumentFieldValidationElementId,
  argumentFieldValueId,
  argumentFieldWrapperId
} from './argumentFieldIds.js'

test('argumentFieldId namespaces argument names with a field role', () => {
  assert.equal(argumentFieldId('content'), 'arg-field-content')
  assert.equal(argumentFieldId('banner'), 'arg-field-banner')
  assert.equal(argumentFieldId('confirm'), 'arg-field-confirm')
})

test('role-prefixed ids avoid collisions between related argument elements', () => {
  assert.notEqual(argumentFieldId('content'), argumentFieldChoicesId('content'))
  assert.notEqual(argumentFieldId('content-choices'), argumentFieldChoicesId('content'))
  assert.notEqual(argumentFieldId('segments-value'), argumentFieldValueId('segments'))
  assert.notEqual(argumentFieldId('segments-wrapper'), argumentFieldWrapperId('segments'))
  assert.notEqual(argumentFieldId('segments-0'), argumentFieldOptionId('segments', 0))
  assert.notEqual(argumentFieldId('host-listbox'), argumentFieldListboxId('host'))
  assert.notEqual(
    argumentFieldId('host-listbox-option-0'),
    argumentFieldListboxOptionId('host', 0)
  )
})

test('argumentFieldChoicesId uses a distinct choices role', () => {
  assert.equal(argumentFieldChoicesId('content'), 'arg-choices-content')
})

test('argumentFieldValidationElementId uses checklist value id', () => {
  assert.equal(
    argumentFieldValidationElementId({ name: 'segments', type: 'checklist' }),
    'arg-value-segments'
  )
})

test('argumentFieldValidationElementId uses field id for other argument types', () => {
  assert.equal(
    argumentFieldValidationElementId({ name: 'content', type: 'raw_string_multiline' }),
    'arg-field-content'
  )
  assert.equal(
    argumentFieldValidationElementId({ name: 'datetime', type: 'datetime' }),
    'arg-field-datetime'
  )
})

test('namespaced ids avoid known app-shell element ids', () => {
  const appShellIds = [
    'content',
    'banner',
    'layout',
    'mainnav',
    'app',
    'big-error',
    'available-version',
    'link-login',
    'username-text',
    'theme-style',
    'olivetin-custom-js',
    'justification',
    'username',
    'password',
    'connection-banner',
    'execution-results-popup',
    'logs-filter-suggestions',
    'argument-popup'
  ]

  for (const shellId of appShellIds) {
    assert.notEqual(argumentFieldId(shellId), shellId)
    assert.ok(argumentFieldId(shellId).startsWith(ARGUMENT_FIELD_ID_PREFIX))
    assert.notEqual(argumentFieldChoicesId(shellId), shellId)
    assert.notEqual(argumentFieldValueId(shellId), shellId)
  }
})
