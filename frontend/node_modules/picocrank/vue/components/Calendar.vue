<script setup lang="ts">
import { ref, computed } from 'vue'

export interface CalendarEvent {
  id: string | number
  title: string
  startDate?: Date | string | null
  endDate?: Date | string | null
  date?: Date | string | null
  [key: string]: any
}

export interface CalendarProps {
  events: CalendarEvent[]
  monthNames?: string[]
  dayNames?: string[]
  loading?: boolean
  error?: string | null
  // Formatter functions
  getEventDate?: (event: CalendarEvent) => Date | null
  getEventDateRange?: (event: CalendarEvent) => { start: Date | null; end: Date | null }
  formatEventTime?: (event: CalendarEvent, date: Date) => string
  // Customization
  showNavigation?: boolean
  currentMonth?: number
  currentYear?: number
}

const props = withDefaults(defineProps<CalendarProps>(), {
  section: false,
  monthNames: () => [
    'January', 'February', 'March', 'April', 'May', 'June',
    'July', 'August', 'September', 'October', 'November', 'December'
  ],
  dayNames: () => ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'],
  loading: false,
  error: null,
  showNavigation: true,
})

const emit = defineEmits<{
  'event-click': [event: CalendarEvent]
  'date-click': [date: Date]
  'month-change': [month: number, year: number]
  'event-context-menu': [event: CalendarEvent, mouseEvent: MouseEvent]
}>()

// Internal calendar state
const internalCurrentDate = ref(new Date())
const currentDate = computed(() => {
  if (props.currentMonth !== undefined && props.currentYear !== undefined) {
    return new Date(props.currentYear, props.currentMonth, 1)
  }
  return internalCurrentDate.value
})

const currentMonth = computed(() => currentDate.value.getMonth())
const currentYear = computed(() => currentDate.value.getFullYear())

// Helper function to check if a time is midnight (00:00)
function isMidnight(dateValue: any): boolean {
  if (!dateValue) return false
  const date = new Date(dateValue)
  return date.getHours() === 0 && date.getMinutes() === 0
}

