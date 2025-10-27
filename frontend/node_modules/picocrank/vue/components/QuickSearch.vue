<template>
  <div class="quick-search">
    <div class="search-container">
      <input
        ref="searchInput"
        v-model="searchQuery"
        type="text"
        class="search-input"
        :placeholder="placeholder"
        @input="onSearch"
        @focus="onFocus"
        @blur="onBlur"
        @keydown="onKeydown"
      />
      <div class="search-icon">
        <HugeiconsIcon :icon="SearchIcon" width="1em" height="1em" />
      </div>
    </div>
    
    <div v-if="showResults && filteredItems.length > 0" class="search-results">
      <div
        v-for="(item, index) in filteredItems"
        :key="item.id"
        :class="['search-result-item', { active: selectedIndex === index }]"
        @click="selectItem(item)"
        @mouseenter="selectedIndex = index"
      >
        <div class="result-content">
          <div class="result-title" v-html="highlightText(item.title, searchQuery)"></div>
          <div v-if="item.description" class="result-description" v-html="highlightText(item.description, searchQuery)"></div>
          <div v-if="item.category" class="result-category">{{ item.category }}</div>
        </div>
        <div v-if="item.icon" class="result-icon">
          <HugeiconsIcon :icon="item.icon" width="1.2em" height="1.2em" />
        </div>
      </div>
    </div>
    
    <div v-if="showResults && searchQuery && filteredItems.length === 0" class="no-results">
      <div class="no-results-content">
        <HugeiconsIcon :icon="SearchRemoveIcon" width="2em" height="2em" />
        <p>No results found for "{{ searchQuery }}"</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { HugeiconsIcon } from '@hugeicons/vue'
import { SearchIcon } from '@hugeicons/core-free-icons'
import { SearchRemoveIcon } from '@hugeicons/core-free-icons'
import { ViewIcon } from '@hugeicons/core-free-icons'

const props = defineProps({
  placeholder: {
    type: String,
    default: 'Search...'
  },
  items: {
    type: Array,
    default: () => []
  },
  searchFields: {
    type: Array,
    default: () => ['title', 'description']
  },
  maxResults: {
    type: Number,
    default: 10
  },
  debounceMs: {
    type: Number,
    default: 300
  },
  enableGlobalShortcut: {
    type: Boolean,
    default: true
  },
  autoImportRoutes: {
    type: Boolean,
    default: true
  }
})

const emit = defineEmits(['select', 'search', 'focus', 'blur'])

const router = useRouter()
const searchInput = ref(null)
const searchQuery = ref('')
const showResults = ref(false)
const selectedIndex = ref(-1)
const items = ref([...props.items])
const debounceTimer = ref(null)

function importRoutesFromRouter() {
  const routeItems = router.getRoutes()
    .filter(route => route.name) // Exclude unnamed routes
    .map(route => ({
      id: `route-${route.name}`,
      title: route.meta.title || route.name,
      description: `Navigate to ${route.path}`,
      category: 'Navigation',
      path: route.path,
      icon: route.meta.icon || ViewIcon,
      type: 'route'
    }))
  
  // Add the items to the search
  routeItems.forEach(item => addItem(item))
}

const filteredItems = computed(() => {
  if (!searchQuery.value.trim()) {
    return []
  }
  
  const query = searchQuery.value.toLowerCase()
  const filtered = items.value.filter(item => {
    return props.searchFields.some(field => {
      const value = item[field]
      return value && value.toLowerCase().includes(query)
    })
  })
  
  return filtered.slice(0, props.maxResults)
})

function handleGlobalKeydown(event) {
  // Check if Ctrl+K is pressed
  if (event.ctrlKey && event.key === 'k') {
    event.preventDefault()
    
    // Don't trigger if user is typing in an input/textarea
    if (event.target.tagName === 'INPUT' || event.target.tagName === 'TEXTAREA') {
      return
    }
    
    focus()
  }
}

function onSearch() {
  clearTimeout(debounceTimer.value)
  debounceTimer.value = setTimeout(() => {
    selectedIndex.value = -1
    showResults.value = true
    emit('search', searchQuery.value)
  }, props.debounceMs)
}

function onFocus() {
  showResults.value = true
  emit('focus')
}

function onBlur() {
  // Delay hiding results to allow for click events
  setTimeout(() => {
    showResults.value = false
    selectedIndex.value = -1
    emit('blur')
  }, 150)
}

function onKeydown(event) {
  if (!showResults.value || filteredItems.value.length === 0) return
  
  switch (event.key) {
    case 'ArrowDown':
      event.preventDefault()
      selectedIndex.value = Math.min(selectedIndex.value + 1, filteredItems.value.length - 1)
      break
    case 'ArrowUp':
      event.preventDefault()
      selectedIndex.value = Math.max(selectedIndex.value - 1, -1)
      break
    case 'Enter':
      event.preventDefault()
      if (selectedIndex.value >= 0) {
        selectItem(filteredItems.value[selectedIndex.value])
      }
      break
    case 'Escape':
      showResults.value = false
      selectedIndex.value = -1
      searchInput.value.blur()
      break
  }
}

function selectItem(item) {
  switch(item.type) {
    case 'route':
      router.push({ path: item.path })
      break
    case 'callback':
      item.callback()
      break
    default:
      emit('select', item)
  }
  
  searchQuery.value = ''
  showResults.value = false
  selectedIndex.value = -1
}

