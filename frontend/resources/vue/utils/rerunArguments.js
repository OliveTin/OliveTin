import { needsArgumentForm } from './needsArgumentForm.js'

const nonStorableArgumentTypes = new Set([
  'password',
  'very_dangerous_raw_string'
])

function isNonStorableArgumentType (type) {
  return nonStorableArgumentTypes.has(type)
}

export function logEntryArgumentsToStartActionArgs (logEntry) {
  return (logEntry?.arguments ?? []).map((arg) => ({
    name: arg.name,
    value: arg.value
  }))
}

export function rerunNeedsArgumentForm (action, logEntry) {
  if (action?.justification && !logEntry?.justification) {
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

    if (arg.required && !stored.has(arg.name)) {
      return true
    }
  }

  return false
}

export function buildArgumentFormQuery (logEntry) {
  const query = {}

  for (const arg of logEntry?.arguments ?? []) {
    query[arg.name] = arg.value
  }

  return query
}

export function buildRerunStartActionArgs (bindingId, logEntry, action) {
  const startActionArgs = {
    bindingId,
    arguments: logEntryArgumentsToStartActionArgs(logEntry)
  }

  if (action?.justification && logEntry?.justification) {
    startActionArgs.justification = logEntry.justification
  }

  return startActionArgs
}
