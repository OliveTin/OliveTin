<template>
  <aside :class="{ 'shown': isOpen, 'stuck': isStuck }" class="sidebar">
    <button
      class="stick-toggle"
      :aria-pressed="isStuck"
      :title="isStuck ? 'Unstick sidebar' : 'Stick sidebar'"
      @click="toggleStick"
    >
      <span v-if="isStuck">📌 Unstick</span>
      <span v-else>📍 Stick</span>
    </button>
    <nav class="mainnav">
      <ul class="navigation-links">
        <li v-for="link in navigationLinks" :key="link.id" :title="link.title">
          <router-link :to="link.path" :class="{ active: isActive(link.path) }">
            <span v-if="link.icon" class="icon" v-html="link.icon"></span>
            <span class="title">{{ link.title }}</span>
          </router-link>
        </li>
      </ul>

      <ul class="supplemental-links">
        <li v-for="link in supplementalLinks" :key="link.id" :title="link.title">
          <a :href="link.url" :target="link.target || '_self'">
            <span v-if="link.icon" class="icon" v-html="link.icon"></span>
            <span class="title">{{ link.title }}</span>
          </a>
        </li>
      </ul>
    </nav>
  </aside>
</template>

<script setup>
import { ref, onMounted, getCurrentInstance } from 'vue'
import { useRoute } from 'vue-router'

const isOpen = ref(false)
const isStuck = ref(false)
const navigationLinks = ref([
  {
    id: 'actions',
    title: 'Actions',
    path: '/',
    icon: '⚡'
  },
  {
    id: 'logs',
    title: 'Logs',
    path: '/logs',
    icon: '📋'
  },
  {
    id: 'diagnostics',
    title: 'Diagnostics',
    path: '/diagnostics',
    icon: '🔧'
  }
])
const supplementalLinks = ref([])

const route = useRoute()
const instance = getCurrentInstance()

function toggleStick() {
  isStuck.value = !isStuck.value
}

function toggle() {
  isOpen.value = !isOpen.value
}

function open() {
  isOpen.value = true
}

function close() {
  isOpen.value = false
}

function isActive(path) {
  return route.path === path
}

// Method to add navigation links from other components
function addNavigationLink(link) {
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
.mainnav {
  padding: 1rem 0;
}

.navigation-links,
.supplemental-links {
  list-style: none;
  margin: 0;
  padding: 0;
}

.navigation-links li,
.supplemental-links li {
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
  text-decoration: none;
  transition: background-color 0.2s ease;
  border-left: 3px solid transparent;
}

.navigation-links a:hover,
.supplemental-links a:hover {
  background: #f8f9fa;
  color: #007bff;
}

.navigation-links a.active {
  background: #e3f2fd;
  color: #007bff;
  border-left-color: #007bff;
}

.navigation-links a.router-link-active {
  background: #e3f2fd;
  color: #007bff;
  border-left-color: #007bff;
}

.icon {
  font-size: 1.2em;
  width: 1.5rem;
  text-align: center;
}

.title {
  font-weight: 500;
}

.supplemental-links {
  border-top: 1px solid #eee;
  margin-top: 1rem;
  padding-top: 1rem;
}

.supplemental-links a {
  font-size: 0.9rem;
  color: #666;
}

.supplemental-links a:hover {
  color: #007bff;
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