function highlightText(text, query) {
  if (!query || !text) return text
  
  const regex = new RegExp(`(${query.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')})`, 'gi')
  return text.replace(regex, '<mark>$1</mark>')
}

// API Methods
function addItem(item) {
  const existingIndex = items.value.findIndex(i => i.id === item.id)
  if (existingIndex >= 0) {
    items.value[existingIndex] = { ...item }
  } else {
    items.value.push({ ...item })
  }
}

function removeItem(itemId) {
  items.value = items.value.filter(item => item.id !== itemId)
}

function clearItems() {
  items.value = []
}

function getItems() {
  return [...items.value]
}

function setItems(newItems) {
  items.value = [...newItems]
}

function focus() {
  searchInput.value?.focus()
}

function blur() {
  searchInput.value?.blur()
}

function clear() {
  searchQuery.value = ''
  showResults.value = false
  selectedIndex.value = -1
}

function refreshRoutes() {
  if (props.autoImportRoutes) {
    importRoutesFromRouter()
  }
}

// Global shortcut management
onMounted(() => {
  if (props.enableGlobalShortcut) {
    document.addEventListener('keydown', handleGlobalKeydown)
  }
  
  // Auto-import routes on mount
  if (props.autoImportRoutes) {
    importRoutesFromRouter()
  }
})

onUnmounted(() => {
  if (props.enableGlobalShortcut) {
    document.removeEventListener('keydown', handleGlobalKeydown)
  }
})

// Watch for external items changes
watch(() => props.items, (newItems) => {
  items.value = [...newItems]
}, { deep: true })

defineExpose({
  addItem,
  removeItem,
  clearItems,
  getItems,
  setItems,
  focus,
  blur,
  clear,
  refreshRoutes,
  searchQuery,
  filteredItems,
  showResults
})
</script>

<style scoped>
.quick-search {
  position: relative;
  width: 100%;
  max-width: 400px;
}

.search-container {
  position: relative;
  display: flex;
  align-items: center;
}

.search-input {
  width: 100%;
  padding: 0.4em;
  border: none;
  border-radius: .4em;
  font-size: 1rem;
  outline: none;
  transition: all 0.2s ease;
  background-color: #666;
  color: #fff;
}

.search-input::placeholder {
  color: #bbb;
}

.search-input:focus {
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.search-icon {
  position: absolute;
  right: 0.75rem;
  color: #6b7280;
  pointer-events: none;
}

.search-results {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  background: #fff;
  border: 2px solid #e1e5e9;
  border-top: none;
  border-radius: 0 0 8px 8px;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
  max-height: 300px;
  overflow-y: auto;
  z-index: 1000;
}

.search-result-item {
  display: flex;
  align-items: center;
  padding: 0.75rem 1rem;
  cursor: pointer;
  transition: background-color 0.15s ease;
  border-bottom: 1px solid #f3f4f6;
}

.search-result-item:last-child {
  border-bottom: none;
}

.search-result-item:hover,
.search-result-item.active {
  background-color: #f8fafc;
}

.result-content {
  flex: 1;
  min-width: 0;
}

.result-title {
  font-weight: 500;
  color: #1f2937;
  margin-bottom: 0.25rem;
}

.result-description {
  font-size: 0.875rem;
  color: #6b7280;
  margin-bottom: 0.25rem;
}

.result-category {
  font-size: 0.75rem;
  color: #9ca3af;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.result-icon {
  margin-left: 0.75rem;
  color: #6b7280;
  flex-shrink: 0;
}

.no-results {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  background: #fff;
  border: 2px solid #e1e5e9;
  border-top: none;
  border-radius: 0 0 8px 8px;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
  z-index: 1000;
}

.no-results-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 2rem 1rem;
  color: #6b7280;
  text-align: center;
}

.no-results-content p {
  margin: 0.5rem 0 0 0;
  font-size: 0.875rem;
}

mark {
  background-color: #fef3c7;
  color: #92400e;
  padding: 0;
  border-radius: 2px;
}

/* Dark theme */
@media (prefers-color-scheme: dark) {
  .search-input {
    background-color: #1f2937;
    border-color: #374151;
    color: #f9fafb;
  }

  .search-input:focus {
    border-color: #60a5fa;
    box-shadow: 0 0 0 3px rgba(96, 165, 250, 0.1);
  }

  .search-icon {
    color: #9ca3af;
  }

  .search-results {
    background: #1f2937;
    border-color: #374151;
  }

  .search-result-item {
    border-bottom-color: #374151;
  }

  .search-result-item:hover,
  .search-result-item.active {
    background-color: #374151;
  }

  .result-title {
    color: #f9fafb;
  }

  .result-description {
    color: #d1d5db;
  }

  .result-category {
    color: #9ca3af;
  }

  .result-icon {
    color: #9ca3af;
  }

  .no-results {
    background: #1f2937;
    border-color: #374151;
  }

  .no-results-content {
    color: #9ca3af;
  }

  mark {
    background-color: #451a03;
    color: #fbbf24;
  }
}

/* Responsive */
@media (max-width: 640px) {
  .quick-search {
    max-width: 100%;
  }
  
  .search-results {
    max-height: 250px;
  }
}
</style>
