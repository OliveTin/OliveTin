<template>
	<Section v-if="!definitionsLoaded" title="Loading entity definitions..." />
	<Section
		v-else-if="totalInstances === 0"
		title="There are no entities to show yet."
	>
		<p>
			When OliveTin has registered entity instances (for example from entity files or your setup), they will be listed here.
		</p>
	</Section>
	<template v-else>
		<Section v-for="def in entityDefinitions" :key="def.title" :title="'Entity: ' + def.title ">
			<p>{{ def.instances.length }} instances.</p>

			<ul>
				<li v-for="inst in def.instances" :key="inst.uniqueKey">
					<router-link :to="{ name: 'EntityDetails', params: { entityType: inst.type, entityKey: inst.uniqueKey } }">
						{{ inst.title }}
					</router-link>
				</li>
			</ul>

			<h3>Used on Dashboards:</h3>
			<ul>
				<li v-for="dash in filteredDashboards(def.usedOnDashboards)" :key="dash">
					<template v-if="isEntityDirectory(dash)">
						{{ getDashboardTitle(dash) }} <span class="entity-directory-label">[Entity Directory]</span>
					</template>
					<router-link v-else-if="!dash.includes('entity:')" :to="{ name: 'Dashboard', params: { title: getDashboardTitle(dash) } }">
						{{ getDashboardTitle(dash) }}
					</router-link>
					<span v-else>{{ dash }}</span>
				</li>
			</ul>
		</Section>
	</template>
</template>

<script setup>
	import { ref, computed, onMounted } from 'vue'
	import Section from 'picocrank/vue/components/Section.vue'

	const definitionsLoaded = ref(false)
	const entityDefinitions = ref([])

	const totalInstances = computed(() =>
		entityDefinitions.value.reduce(
			(sum, def) => sum + (def.instances?.length ?? 0),
			0,
		),
	)

	async function fetchEntities() {
		try {
			const ret = await window.client.getEntities()
			entityDefinitions.value = ret.entityDefinitions ?? []
		} catch (err) {
			console.error('Failed to fetch entities:', err)
			window.showBigError('fetch-entities', 'getting entities', err, false)
			entityDefinitions.value = []
		} finally {
			definitionsLoaded.value = true
		}
	}

	function filteredDashboards(dashboards) {
		return dashboards.filter(d => d && !d.includes('{{'))
	}

	function isEntityDirectory(dashboardTitle) {
		return dashboardTitle.endsWith(' [Entity Directory]')
	}

	function getDashboardTitle(dashboardTitle) {
		if (isEntityDirectory(dashboardTitle)) {
			return dashboardTitle.slice(0, -' [Entity Directory]'.length)
		}
		return dashboardTitle
	}

    onMounted(() => {
        fetchEntities()
	})
</script>
