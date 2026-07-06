<template>
	<Section :padding="!hasTable">
		<template #title>
			<span class="section-title-with-icon">
				Entity:
				<ActionIconGlyph v-if="definition.icon" class="entity-title-icon" :glyph="definition.icon" />
				{{ definition.title }}
			</span>
		</template>
		<template v-if="hasTable" #toolbar>
			<EntityListFilter v-model="searchText" />
		</template>

		<p v-if="!hasTable">{{ definition.instances.length }} instances.</p>

		<template v-if="hasTable">
			<p v-if="tableError" class="table-error padding" role="alert">{{ tableError }}</p>
			<EntityInstancesTable
				v-else
				:instances="tableInstances"
				:properties="definition.properties"
				:total-instances="totalInstances"
				v-model:page="currentPage"
				v-model:page-size="pageSize"
			/>
		</template>

		<ul v-else>
			<li v-for="inst in definition.instances" :key="inst.uniqueKey">
				<router-link :to="entityDetailsRoute(inst)">
					{{ inst.title }}
				</router-link>
			</li>
		</ul>

		<div v-if="usedDashboards.length > 0" :class="{ padding: hasTable }">
			<h3>Used on Dashboards:</h3>
			<ul>
				<li v-for="dash in usedDashboards" :key="dash">
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

<script setup>
	import { computed, ref, watch, onMounted } from 'vue'
	import Section from 'picocrank/vue/components/Section.vue'
	import ActionIconGlyph from './ActionIconGlyph.vue'
	import EntityInstancesTable from './EntityInstancesTable.vue'
	import EntityListFilter from './EntityListFilter.vue'

	const props = defineProps({
		definition: {
			type: Object,
			required: true
		}
	})

	const searchText = ref('')
	const tableInstances = ref([])
	const totalInstances = ref(0)
	const currentPage = ref(1)
	const pageSize = ref(10)
	const tableError = ref('')
	let fetchTimer = null

	const hasTable = computed(() => (props.definition.properties?.length ?? 0) > 0)

	const usedDashboards = computed(() => filteredDashboards(props.definition.usedOnDashboards ?? []))

	watch(searchText, () => {
		currentPage.value = 1
		scheduleFetchTableInstances()
	})

	watch([currentPage, pageSize], () => {
		scheduleFetchTableInstances()
	})

	function entityDetailsRoute(inst) {
		return {
			name: 'EntityDetails',
			params: {
				entityType: inst.type,
				entityKey: inst.uniqueKey
			}
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

	function scheduleFetchTableInstances() {
		if (!hasTable.value) {
			return
		}

		if (fetchTimer) {
			clearTimeout(fetchTimer)
		}

		fetchTimer = setTimeout(() => {
			fetchTableInstances()
		}, 250)
	}

	async function fetchTableInstances() {
		if (!hasTable.value) {
			return
		}

		tableError.value = ''
		try {
			const response = await window.client.getEntities({
				entityType: props.definition.title,
				filter: searchText.value.trim(),
				page: currentPage.value,
				pageSize: pageSize.value
			})

			const definition = response.entityDefinitions?.find(def => def.title === props.definition.title)
			tableInstances.value = definition?.instances ?? []
			totalInstances.value = definition?.totalInstances ?? 0
		} catch (err) {
			console.error('Failed to fetch entity instances:', err)
			tableError.value = 'Failed to load entity instances.'
			tableInstances.value = []
			totalInstances.value = 0
		}
	}

	onMounted(() => {
		if (hasTable.value) {
			fetchTableInstances()
		}
	})
</script>

<style scoped>
.section-title-with-icon {
	display: inline-flex;
	align-items: center;
	gap: 0.5em;
}

.entity-title-icon {
	font-size: 1.2em;
}

.table-error {
	color: var(--error, #c00);
}
</style>
