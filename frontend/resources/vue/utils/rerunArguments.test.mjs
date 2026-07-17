import test from 'node:test'
import assert from 'node:assert/strict'
import {
  buildRerunPrefilledArguments,
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
      { name: 'user', type: 'ascii_identifier' },
      { name: 'pass', type: 'password' }
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
      { name: 'host', type: 'ascii_identifier' },
      { name: 'payload', type: 'very_dangerous_raw_string' }
    ]
  }

  assert.equal(
    hasMissingRerunArguments(action, [{ name: 'host', value: 'db-1' }]),
    true
  )
})

test('hasMissingRerunArguments detects missing stored args without relying on required flag', () => {
  // Proto ActionArgument has no `required` field — only defaultValue.
  const action = {
    arguments: [{ name: 'host', type: 'ascii_identifier' }]
  }

  assert.equal(hasMissingRerunArguments(action, []), true)
  assert.equal(
    hasMissingRerunArguments(action, [{ name: 'host', value: 'db-1' }]),
    false
  )
})

test('hasMissingRerunArguments treats empty defaultValue as required', () => {
  const action = {
    arguments: [{ name: 'host', type: 'ascii_identifier', defaultValue: '' }]
  }

  assert.equal(hasMissingRerunArguments(action, []), true)
})

test('hasMissingRerunArguments allows args that have a defaultValue', () => {
  const action = {
    arguments: [{ name: 'host', type: 'ascii_identifier', defaultValue: 'example.com' }]
  }

  assert.equal(hasMissingRerunArguments(action, []), false)
})

test('hasMissingRerunArguments ignores confirmation and html args', () => {
  const action = {
    arguments: [
      { name: 'confirm', type: 'confirmation' },
      { name: 'help', type: 'html' },
      { name: 'host', type: 'ascii_identifier' }
    ]
  }

  assert.equal(
    hasMissingRerunArguments(action, [{ name: 'host', value: 'db-1' }]),
    false
  )
})

test('rerunNeedsArgumentForm can start directly when stored args are complete', () => {
  const action = {
    arguments: [{ name: 'host', type: 'ascii_identifier' }]
  }
  const logEntry = {
    arguments: [{ name: 'host', value: 'db-1' }]
  }

  assert.equal(rerunNeedsArgumentForm(action, logEntry), false)
})

test('rerunNeedsArgumentForm always opens the form when justification is required', () => {
  const action = {
    justification: ' ',
    arguments: [{ name: 'host', type: 'ascii_identifier' }]
  }
  const logEntry = {
    arguments: [{ name: 'host', value: 'db-1' }],
    justification: 'approved change'
  }

  assert.equal(rerunNeedsArgumentForm(action, logEntry), true)
  assert.equal(rerunNeedsArgumentForm(action, {}), true)
})

test('buildRerunStartActionArgs uses stored arguments without justification', () => {
  assert.deepEqual(
    buildRerunStartActionArgs('binding-1', {
      arguments: [{ name: 'host', value: 'db-1' }],
      justification: 'maintenance window'
    }),
    {
      bindingId: 'binding-1',
      arguments: [{ name: 'host', value: 'db-1' }]
    }
  )
})

test('buildRerunPrefilledArguments maps stored args for history.state', () => {
  assert.deepEqual(
    buildRerunPrefilledArguments({
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

test('buildRerunPrefilledArguments returns empty object when nothing was stored', () => {
  assert.deepEqual(buildRerunPrefilledArguments({}), {})
  assert.deepEqual(buildRerunPrefilledArguments(undefined), {})
})
