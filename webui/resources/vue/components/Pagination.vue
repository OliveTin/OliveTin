<template>
  <div class="pagination">
    <div class="pagination-info">
      <span class="pagination-text">
        Showing {{ startItem + 1 }}-{{ endItem }} of {{ total }} {{ itemTitle }}
      </span>
    </div>
    
    <div class="pagination-controls">
      <button 
        class="pagination-btn"
        :disabled="currentPage === 1"
        @click="goToPage(currentPage - 1)"
        title="Previous page"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24">
          <path fill="currentColor" d="M15.41 7.41L14 6l-6 6l6 6l1.41-1.41L10.83 12z"/>
        </svg>
      </button>

      
      <div class="pagination-pages">
        <!-- First page -->
        <button 
          v-if="showFirstPage"
          class="pagination-btn"
          :class="{ active: currentPage === 1 }"
          @click="goToPage(1)"
        >
          1
        </button>
        
        <!-- Ellipsis after first page -->
        <span v-if="showFirstEllipsis" class="pagination-ellipsis">...</span>
        
        <!-- Page numbers around current page -->
        <button 
          v-for="page in visiblePages" 
          :key="page"
          class="pagination-btn"
          :class="{ active: currentPage === page }"
          @click="goToPage(page)"
        >
          {{ page }}
        </button>
        
        <!-- Ellipsis before last page -->
        <span v-if="showLastEllipsis" class="pagination-ellipsis">...</span>
        
        <!-- Last page -->
        <button 
          v-if="showLastPage"
          class="pagination-btn"
          :class="{ active: currentPage === totalPages }"
          @click="goToPage(totalPages)"
        >
          {{ totalPages }}
        </button>
      </div>
      
      <button 
        class="pagination-btn"
        :disabled="currentPage === totalPages"
        @click="goToPage(currentPage + 1)"
        title="Next page"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24">
          <path fill="currentColor" d="M8.59 16.59L10 18l6-6l-6-6L8.59 7.41L13.17 12z"/>
        </svg>
      </button>
    </div>
    
    <div class="pagination-size" v-if="canChangePageSize">
      <label for="page-size">Items per page:</label>
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
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'

const props = defineProps({
  pageSize: {
    type: Number,
    default: 25
  },
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
    default: false
  },
  itemTitle: {
    type: String,
    default: 'items'
  }
})

const emit = defineEmits(['page-change', 'page-size-change'])

const localPageSize = ref(props.pageSize)
const localCurrentPage = ref(props.currentPage)

// Computed properties
const totalPages = computed(() => Math.ceil(props.total / localPageSize.value))

const startItem = computed(() => (localCurrentPage.value - 1) * localPageSize.value)
const endItem = computed(() => Math.min(localCurrentPage.value * localPageSize.value, props.total))

// Pagination logic
const maxVisiblePages = 5
const visiblePages = computed(() => {
  const pages = []
  const halfVisible = Math.floor(maxVisiblePages / 2)
  
  let start = Math.max(1, localCurrentPage.value - halfVisible)
  let end = Math.min(totalPages.value, start + maxVisiblePages - 1)
  
  // Adjust start if we're near the end
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

// Methods
function goToPage(page) {
  if (page >= 1 && page <= totalPages.value && page !== localCurrentPage.value) {
    localCurrentPage.value = page
    emit('page-change', page)
  }
}

function handlePageSizeChange() {
  // Reset to first page when changing page size
  localCurrentPage.value = 1
  emit('page-size-change', localPageSize.value)
  emit('page-change', 1)
}

// Watch for prop changes
watch(() => props.currentPage, (newPage) => {
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

.pagination-pages {
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.pagination-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 2.5rem;
  height: 2.5rem;
  padding: 0.5rem;
  border: 1px solid #dee2e6;
  background: #fff;
  color: #495057;
  text-decoration: none;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s ease;
  font-size: 0.875rem;
}

.pagination-btn:hover:not(:disabled) {
  background: #e9ecef;
  border-color: #adb5bd;
  color: #495057;
}

.pagination-btn.active {
  background: #c6d0d7;
  color: #333;
}

.pagination-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.pagination-ellipsis {
  padding: 0.5rem;
  color: #6c757d;
  font-size: 0.875rem;
}

.pagination-size {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.875rem;
  color: #6c757d;
}

.page-size-select {
  padding: 0.25rem 0.5rem;
  border: 1px solid #dee2e6;
  border-radius: 4px;
  background: #fff;
  font-size: 0.875rem;
}

.page-size-select:focus {
  outline: none;
  border-color: #5681af;
  box-shadow: 0 0 0 2px rgba(0, 123, 255, 0.25);
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