export function readPrefilledArgumentsFromNavigation() {
	const state = window.history.state
	if (state?.prefilledArguments && typeof state.prefilledArguments === 'object') {
		return { ...state.prefilledArguments }
	}

	return {}
}

export function getInitialArgumentValue(paramName, prefilledArguments = {}) {
	const safePrefilledArguments = prefilledArguments && typeof prefilledArguments === 'object'
		? prefilledArguments
		: {}

	if (Object.prototype.hasOwnProperty.call(safePrefilledArguments, paramName)) {
		return safePrefilledArguments[paramName]
	}

	const params = new URLSearchParams(window.location.search)
	return params.get(paramName)
}
