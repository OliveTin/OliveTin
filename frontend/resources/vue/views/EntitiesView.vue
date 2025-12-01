<template>
	<Section class = "with-header-and-content" v-if="entityDefinitions.length === 0" title="Loading entity definitions...">
		<div class = "section-header">
			<h2 class="loading-message">
				Loading entity definitions...
			</h2>
		</div>
	</Section>
	<template v-else>
		<Section v-for="def in entityDefinitions" :key="def.title" :title="'Entity: ' + def.title ">
			<div class = "section-content">
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
			</div>
		</Section>
	</template>
</template>

<script setup>
	import { ref, onMounted } from 'vue'
	import Section from 'picocrank/vue/components/Section.vue'

	const entityDefinitions = ref([])

	async function fetchEntities() {
	    const ret = await window.client.getEntities()

        entityDefinitions.value = ret.entityDefinitions
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
