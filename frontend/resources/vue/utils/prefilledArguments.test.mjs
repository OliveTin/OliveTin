import test from 'node:test'
import assert from 'node:assert/strict'

import { getInitialArgumentValue, readPrefilledArgumentsFromNavigation } from './prefilledArguments.js'

test('readPrefilledArgumentsFromNavigation returns navigation state values', () => {
	assert.deepEqual(
		readPrefilledArgumentsFromNavigation({ prefilledArguments: { ansible_host: '10.0.0.1' } }),
		{ ansible_host: '10.0.0.1' }
	)
})

test('readPrefilledArgumentsFromNavigation returns empty object when state is absent', () => {
	assert.deepEqual(readPrefilledArgumentsFromNavigation({}), {})
	assert.deepEqual(readPrefilledArgumentsFromNavigation(undefined), {})
})

test('getInitialArgumentValue prefers navigation state over query params', () => {
	assert.equal(
		getInitialArgumentValue(
			'ansible_host',
			{ ansible_host: '10.0.0.1' },
			'?ansible_host=10.0.0.2'
		),
		'10.0.0.1'
	)
})

test('getInitialArgumentValue falls back to query params when state is absent', () => {
	assert.equal(
		getInitialArgumentValue('ansible_host', {}, '?ansible_host=10.0.0.2'),
		'10.0.0.2'
	)
})