// Helper function to format time, returning "No time" if midnight
function formatTimeOrNoTime(dateValue: any): string {
  if (!dateValue) return 'No time'
  if (isMidnight(dateValue)) return 'No time'
  const date = new Date(dateValue)
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

// Get event date range
function getEventDateRange(event: CalendarEvent): { start: Date | null; end: Date | null } {
  if (props.getEventDateRange) {
    return props.getEventDateRange(event)
  }
  
  // Default implementation
  let start: Date | null = null
  let end: Date | null = null
  
  if (event.startDate) {
    start = new Date(event.startDate)
  } else if (event.date) {
    start = new Date(event.date)
  }
  
  if (event.endDate) {
    end = new Date(event.endDate)
  }
  
  return { start, end }
}

// Get events for a specific date
function getEventsForDate(date: Date): CalendarEvent[] {
  const targetDate = new Date(date)
  targetDate.setHours(0, 0, 0, 0)

  return props.events.filter(event => {
    const { start, end } = getEventDateRange(event)
    
    if (!start) return false
    
    const startDateOnly = new Date(start)
    startDateOnly.setHours(0, 0, 0, 0)
    
    if (end) {
      const endDateOnly = new Date(end)
      endDateOnly.setHours(0, 0, 0, 0)
      return targetDate >= startDateOnly && targetDate <= endDateOnly
    } else {
      return targetDate.getTime() === startDateOnly.getTime()
    }
  })
}

// Check if an event is a multi-day event
function isMultiDayEvent(event: CalendarEvent): boolean {
  const { start, end } = getEventDateRange(event)
  
  if (!start || !end) return false

  const startDate = new Date(start)
  const endDate = new Date(end)

  startDate.setHours(0, 0, 0, 0)
  endDate.setHours(0, 0, 0, 0)

  return startDate.getTime() !== endDate.getTime()
}

// Get the position of an event within a multi-day range for a specific date
function getMultiDayPosition(event: CalendarEvent, date: Date): 'start' | 'middle' | 'end' | 'single' {
  const { start, end } = getEventDateRange(event)
  
  if (!start || !end) return 'single'

  const startDate = new Date(start)
  const endDate = new Date(end)
  const targetDate = new Date(date)

  startDate.setHours(0, 0, 0, 0)
  endDate.setHours(0, 0, 0, 0)
  targetDate.setHours(0, 0, 0, 0)

  if (startDate.getTime() === endDate.getTime()) return 'single'
  if (targetDate.getTime() === startDate.getTime()) return 'start'
  if (targetDate.getTime() === endDate.getTime()) return 'end'
  if (targetDate > startDate && targetDate < endDate) return 'middle'

  return 'single'
}

// Format event time based on position in multi-day event
function formatEventTimeDefault(event: CalendarEvent, date: Date): string {
  if (props.formatEventTime) {
    return props.formatEventTime(event, date)
  }
  
  const { start, end } = getEventDateRange(event)
  const position = getMultiDayPosition(event, date)

  if (!start || !end) return 'No time'

  if (position === 'start') {
    return formatTimeOrNoTime(start)
  } else if (position === 'end') {
    return formatTimeOrNoTime(end)
  } else if (position === 'middle') {
    return 'All day'
  } else if (position === 'single') {
    return formatTimeOrNoTime(start)
  }

  return 'All day'
}

// Calendar generation
const calendarDays = computed(() => {
  const year = currentYear.value
  const month = currentMonth.value

  const firstDay = new Date(year, month, 1)
  const lastDay = new Date(year, month + 1, 0)
  const daysInMonth = lastDay.getDate()

  // Get starting day of week (Monday as first day)
  const startDay = (firstDay.getDay() + 6) % 7

  const days = []

  // Add days from previous month to fill the first row
  const prevMonth = month === 0 ? 11 : month - 1
  const prevYear = month === 0 ? year - 1 : year
  const prevMonthLastDay = new Date(prevYear, prevMonth + 1, 0).getDate()

  for (let i = 0; i < startDay; i++) {
    const day = prevMonthLastDay - startDay + i + 1
    const date = new Date(prevYear, prevMonth, day)
    days.push({
      date,
      events: getEventsForDate(date)
    })
  }

  // Add days of the month
  for (let day = 1; day <= daysInMonth; day++) {
    const date = new Date(year, month, day)
    days.push({
      date,
      events: getEventsForDate(date)
    })
  }

  // Add days from next month to complete the last row
  const totalCells = 42
  const remainingCells = totalCells - days.length

  for (let day = 1; day <= remainingCells; day++) {
    const date = new Date(year, month + 1, day)
    days.push({
      date,
      events: getEventsForDate(date)
    })
  }

  return days
})

// Navigation functions
function previousMonth() {
  if (props.currentMonth !== undefined && props.currentYear !== undefined) {
    const newMonth = currentMonth.value === 0 ? 11 : currentMonth.value - 1
    const newYear = currentMonth.value === 0 ? currentYear.value - 1 : currentYear.value
    emit('month-change', newMonth, newYear)
  } else {
    internalCurrentDate.value = new Date(currentYear.value, currentMonth.value - 1, 1)
  }
}

function nextMonth() {
  if (props.currentMonth !== undefined && props.currentYear !== undefined) {
    const newMonth = currentMonth.value === 11 ? 0 : currentMonth.value + 1
    const newYear = currentMonth.value === 11 ? currentYear.value + 1 : currentYear.value
    emit('month-change', newMonth, newYear)
  } else {
    internalCurrentDate.value = new Date(currentYear.value, currentMonth.value + 1, 1)
  }
}

function goToToday() {
  if (props.currentMonth !== undefined && props.currentYear !== undefined) {
    const today = new Date()
    emit('month-change', today.getMonth(), today.getFullYear())
  } else {
    internalCurrentDate.value = new Date()
  }
}

// Event handlers
function handleEventClick(event: CalendarEvent) {
  emit('event-click', event)
}

function handleDateClick(date: Date) {
  emit('date-click', date)
}

function handleContextMenu(event: CalendarEvent, mouseEvent: MouseEvent) {
  mouseEvent.preventDefault()
  mouseEvent.stopPropagation()
  emit('event-context-menu', event, mouseEvent)
}
</script>

<template>
  <div class="calendar-wrapper">
    <div v-if="showNavigation" class="calendar-header-nav">
      <h2 class="calendar-title">{{ monthNames[currentMonth] }} {{ currentYear }}</h2>
      <div class="calendar-nav-buttons">
        <slot name="nav-buttons">
          <button @click="previousMonth" class="button neutral">‹</button>
          <button @click="goToToday" class="button neutral">Today</button>
          <button @click="nextMonth" class="button neutral">›</button>
        </slot>
      </div>
    </div>

    <div v-if="error" class="calendar-error">{{ error }}</div>
    <div v-else-if="loading" class="calendar-loading">Loading…</div>
    <div v-else class="calendar-container">
      <div class="calendar-grid">
        <!-- Day headers -->
        <div v-for="day in dayNames" :key="day" class="day-header">{{ day }}</div>

        <!-- Calendar days -->
        <div
          v-for="(day, index) in calendarDays"
          :key="index"
          class="calendar-day"
          :class="{
            'today': day && day.date.toDateString() === new Date().toDateString(),
            'weekend': day && (day.date.getDay() === 0 || day.date.getDay() === 6),
            'prev-month': day && day.date.getMonth() !== currentMonth && day.date.getMonth() !== (currentMonth + 1) % 12,
            'next-month': day && day.date.getMonth() !== currentMonth && day.date.getMonth() === (currentMonth + 1) % 12
          }"
        >
          <div v-if="day" class="day-content" @click="handleDateClick(day.date)">
            <div class="day-number clickable">
              {{ day.date.getDate() }}
            </div>
            <div class="day-events">
              <div
                v-for="event in day.events.sort((a, b) => {
                  const aMultiDay = isMultiDayEvent(a);
                  const bMultiDay = isMultiDayEvent(b);
                  if (aMultiDay && !bMultiDay) return -1;
                  if (!aMultiDay && bMultiDay) return 1;
                  return 0;
                }).slice(0, 3)"
                :key="event.id"
                class="calendar-event"
                :class="{
                  'multi-day': isMultiDayEvent(event),
                  'multi-day-start': isMultiDayEvent(event) && getMultiDayPosition(event, day.date) === 'start',
                  'multi-day-middle': isMultiDayEvent(event) && getMultiDayPosition(event, day.date) === 'middle',
                  'multi-day-end': isMultiDayEvent(event) && getMultiDayPosition(event, day.date) === 'end'
                }"
                @click.stop="handleEventClick(event)"
                @contextmenu.stop="handleContextMenu(event, $event)"
              >
                <slot name="event" :event="event" :date="day.date" :position="getMultiDayPosition(event, day.date)">
                  <div class="event-content">
                    <div class="event-title">
                      {{ event.title }}
                      <span v-if="isMultiDayEvent(event)" class="multi-day-indicator">
                        {{ getMultiDayPosition(event, day.date) === 'start' ? '▶' :
                           getMultiDayPosition(event, day.date) === 'end' ? '◀' :
                           getMultiDayPosition(event, day.date) === 'middle' ? '▬' : '' }}
                      </span>
                    </div>
                    <div class="event-time">
                      {{ formatEventTimeDefault(event, day.date) }}
                    </div>
                  </div>
                </slot>
              </div>
              <div v-if="day.events.length > 3" class="more-events">
                +{{ day.events.length - 3 }} more
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.calendar-wrapper {
  width: 100%;
}

