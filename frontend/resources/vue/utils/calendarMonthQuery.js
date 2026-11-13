const MONTH_QUERY_PATTERN = /^(\d{4})-(\d{2})$/

export function formatCalendarMonthQuery (monthIndex, year) {
  return `${String(year).padStart(4, '0')}-${String(monthIndex + 1).padStart(2, '0')}`
}

// Falls back to the month of `today` so that a missing or hand-edited bad value still shows a calendar.
export function parseCalendarMonthQuery (value, today = new Date()) {
  const fallback = { monthIndex: today.getMonth(), year: today.getFullYear() }

  if (typeof value !== 'string') {
    return fallback
  }

  const match = MONTH_QUERY_PATTERN.exec(value)
  if (!match) {
    return fallback
  }

  const monthIndex = Number(match[2]) - 1
  if (monthIndex < 0 || monthIndex > 11) {
    return fallback
  }

  return { monthIndex, year: Number(match[1]) }
}
