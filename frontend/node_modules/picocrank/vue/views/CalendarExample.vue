<template>
	<Section 
		title="Example Calendar" 
		padding
	>
        <template #toolbar>
            <button @click="goToToday">Today</button>
            <button @click="previousMonth">‹</button>
            <div class="date-picker-container">
                <input 
                    id="goto-date"
                    type="month" 
                    v-model="selectedDate"
                    @change="goToSelectedDate"
                    class="date-picker-input"
                />
            </div>
            <button @click="nextMonth">›</button>
        </template>
		<Calendar 
            :section="true"
			:events="events"
			:loading="loading"
			:error="error"
			:show-navigation="false"
			:current-month="currentMonthIndex"
			:current-year="currentYear"
			@event-click="handleEventClick"
			@date-click="handleDateClick"
			@month-change="handleMonthChange"
		/>
	</Section>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import Calendar from '../components/Calendar.vue'
import Section from '../components/Section.vue'

const events = ref([])
const loading = ref(false)
const error = ref(null)
const currentMonth = ref('?')
const currentMonthIndex = ref(new Date().getMonth())
const currentYear = ref(new Date().getFullYear())

// Date picker state - using native HTML5 month input
const selectedDate = ref('')

// Generate dummy events for the current month
function generateDummyEvents() {
	const now = new Date()
	const currentMonth = now.getMonth()
	const currentYear = now.getFullYear()
	
	// Get the number of days in the current month
	const daysInMonth = new Date(currentYear, currentMonth + 1, 0).getDate()
	
	const dummyEvents = []
	
	// Add some recurring events
	for (let day = 1; day <= daysInMonth; day++) {
		const date = new Date(currentYear, currentMonth, day)
		const dayOfWeek = date.getDay()
		
		// Add weekend events (Saturday and Sunday)
		if (dayOfWeek === 0 || dayOfWeek === 6) {
			dummyEvents.push({
				id: `weekend-${day}`,
				title: 'Weekend Activity',
				date: new Date(currentYear, currentMonth, day, 10, 0),
				description: 'Relaxing weekend activity'
			})
		}
		
		// Add some random events
		if (Math.random() > 0.7) {
			const eventTypes = [
				{ title: 'Team Meeting', time: [9, 0], duration: 1 },
				{ title: 'Client Call', time: [14, 30], duration: 0.5 },
				{ title: 'Project Review', time: [16, 0], duration: 2 },
				{ title: 'Lunch Break', time: [12, 0], duration: 1 },
				{ title: 'Code Review', time: [10, 0], duration: 1.5 }
			]
			
			const eventType = eventTypes[Math.floor(Math.random() * eventTypes.length)]
			const [hours, minutes] = eventType.time
			
			dummyEvents.push({
				id: `event-${day}-${Math.random().toString(36).substr(2, 9)}`,
				title: eventType.title,
				date: new Date(currentYear, currentMonth, day, hours, minutes),
				description: `Scheduled ${eventType.title.toLowerCase()}`
			})
		}
	}
	
	// Add some multi-day events
	dummyEvents.push({
		id: 'conference-1',
		title: 'Tech Conference',
		startDate: new Date(currentYear, currentMonth, 5, 9, 0),
		endDate: new Date(currentYear, currentMonth, 7, 17, 0),
		description: 'Annual technology conference'
	})
	
	dummyEvents.push({
		id: 'vacation-1',
		title: 'Vacation',
		startDate: new Date(currentYear, currentMonth, 15, 0, 0),
		endDate: new Date(currentYear, currentMonth, 18, 23, 59),
		description: 'Family vacation time'
	})
	
	// Add some all-day events
	dummyEvents.push({
		id: 'holiday-1',
		title: 'Public Holiday',
		date: new Date(currentYear, currentMonth, 1, 0, 0),
		description: 'National holiday - office closed'
	})
	
	if (daysInMonth >= 25) {
		dummyEvents.push({
			id: 'deadline-1',
			title: 'Project Deadline',
			date: new Date(currentYear, currentMonth, 25, 0, 0),
			description: 'Important project deadline'
		})
	}
	
	return dummyEvents
}

function handleEventClick(event) {
	console.log('Event clicked:', event)
	alert(`Event: ${event.title}\nDescription: ${event.description || 'No description'}`)
}

function handleDateClick(date) {
	console.log('Date clicked:', date)
	alert(`Date selected: ${date.toLocaleDateString()}`)
}

// Helper function to update the date picker value
function updateDatePicker(monthIndex, year) {
	const month = String(monthIndex + 1).padStart(2, '0')
	selectedDate.value = `${year}-${month}`
}

