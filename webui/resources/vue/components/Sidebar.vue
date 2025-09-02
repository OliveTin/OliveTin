<template>
  <aside :class="{ 'shown': isOpen, 'stuck': isStuck }" class="sidebar">
    <div class = "flex-row">
      <h2>Navigation</h2>
      <div class = "fg1" />
      <button
        class="stick-toggle"
        :aria-pressed="isStuck"
        :title="isStuck ? 'Unstick sidebar' : 'Stick sidebar'"
        @click="toggleStick"
      >
        <span v-if="isStuck">
          <HugeiconsIcon :icon="Pin02Icon" width = "1em" height = "1em" />
        </span>
        <span v-else>
          <HugeiconsIcon :icon="PinIcon" width = "1em" height = "1em" />
        </span>
      </button>
    </div>

    <nav class="mainnav">
      <ul class="navigation-links">
        <li v-for="link in navigationLinks" :key="link.id" :title="link.title">
          <router-link :to="link.path" :class="{ active: isActive(link.path) }">
            <HugeiconsIcon :icon="link.icon" />
            <span>{{ link.title }}</span>
          </router-link>
        </li>
      </ul>

      <ul class="supplemental-links">
        <li v-for="link in supplementalLinks" :key="link.id" :title="link.title">
          <router-link :to="link.path" :class="{ active: isActive(link.path) }">
            <HugeiconsIcon :icon="link.icon" />
            <span>{{ link.title }}</span>
          </router-link>
        </li>
      </ul>
    </nav>
  </aside>
</template>

<script setup>
import { ref, onMounted, getCurrentInstance } from 'vue'
import { useRoute } from 'vue-router'
import { HugeiconsIcon } from '@hugeicons/vue'
import { DashboardSquare01Icon } from '@hugeicons/core-free-icons'
import { LeftToRightListDashIcon } from '@hugeicons/core-free-icons'
import { Wrench01Icon } from '@hugeicons/core-free-icons'
import { Pin02Icon } from '@hugeicons/core-free-icons'
import { PinIcon } from '@hugeicons/core-free-icons'
import { CellsIcon } from '@hugeicons/core-free-icons'

const isOpen = ref(false)
const isStuck = ref(false)
const navigationLinks = ref([
  {
    id: 'actions',
    title: 'Actions',
    path: '/',
    icon: DashboardSquare01Icon,
  }
])

const supplementalLinks = ref([
  {
    id: 'entities',
	title: 'Entities',
	path: '/entities',
	icon: CellsIcon,
  },
  {
    id: 'logs',
    title: 'Logs',
    path: '/logs',
    icon: LeftToRightListDashIcon,
  },
  {
    id: 'diagnostics',
    title: 'Diagnostics',
    path: '/diagnostics',
    icon: Wrench01Icon,
  }
])

const route = useRoute()

function toggleStick() {
  isStuck.value = !isStuck.value
}

function toggle() {
  isOpen.value = !isOpen.value
  isStuck.value = false
}

function open() {
  isOpen.value = true
}

function close() {
  isOpen.value = false
  isStuck.value = false
}

function isActive(path) {
  return route.path === path
}

// Method to add navigation links from other components
function addNavigationLink(link) {
  link.icon = DashboardSquare01Icon

  const existingIndex = navigationLinks.value.findIndex(l => l.id === link.id)
  if (existingIndex >= 0) {
    navigationLinks.value[existingIndex] = { ...link }
  } else {
    navigationLinks.value.push({ ...link })
  }
}

// Method to add supplemental links from other components
function addSupplementalLink(link) {
  const existingIndex = supplementalLinks.value.findIndex(l => l.id === link.id)
  if (existingIndex >= 0) {
    supplementalLinks.value[existingIndex] = { ...link }
  } else {
    supplementalLinks.value.push({ ...link })
  }
}

// Method to remove links
function removeNavigationLink(linkId) {
  navigationLinks.value = navigationLinks.value.filter(link => link.id !== linkId)
}

function removeSupplementalLink(linkId) {
  supplementalLinks.value = supplementalLinks.value.filter(link => link.id !== linkId)
}

// Method to clear all links
function clearNavigationLinks() {
  navigationLinks.value = []
}

function clearSupplementalLinks() {
  supplementalLinks.value = []
}

// Method to get all links
function getNavigationLinks() {
  return [...navigationLinks.value]
}

function getSupplementalLinks() {
  return [...supplementalLinks.value]
}

onMounted(() => {
  // Make the sidebar globally accessible
  window.sidebar = {
    get isOpen() { return isOpen.value },
    set isOpen(val) { isOpen.value = val },
    toggle,
    open,
    close,
    addNavigationLink,
    addSupplementalLink,
    removeNavigationLink,
    removeSupplementalLink,
    clearNavigationLinks,
    clearSupplementalLinks,
    getNavigationLinks,
    getSupplementalLinks
  }
})

defineExpose({
  isOpen,
  navigationLinks,
  supplementalLinks,
  toggle,
  open,
  close,
  isActive,
  addNavigationLink,
  addSupplementalLink,
  removeNavigationLink,
  removeSupplementalLink,
  clearNavigationLinks,
  clearSupplementalLinks,
  getNavigationLinks,
  getSupplementalLinks
})
</script>

<style scoped>

.active {
  text-decoration: underline;
}

li {
  margin: 0;
  padding: 0;
}
.navigation-links a,
.supplemental-links a {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  color: #333;
  transition: background-color 0.2s ease;
  border-left: 3px solid transparent;
}

.navigation-links a:hover,
.supplemental-links a:hover {
  background: #f8f9fa;
  color: #007bff;
}

.icon {
  font-size: 1.2em;
  width: 1.5rem;
  text-align: center;
}

.supplemental-links {
  border-top: 1px solid #eee;
  margin-top: 1rem;
}

/* Responsive design */
@media (max-width: 768px) {
  .sidebar {
    width: 100%;
    left: -100%;
  }
  
  .sidebar.shown {
    left: 0;
  }
}
</style> 
