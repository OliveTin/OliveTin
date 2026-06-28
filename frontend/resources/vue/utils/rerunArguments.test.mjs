import test from 'node:test'
import assert from 'node:assert/strict'
import {
  buildArgumentFormQuery,
  buildRerunStartActionArgs,
  hasMissingRerunArguments,
  logEntryArgumentsToStartActionArgs,
  rerunNeedsArgumentForm
} from '../utils/rerunArguments.js'

test('logEntryArgumentsToStartActionArgs maps proto arguments for StartAction', () => {
  assert.deepEqual(
    logEntryArgumentsToStartActionArgs({
      arguments: [
        { name: 'host', value: 'example.com' },
        { name: 'port', value: '443' }
      ]
    }),
    [
      { name: 'host', value: 'example.com' },
      { name: 'port', value: '443' }
    ]
  )
})

test('hasMissingRerunArguments requires password fields to be re-entered', () => {
  const action = {
    arguments: [
      { name: 'user', type: 'ascii_identifier', required: true },
      { name: 'pass', type: 'password', required: true }
    ]
  }

  assert.equal(
    hasMissingRerunArguments(action, [{ name: 'user', value: 'alice' }]),
    true
  )
})

test('hasMissingRerunArguments requires very_dangerous_raw_string fields to be re-entered', () => {
  const action = {
    arguments: [
      { name: 'host', type: 'ascii_identifier', required: true },
      { name: 'payload', type: 'very_dangerous_raw_string', required: false }
    ]
  }

  assert.equal(
    hasMissingRerunArguments(action, [{ name: 'host', value: 'db-1' }]),
    true
  )
})

test('hasMissingRerunArguments detects missing required stored arguments', () => {
  const action = {
    arguments: [{ name: 'host', type: 'ascii_identifier', required: true }]
  }

  assert.equal(hasMissingRerunArguments(action, []), true)
  assert.equal(
    hasMissingRerunArguments(action, [{ name: 'host', value: 'db-1' }]),
    false
  )
})

test('rerunNeedsArgumentForm can start directly when stored args are complete', () => {
  const action = {
    arguments: [{ name: 'host', type: 'ascii_identifier', required: true }]
  }
  const logEntry = {
    arguments: [{ name: 'host', value: 'db-1' }]
  }

  assert.equal(rerunNeedsArgumentForm(action, logEntry), false)
})

test('rerunNeedsArgumentForm opens the form when justification is missing', () => {
  const action = { justification: true, arguments: [] }

  assert.equal(rerunNeedsArgumentForm(action, {}), true)
  assert.equal(
    rerunNeedsArgumentForm(action, { justification: 'approved change' }),
    false
  )
})

test('buildRerunStartActionArgs includes stored justification', () => {
  assert.deepEqual(
    buildRerunStartActionArgs('binding-1', {
      arguments: [{ name: 'host', value: 'db-1' }],
      justification: 'maintenance window'
    }, {
      justification: true,
      arguments: [{ name: 'host', type: 'ascii_identifier' }]
    }),
    {
      bindingId: 'binding-1',
      arguments: [{ name: 'host', value: 'db-1' }],
      justification: 'maintenance window'
    }
  )
})

test('buildArgumentFormQuery prefills non-password stored arguments', () => {
  assert.deepEqual(
    buildArgumentFormQuery({
      arguments: [
        { name: 'host', value: 'db-1' },
        { name: 'port', value: '5432' }
      ]
    }),
    {
      host: 'db-1',
      port: '5432'
    }
  )
})
