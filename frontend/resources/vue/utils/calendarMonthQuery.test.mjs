import assert from 'node:assert/strict'
import test from 'node:test'

import { formatCalendarMonthQuery, parseCalendarMonthQuery } from './calendarMonthQuery.js'

const today = new Date(2026, 10, 13)

test('formatCalendarMonthQuery pads the month and uses a 1-based month', () => {
  assert.equal(formatCalendarMonthQuery(0, 2026), '2026-01')
  assert.equal(formatCalendarMonthQuery(11, 2025), '2025-12')
})

test('parseCalendarMonthQuery reads a valid month', () => {
  assert.deepEqual(parseCalendarMonthQuery('2026-08', today), { monthIndex: 7, year: 2026 })
})

test('parseCalendarMonthQuery round-trips with formatCalendarMonthQuery', () => {
  const { monthIndex, year } = parseCalendarMonthQuery(formatCalendarMonthQuery(2, 2024), today)
  assert.deepEqual({ monthIndex, year }, { monthIndex: 2, year: 2024 })
})

for (const value of [undefined, null, '', ['2026-08'], '2026-8', '2026-13', '2026-00', 'august', '2026-08-01']) {
  test(`parseCalendarMonthQuery falls back to the current month for ${JSON.stringify(value)}`, () => {
    assert.deepEqual(parseCalendarMonthQuery(value, today), { monthIndex: 10, year: 2026 })
  })
}
