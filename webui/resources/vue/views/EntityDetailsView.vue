<template>
	<Section title="Entity Details">
		<div>
			<p v-if="!entityDetails">Loading entity details...</p>
			<p v-else-if="!entityDetails.title">No details available for this entity.</p>
			<p v-else>{{ entityDetails.title }}</p>
		</div>
	</Section>
</template>

<script setup>
	import { ref, onMounted } from 'vue'
	import Section from 'picocrank/vue/components/Section.vue'

	const entityDetails = ref(null)

	const props = defineProps({
		entityType: String,
		entityKey: String
	})

	async function fetchEntityDetails() {
		try {
			const response = await window.client.getEntity({
				type: props.entityType,
				uniqueKey: props.entityKey
			})

			entityDetails.value = response
		} catch (err) {
			console.error('Failed to fetch entity details:', err)
			window.showBigError('fetch-entity-details', 'getting entity details', err, false)
		}
	}

	onMounted(() => {
	    fetchEntityDetails()
	})

</script>