.calendar-header-nav {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
  padding: 1rem;
  background: #f8f9fa;
  border-radius: 8px;
}

.calendar-title {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 600;
}

.calendar-nav-buttons {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.calendar-error {
  color: #b00020;
  padding: 1rem;
}

.calendar-loading {
  padding: 1rem;
  text-align: center;
}

.calendar-container {
  background: white;
  border: 1px solid #e0e0e0;
  overflow: hidden;
  border-radius: 8px;
}

.calendar-grid {
  display: grid;
  grid-template-columns: repeat(7, minmax(0, 1fr));
  min-height: 400px;
}

.day-header {
  padding: 1rem;
  text-align: center;
  font-weight: 600;
  color: #666;
  border-right: 1px solid #e0e0e0;
  border-bottom: 1px solid #e0e0e0;
  background: #f8f9fa;
}

.day-header:last-child {
  border-right: none;
}

.calendar-day {
  border-right: 1px solid #e0e0e0;
  border-bottom: 1px solid #e0e0e0;
  min-height: 120px;
  position: relative;
  transition: background-color 0.2s ease;
  cursor: pointer;
}

.calendar-day:hover {
  background-color: #f0f8ff;
}

.calendar-day:nth-child(7n) {
  border-right: none;
}

.calendar-day.weekend {
  background: #f8f9fa;
}

.calendar-day.weekend:hover {
  background: #e9ecef;
}

.calendar-day.next-month,
.calendar-day.prev-month {
  background: #f8f9fa;
  opacity: 0.6;
}

.calendar-day.next-month:hover,
.calendar-day.prev-month:hover {
  background: #e9ecef;
  opacity: 0.8;
}

.calendar-day.today {
  background: #f7f8d7;
  font-weight: bold;
}

.calendar-day.today:hover {
  background: #bbdefb;
}


.day-content {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.day-number {
  font-size: 0.9rem;
  margin-bottom: 0.25rem;
  color: #333;
  text-decoration: none;
  display: inline-block;
  padding: 0.1rem;
  border-radius: 4px;
  transition: all 0.2s;
  min-width: 1.5rem;
  text-align: center;
}

.day-number.clickable {
  cursor: pointer;
  padding: 0.25rem;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.day-content:hover .day-number.clickable {
  color: #007bff;
}

.day-events {
  flex: 1;
  overflow: hidden;
}

.calendar-event {
  background: #d6f1aa;
  border: 1px solid #c4db96;
  margin-bottom: 0.25rem;
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
  padding: 0.1rem 0.1rem;
}

.calendar-event:hover {
  background: #b3de6e;
  border-color: #c4db96;
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.15);
}

.calendar-event.multi-day-start {
  border-top-left-radius: 4px;
  border-bottom-left-radius: 4px;
}

.calendar-event.multi-day-middle {
  border-radius: 0;
}

.calendar-event.multi-day-end {
  border-top-right-radius: 4px;
  border-bottom-right-radius: 4px;
}

.event-content {
  cursor: pointer;
}

.event-title {
  font-weight: bold;
  font-size: 0.85rem;
  color: #333;
  white-space: normal;
  overflow: hidden;
  text-overflow: ellipsis;
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.multi-day-indicator {
  font-size: 0.7rem;
  color: #007bff;
  font-weight: bold;
}

.event-time {
  font-size: 0.75rem;
  color: #666;
  margin-top: 0.125rem;
}

.more-events {
  font-size: 0.75rem;
  color: #666;
  text-align: center;
  padding: 0.25rem;
  background: #f8f9fa;
  border-radius: 4px;
  margin-top: 0.25rem;
}

/* Responsive design */
@media (max-width: 768px) {
  .calendar-header-nav {
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  .calendar-title {
    font-size: 1.2rem;
  }

  .calendar-day {
    min-height: 80px;
  }

  .day-number {
    font-size: 1rem;
  }

  .event-title {
    font-size: 0.8rem;
  }

  .event-time {
    font-size: 0.7rem;
  }
}

@media (max-width: 480px) {
  .calendar-grid {
    min-height: 300px;
  }

  .calendar-day {
    min-height: 60px;
  }

  .day-header {
    padding: 0.5rem 0.25rem;
    font-size: 0.8rem;
  }
}

@media (min-width: 768px) {
  .day-content {
    padding: 0.45rem;
  }

  .calendar-event {
    border-radius: 4px;
    padding: 0.25rem 0.5rem;
  }
}

@media (prefers-color-scheme: dark) {
  .calendar-container {
    background: #565656;
    border-color: #374151;
    border: 1px solid #565656;
  }

  .calendar-day.today {
    background: #646c70;
  }

  .calendar-day.weekend {
    background: #444;
  }

  .calendar-day.weekend:hover {
    background: #1a1a1a;
  }

  .calendar-day.next-month,
  .calendar-day.prev-month {
    background: #1a1a1a !important;
  }
  
  .calendar-day.next-month:hover,
  .calendar-day.prev-month:hover {
    background: #374151;
  }

  .calendar-day.today:hover {
    background: #374151;
  }

  .calendar-day {
    border: 1px solid #3b3b3b;
  }

  .day-number {
    color: #f9fafb;
  }

  .day-header {
    color: #f9fafb;
    background: #444 !important;
    border: 1px solid #374151;
    border-color: #374151;
  }

  .day-content:hover .day-number.clickable {
    color: #f9fafb;
  }

  .calendar-day:hover {
    background-color: #374151;
  }

  .calendar-day.today:hover {
    background: #374151;
  }

  .calendar-day.weekend:hover {
    background: #374151;
  }
}
</style>
