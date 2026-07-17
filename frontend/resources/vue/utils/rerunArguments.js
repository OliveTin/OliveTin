import { needsArgumentForm } from './needsArgumentForm.js'
import { actionRequiresJustification } from './justificationTemplate.js'

const nonStorableArgumentTypes = new Set([
  'password',
  'very_dangerous_raw_string'
])

function isNonStorableArgumentType (type) {
  return nonStorableArgumentTypes.has(type)
}

function argumentSkipsValidation (type) {
  return type === 'confirmation' || type === 'html'
}

/**
 * Mirrors backend restartArgumentRequired: an argument needs a stored value
 * when it is validated and has no default. Proto ActionArgument has no
 * `required` field — only `defaultValue`.
 */
function rerunArgumentRequired (arg) {
  if (argumentSkipsValidation(arg?.type)) {
    return false
  }

  const defaultValue = arg?.defaultValue ?? ''
  return defaultValue === ''
}

export function logEntryArgumentsToStartActionArgs (logEntry) {
  return (logEntry?.arguments ?? []).map((arg) => ({
    name: arg.name,
    value: arg.value
  }))
}

export function rerunNeedsArgumentForm (action, logEntry) {
  // Always re-prompt when justification is required so each execution is
  // explicitly justified rather than silently reusing a prior reason.
  if (actionRequiresJustification(action?.justification)) {
    return true
  }

  if (!needsArgumentForm(action)) {
    return false
  }

  return hasMissingRerunArguments(action, logEntry?.arguments ?? [])
}

export function hasMissingRerunArguments (action, storedArgs) {
  const stored = new Map(storedArgs.map((arg) => [arg.name, arg.value]))

  for (const arg of action?.arguments ?? []) {
    if (isNonStorableArgumentType(arg.type)) {
      return true
    }

    if (rerunArgumentRequired(arg) && !stored.has(arg.name)) {
      return true
    }
  }

  return false
}

/**
 * Builds history.state.prefilledArguments for ArgumentForm (not URL query),
 * matching ActionButton's prefill pattern and keeping values out of the URL.
 */
export function buildRerunPrefilledArguments (logEntry) {
  const prefilled = {}

  for (const arg of logEntry?.arguments ?? []) {
    prefilled[arg.name] = arg.value
  }

  return prefilled
}

export function buildRerunStartActionArgs (bindingId, logEntry) {
  return {
    bindingId,
    arguments: logEntryArgumentsToStartActionArgs(logEntry)
  }
}
