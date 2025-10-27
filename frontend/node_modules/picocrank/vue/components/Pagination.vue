<template>
  <div class="pagination">
    <div class="pagination-info">
      <span class="pagination-text">
      Page size:
      <select 
        id="page-size" 
        v-model="localPageSize" 
        @change="handlePageSizeChange"
        class="page-size-select"
      >
        <option value="10">10</option>
        <option value="25">25</option>
        <option value="50">50</option>
        <option value="100">100</option>
      </select>

      Showing {{ startItem + 1 }}-{{ endItem }} of {{ total }} {{ itemTitle }}
      </span>
    </div>
    
    <div class="pagination-controls">
      <button 
        class="button"
        :disabled="currentPageValue === 1"
        @click="goToPage(currentPageValue - 1)"
        title="Previous page"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24">
          <path fill="currentColor" d="M15.41 7.41L14 6l-6 6l6 6l1.41-1.41L10.83 12z"/>
        </svg>
      </button>

      
        <button 
          v-if="showFirstPage"
          class="button"
          :class="{ active: currentPageValue === 1 }"
          @click="goToPage(1)"
        >
          1
        </button>
        
        <span v-if="showFirstEllipsis" class="pagination-ellipsis">...</span>
        
        <button 
          v-for="page in visiblePages" 
          :key="page"
          class="button"
          :class="{ active: currentPageValue === page }"
          @click="goToPage(page)"
        >
          {{ page }}
        </button>
        
        <span v-if="showLastEllipsis" class="pagination-ellipsis">...</span>
        
        <button 
          v-if="showLastPage"
          class="button"
          :class="{ active: currentPageValue === totalPages }"
          @click="goToPage(totalPages)"
        >
        {{ totalPages }}
      </button>
      
      <button 
        class="button"
        :disabled="currentPageValue === totalPages"
        @click="goToPage(currentPageValue + 1)"
        title="Next page"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24">
          <path fill="currentColor" d="M8.59 16.59L10 18l6-6l-6-6L8.59 7.41L13.17 12z"/>
        </svg>
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'

const props = defineProps({
  total: {
    type: Number,
    required: true
  },
  currentPage: {
    type: Number,
    default: 1
  },
  canChangePageSize: {
    type: Boolean,
    default: true,
  },
  itemTitle: {
    type: String,
    default: 'items'
  },
  // Support for v-model binding
  page: {
    type: Number,
    default: 1
  },
  pageSize: {
    type: Number,
    default: 25
  }
})

const emit = defineEmits(['page-change', 'page-size-change', 'update:page', 'update:pageSize'])

const localPageSize = ref(props.pageSize)
const localCurrentPage = ref(props.currentPage)

// Computed property to get the current page value (supports both v-model and regular props)
const currentPageValue = computed(() => {
  // When using v-model, the page prop will be reactive and change
  // When using regular props, currentPage will be reactive
  return props.page
})

const totalPages = computed(() => Math.ceil(props.total / localPageSize.value))

const startItem = computed(() => (currentPageValue.value - 1) * localPageSize.value)
const endItem = computed(() => Math.min(currentPageValue.value * localPageSize.value, props.total))

const maxVisiblePages = 5
const visiblePages = computed(() => {
  const pages = []
  const halfVisible = Math.floor(maxVisiblePages / 2)
  
  let start = Math.max(1, currentPageValue.value - halfVisible)
  let end = Math.min(totalPages.value, start + maxVisiblePages - 1)
  
  if (end - start < maxVisiblePages - 1) {
    start = Math.max(1, end - maxVisiblePages + 1)
  }
  
  for (let i = start; i <= end; i++) {
    pages.push(i)
  }
  
  return pages
})

const showFirstPage = computed(() => visiblePages.value[0] > 1)
const showLastPage = computed(() => visiblePages.value[visiblePages.value.length - 1] < totalPages.value)
const showFirstEllipsis = computed(() => visiblePages.value[0] > 2)
const showLastEllipsis = computed(() => visiblePages.value[visiblePages.value.length - 1] < totalPages.value - 1)

function goToPage(page) {
  if (page >= 1 && page <= totalPages.value && page !== currentPageValue.value) {
    localCurrentPage.value = page
    emit('page-change', page)
    emit('update:page', page)
  }
}

function handlePageSizeChange() {
  localCurrentPage.value = 1
  emit('page-size-change', localPageSize.value)
  emit('update:pageSize', localPageSize.value)
  emit('page-change', 1)
  emit('update:page', 1)
}

watch(() => props.currentPage, (newPage) => {
  localCurrentPage.value = newPage
})

watch(() => props.page, (newPage) => {
  localCurrentPage.value = newPage
})

watch(() => props.pageSize, (newSize) => {
  localPageSize.value = newSize
})
</script>

<style scoped>
.pagination {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 1rem;
}

.pagination-info {
  flex: 1;
}

.pagination-text {
  font-size: 0.875rem;
  color: #6c757d;
}

.pagination-controls {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.button {
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 2.5rem;
  height: 2.5rem;
}

.button:disabled {
  opacity: 0.5;
  background: transparent;
  cursor: not-allowed;
}

.button.active {
  background: #545f69;
  color: #fff;
}

.pagination-ellipsis {
  padding: 0.5rem;
  color: #6c757d;
  font-size: 0.875rem;
}

#page-size {
  background: transparent;
  margin-left: 0.5rem;
  margin-right: 0.5rem;
}

option {
  background: #545f69;
}

/* Responsive design */
@media (max-width: 768px) {
  .pagination {
    flex-direction: column;
    gap: 1rem;
    align-items: stretch;
  }
  
  .pagination-controls {
    justify-content: center;
  }
  
  .pagination-size {
    justify-content: center;
  }
}
</style> 
