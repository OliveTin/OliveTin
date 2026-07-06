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
		<EntityDefinitionSection
			v-for="def in entityDefinitions"
			:key="def.title"
			:definition="def"
		/>
	</template>
</template>

<script setup>
	import { ref, computed, onMounted } from 'vue'
	import Section from 'picocrank/vue/components/Section.vue'
	import EntityDefinitionSection from '../components/EntityDefinitionSection.vue'

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

	onMounted(() => {
	    fetchEntities()
	})
</script>
