<template>
	<aside :class="{ 'shown': isOpen, 'stuck': isStuck }" class="sidebar">
		<div class = "flex-row">
			<h2>Navigation</h2>
			<div class = "fg1" />
				<button class="stick-toggle" :aria-pressed="isStuck" :title="isStuck ? 'Unstick sidebar' : 'Stick sidebar'"	@click="toggleStick">
					<span v-if="isStuck">
						<HugeiconsIcon :icon="Pin02Icon" width = "1em" height = "1em" :strokeWidth = 3 />
					</span>
					<span v-else>
						<HugeiconsIcon :icon="PinIcon" width = "1em" height = "1em" :strokeWidth = 3 />
					</span>
				</button>
			</div>

			<nav class="mainnav">
				<ul class="navigation-links">
					<li v-for="link in navigationLinks" :key="link.name" :title="link.title">
						<!-- Render separator if link is a separator -->
						<div v-if="link.type === 'separator'" class="separator"></div>
						<div v-else-if="link.type === 'callback'">
							<a href="#" @click.prevent="link.callback()">
								<HugeiconsIcon :icon="link.icon" />
								<span>{{ link.title }}</span>
							</a>
						</div>
						<div v-else-if="link.type === 'html'" v-html="link.html"></div>
						<router-link v-else :to="link.path" :class="{ active: isActive(link.path) }">
							<HugeiconsIcon :icon="link.icon" />
							<span>{{ link.title }}</span>
						</router-link>
					</li>
				</ul>
			</nav>
	</aside>
</template>

<script setup>
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { HugeiconsIcon } from '@hugeicons/vue'
import { Pin02Icon } from '@hugeicons/core-free-icons'
import { PinIcon } from '@hugeicons/core-free-icons'

const isOpen = ref(false)
const isStuck = ref(false)
const navigationLinks = ref([])

const route = useRoute()
const router = useRouter()

function addRouterLink(link) {
	const route = router.getRoutes().find(r => r.name === link)

	if (route) {
		const routeLink = {
			name: link,
			title: route.meta.title || route.name,
			path: route.path,
			icon: route.meta.icon || PinIcon,
			type: 'route'
		}

		addNavigationLink(routeLink)
	}
}

function addNavigationLink(link) {
  const existingIndex = navigationLinks.value.findIndex(l => l.name === link.name)
  if (existingIndex >= 0) {
	navigationLinks.value[existingIndex] = { ...link }
  } else {
	navigationLinks.value.push({ ...link })
  }
}

function addCallback(title, callback, options = {}) {
  const callbackLink = {
	name: options.name || title.toLowerCase().replace(/\s+/g, '-'),
	type: 'callback',
	icon: options.icon || PinIcon,
	callback: callback || (() => {}),
	title: title
  }

  addNavigationLink(callbackLink)
}

function addSeparator(id) {
  const separator = {
    name: id || `nav-separator-${Date.now()}`,
    type: 'separator',
    title: 'Separator'
  }
  addNavigationLink(separator)
}

function addHtml(html, options = {}) {
  const htmlLink = {
    name: options.name || `html-item-${Date.now()}`,
    type: 'html',
    html: html,
    title: options.title || 'HTML Item'
  }
  addNavigationLink(htmlLink)
}

function removeNavigationLink(linkId) {
  navigationLinks.value = navigationLinks.value.filter(link => link.id !== linkId)
}

function clearNavigationLinks() {
  navigationLinks.value = []
}

function getNavigationLinks() {
  return [...navigationLinks.value]
}

function toggleStick() {
  isStuck.value = !isStuck.value
}

function stick() {
  isStuck.value = true
}

function unstick() {
  isStuck.value = false
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

defineExpose({
  isOpen,
  navigationLinks,
  stick,
  unstick,
  toggleStick,
  toggle,
  open,
  close,
  isActive,
  addNavigationLink,
  addRouterLink,
  addCallback,
  addSeparator,
  addHtml,
  removeNavigationLink,
  clearNavigationLinks,
  getNavigationLinks,
})
</script>

<style scoped>

h2 {
    padding: .75em;
}

.active {
	text-decoration: underline;
}

li {
	margin: 0;
	padding: 0;
}

button {
	border: 0;
}

.navigation-links a {
	display: flex;
	align-items: center;
	gap: 0.75rem;
	padding: .75em;
	border-radius: 0;
}

.separator {
	height: 1px;
	background-color: #eee;
	margin: 0.5rem 0.75rem;
}

.icon {
	font-size: 1.2em;
	width: 1.5rem;
	text-align: center;
}

@media (prefers-color-scheme: dark) {
  .navigation-links a {
	color: #f8f9fa;
  }

  .separator {
	background-color: #444;
  }

  .supplemental-links {
	border-top: 1px solid #444;
  }
}

@media (max-width: 768px) {
  .sidebar {
	  left: -100%;
  }

  .sidebar.shown {
	  left: 0;
  }
}
</style> 
