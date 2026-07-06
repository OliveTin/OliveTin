import test from 'node:test'
import assert from 'node:assert/strict'

import { getInitialArgumentValue, readPrefilledArgumentsFromNavigation } from './prefilledArguments.js'

test('readPrefilledArgumentsFromNavigation returns navigation state values', () => {
	const originalState = window.history.state
	window.history.replaceState({ prefilledArguments: { ansible_host: '10.0.0.1' } }, '')

	assert.deepEqual(readPrefilledArgumentsFromNavigation(), { ansible_host: '10.0.0.1' })

	window.history.replaceState(originalState, '')
})

test('getInitialArgumentValue prefers navigation state over query params', () => {
	const originalState = window.history.state
	const originalSearch = window.location.search

	window.history.replaceState({ prefilledArguments: { ansible_host: '10.0.0.1' } }, '')
	window.history.replaceState(window.history.state, '', '?ansible_host=10.0.0.2')

	assert.equal(getInitialArgumentValue('ansible_host', readPrefilledArgumentsFromNavigation()), '10.0.0.1')

	window.history.replaceState(originalState, '')
	window.history.replaceState(window.history.state, '', originalSearch || '/')
})

test('getInitialArgumentValue falls back to query params when state is absent', () => {
	const originalState = window.history.state
	const originalSearch = window.location.search

	window.history.replaceState({}, '')
	window.history.replaceState(window.history.state, '', '?ansible_host=10.0.0.2')

	assert.equal(getInitialArgumentValue('ansible_host', readPrefilledArgumentsFromNavigation()), '10.0.0.2')

	window.history.replaceState(originalState, '')
	window.history.replaceState(window.history.state, '', originalSearch || '/')
})
