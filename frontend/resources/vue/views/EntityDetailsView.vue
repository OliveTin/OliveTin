<template>
	<section class="with-header-and-content">
		<div class="section-header">
			<h2>Entity Details</h2>
		</div>

		<div class="section-content">
			<p v-if="!entityDetails">Loading entity details...</p>
			<p v-else-if="!entityDetails.title">No details available for this entity.</p>
			<p v-else>{{ entityDetails.title }}</p>
		</div>
	</section>
</template>

<script setup>
	import { ref, onMounted, onBeforeUnmount } from 'vue'

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
