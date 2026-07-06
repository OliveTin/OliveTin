import test from 'node:test'
import assert from 'node:assert/strict'
import {
  allChoiceValues,
  choiceLabel,
  formatChecklistValue,
  parseChecklistValue,
  toggleChoice
} from './choiceChecklistHelpers.js'

const choices = [
  { title: 'Documents', value: 'documents' },
  { title: 'Photos', value: 'photos' }
]

test('parseChecklistValue splits comma-delimited values', () => {
  assert.deepEqual(parseChecklistValue('documents,photos'), ['documents', 'photos'])
  assert.deepEqual(parseChecklistValue('documents, photos'), ['documents', 'photos'])
  assert.deepEqual(parseChecklistValue(''), [])
})

test('formatChecklistValue joins selected values', () => {
  assert.equal(formatChecklistValue(['documents', 'photos']), 'documents,photos')
  assert.equal(formatChecklistValue([]), '')
})

test('toggleChoice adds and removes values', () => {
  assert.deepEqual(toggleChoice([], 'documents'), ['documents'])
  assert.deepEqual(toggleChoice(['documents'], 'photos'), ['documents', 'photos'])
  assert.deepEqual(toggleChoice(['documents', 'photos'], 'documents'), ['photos'])
})

test('choiceLabel prefers title over value', () => {
  assert.equal(choiceLabel(choices[0]), 'Documents')
  assert.equal(choiceLabel({ value: 'music' }), 'music')
})

test('allChoiceValues returns every choice value', () => {
  assert.deepEqual(allChoiceValues(choices), ['documents', 'photos'])
})
