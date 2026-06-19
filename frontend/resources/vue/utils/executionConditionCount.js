function nonEmptyList (list) {
  return Array.isArray(list) && list.length > 0
}

export function countExecutionConditions (action) {
  if (!action) {
    return 0
  }

  let count = 1

  if (action.execOnStartup) {
    count++
  }
  if (nonEmptyList(action.execOnCron)) {
    count++
  }
  if (nonEmptyList(action.execOnFileCreatedInDir)) {
    count++
  }
  if (nonEmptyList(action.execOnFileChangedInDir)) {
    count++
  }
  if (action.execOnCalendarFile) {
    count++
  }
  if (nonEmptyList(action.execOnWebhooks)) {
    count++
  }

  return count
}
