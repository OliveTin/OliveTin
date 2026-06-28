import test from 'node:test'
import assert from 'node:assert/strict'
import {
  displayLabelForModelValue,
  normalizedModelValue,
  syncStateFromModelValue
} from './choiceComboboxHelpers.js'

const choices = [
  { title: 'Production', value: 'prod' },
  { title: 'Staging', value: 'stage' }
]

test('displayLabelForModelValue returns the choice label for valid values', () => {
  assert.equal(displayLabelForModelValue(choices, 'prod'), 'Production')
})

test('displayLabelForModelValue clears invalid enum values instead of echoing them', () => {
  assert.equal(displayLabelForModelValue(choices, 'missing'), '')
})

test('normalizedModelValue keeps valid values and clears invalid ones', () => {
  assert.equal(normalizedModelValue(choices, 'stage'), 'stage')
  assert.equal(normalizedModelValue(choices, 'missing'), '')
  assert.equal(normalizedModelValue(choices, ''), '')
})

test('syncStateFromModelValue clears invalid selections for closed-state sync', () => {
  assert.deepEqual(syncStateFromModelValue(choices, 'missing'), {
    query: '',
    modelValue: ''
  })
})

test('syncStateFromModelValue preserves valid selections for closed-state sync', () => {
  assert.deepEqual(syncStateFromModelValue(choices, 'stage'), {
    query: 'Staging',
    modelValue: 'stage'
  })
})