// Navigation functions - using the same fast path as internal calendar navigation
function previousMonth() {
	const newMonth = currentMonthIndex.value === 0 ? 11 : currentMonthIndex.value - 1
	const newYear = currentMonthIndex.value === 0 ? currentYear.value - 1 : currentYear.value
	
	// Update state directly (fast path)
	currentMonthIndex.value = newMonth
	currentYear.value = newYear
	
	// Update title
	const date = new Date(newYear, newMonth, 1)
	currentMonth.value = date.toLocaleDateString('en-US', { 
		month: 'long', 
		year: 'numeric' 
	})
	
	// Update date picker
	updateDatePicker(newMonth, newYear)
}

function nextMonth() {
	const newMonth = currentMonthIndex.value === 11 ? 0 : currentMonthIndex.value + 1
	const newYear = currentMonthIndex.value === 11 ? currentYear.value + 1 : currentYear.value
	
	// Update state directly (fast path)
	currentMonthIndex.value = newMonth
	currentYear.value = newYear
	
	// Update title
	const date = new Date(newYear, newMonth, 1)
	currentMonth.value = date.toLocaleDateString('en-US', { 
		month: 'long', 
		year: 'numeric' 
	})
	
	// Update date picker
	updateDatePicker(newMonth, newYear)
}

function goToToday() {
	const now = new Date()
	const newMonth = now.getMonth()
	const newYear = now.getFullYear()
	
	// Update state directly (fast path)
	currentMonthIndex.value = newMonth
	currentYear.value = newYear
	
	// Update title
	currentMonth.value = now.toLocaleDateString('en-US', { 
		month: 'long', 
		year: 'numeric' 
	})
	
	// Update date picker
	updateDatePicker(newMonth, newYear)
}

function goToSelectedDate() {
	if (!selectedDate.value) return
	
	// Parse the YYYY-MM format from the native date picker
	const [year, month] = selectedDate.value.split('-').map(Number)
	const monthIndex = month - 1 // Convert to 0-based month index
	
	// Update state directly (fast path)
	currentMonthIndex.value = monthIndex
	currentYear.value = year
	
	// Update title
	const date = new Date(year, monthIndex, 1)
	currentMonth.value = date.toLocaleDateString('en-US', { 
		month: 'long', 
		year: 'numeric' 
	})
}

function handleMonthChange(month, year) {
	console.log('Month changed:', month, year)
	// In a real app, you might fetch events for the new month
	loading.value = true
	
	// Update current month and year state
	currentMonthIndex.value = month
	currentYear.value = year
	
	// Use native JavaScript to format month and year
	const date = new Date(year, month, 1)
	currentMonth.value = date.toLocaleDateString('en-US', { 
		month: 'long', 
		year: 'numeric' 
	})
	
	// Update date picker to reflect the new month
	updateDatePicker(month, year)
	
	setTimeout(() => {
		// Simulate loading delay
		events.value = generateDummyEvents()
		loading.value = false
	}, 500)
}

onMounted(() => {
	// Generate initial events for current month
	events.value = generateDummyEvents()
	
	// Set initial month title using native JavaScript
	const now = new Date()
	currentMonthIndex.value = now.getMonth()
	currentYear.value = now.getFullYear()
	currentMonth.value = now.toLocaleDateString('en-US', { 
		month: 'long', 
		year: 'numeric' 
	})
	
	// Initialize the native date picker with current month/year
	const year = now.getFullYear()
	const month = String(now.getMonth() + 1).padStart(2, '0')
	selectedDate.value = `${year}-${month}`
})
</script>

<style scoped>
.date-picker-container {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.date-picker-label {
  font-size: 0.875rem;
  font-weight: 500;
  color: #374151;
}

.date-picker-input {
  padding: 0.5rem;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 0.875rem;
  background: white;
  cursor: pointer;
  transition: all 0.2s;
}

.date-picker-input:hover {
  border-color: #9ca3af;
}

.date-picker-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

/* Dark theme support */
@media (prefers-color-scheme: dark) {
  .date-picker-label {
    color: #d1d5db;
  }
  
  .date-picker-input {
    background: #374151;
    border-color: #4b5563;
    color: #f9fafb;
  }
  
  .date-picker-input:hover {
    border-color: #6b7280;
  }
  
  .date-picker-input:focus {
    border-color: #60a5fa;
    box-shadow: 0 0 0 3px rgba(96, 165, 250, 0.1);
  }
}
</style>
