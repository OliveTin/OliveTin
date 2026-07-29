export const LOGS_FILTER_STORAGE_KEY = 'olivetin-logs-filter'

export function loadStoredLogsFilter () {
  try {
    return sessionStorage.getItem(LOGS_FILTER_STORAGE_KEY) || ''
  } catch {
    return ''
  }
}

export function storeLogsFilter (value) {
  try {
    if (value) {
      sessionStorage.setItem(LOGS_FILTER_STORAGE_KEY, value)
    } else {
      sessionStorage.removeItem(LOGS_FILTER_STORAGE_KEY)
    }
  } catch {
    // Ignore storage failures (private mode, quota, etc.)
  }
}
