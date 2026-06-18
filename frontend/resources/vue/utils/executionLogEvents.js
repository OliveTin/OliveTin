export function cloneLogEntry (logEntry) {
  if (!logEntry) {
    return null
  }

  return {
    ...logEntry,
    tags: Array.isArray(logEntry.tags) ? [...logEntry.tags] : logEntry.tags
  }
}

export function getExecutionLogEntry (evt) {
  const logEntry = evt?.payload?.logEntry ?? evt?.detail?.logEntry ?? null
  return cloneLogEntry(logEntry)
}

export function updateLogEntryInList (entries, logEntry) {
  if (!logEntry?.executionTrackingId || !entries) {
    return false
  }

  const index = entries.findIndex(
    item => item.executionTrackingId === logEntry.executionTrackingId
  )
  if (index < 0) {
    return false
  }

  entries[index] = logEntry
  return true
}

export function updateLogEntryInGroups (groups, logEntry) {
  const entry = cloneLogEntry(logEntry)
  if (!entry?.executionTrackingId || !groups) {
    return null
  }

  let firstMatch = null

  for (const group of groups) {
    for (const action of group.actions || []) {
      const entries = action.entries || []
      const index = entries.findIndex(
        item => item.executionTrackingId === entry.executionTrackingId
      )
      if (index < 0) {
        continue
      }

      const previous = entries[index]
      entries[index] = entry
      if (!firstMatch) {
        firstMatch = { group, action, index, previous }
      }
    }
  }

  return firstMatch
}
