export function readPrefilledArgumentsFromNavigation() {
	const state = window.history.state
	if (state?.prefilledArguments && typeof state.prefilledArguments === 'object') {
		return { ...state.prefilledArguments }
	}

	return {}
}

export function getInitialArgumentValue(paramName, prefilledArguments) {
	if (Object.prototype.hasOwnProperty.call(prefilledArguments, paramName)) {
		return prefilledArguments[paramName]
	}

	const params = new URLSearchParams(window.location.search.substring(1))
	return params.get(paramName)
}
