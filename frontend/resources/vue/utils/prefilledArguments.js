export function readPrefilledArgumentsFromNavigation (historyState = globalThis.window?.history?.state) {
  if (historyState?.prefilledArguments && typeof historyState.prefilledArguments === 'object') {
    return { ...historyState.prefilledArguments }
  }

  return {}
}

export function getInitialArgumentValue (paramName, prefilledArguments = {}, search = globalThis.window?.location?.search ?? '') {
  const safePrefilledArguments = prefilledArguments && typeof prefilledArguments === 'object'
    ? prefilledArguments
    : {}

  if (Object.prototype.hasOwnProperty.call(safePrefilledArguments, paramName)) {
    return safePrefilledArguments[paramName]
  }

  const params = new URLSearchParams(search)
  return params.get(paramName)
}
