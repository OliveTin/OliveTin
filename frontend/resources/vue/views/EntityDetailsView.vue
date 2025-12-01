<template>
	<Section title="Entity Details">
		<template #toolbar>
			<button @click="goBack" class="back-button">
				<HugeiconsIcon :icon="ArrowLeftIcon" width="1.2em" height="1.2em" />
				<span>Back</span>
			</button>
		</template>
		<div v-if="!entityDetails">
			<p>Loading entity details...</p>
		</div>
		<template v-else>
			<dl>
				<dt>Type</dt>
				<dd>
					<router-link :to="{ name: 'Entities' }" class="entity-type-link">
						{{ entityType }}
					</router-link>
				</dd>
				<dt v-if="entityDetails.title">Title</dt>
				<dd v-if="entityDetails.title">{{ entityDetails.title }}</dd>
			</dl>
			<p v-if="!entityDetails.title">No details available for this entity.</p>

			<hr />
			
			<h3>Dashboard Entity Directories</h3>
			<div v-if="entityDetails.directories && entityDetails.directories.length > 0" class="directories-section">
				<ul class="directory-list">
					<li v-for="directory in entityDetails.directories" :key="directory">
						<router-link 
							:to="{ 
								name: 'Dashboard', 
								params: { 
									title: directory,
									entityType: entityType,
									entityKey: entityKey
								}
							}">
							{{ directory }}
						</router-link>
					</li>
				</ul>
			</div>
			<p v-else>No directories found for this entity.
				<a href = "https://docs.olivetin.app/dashboards/entity-directories.html" target = "_blank">Learn more</a>
			</p>
		</template>
	</Section>
</template>

<script setup>
	import { ref, onMounted } from 'vue'
	import { useRouter } from 'vue-router'
	import { HugeiconsIcon } from '@hugeicons/vue'
	import { ArrowLeftIcon } from '@hugeicons/core-free-icons'
	import Section from 'picocrank/vue/components/Section.vue'

	const router = useRouter()
	const entityDetails = ref(null)

	const props = defineProps({
		entityType: String,
		entityKey: String
	})

	function goBack() {
		router.push({ name: 'Entities' })
	}

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

<style scoped>
.back-button {
    display: flex;
    align-items: center;
    gap: 0.5em;
    padding: 0.5em 1em;
    background-color: var(--bg, #fff);
    border: 1px solid var(--border-color, #ccc);
    border-radius: 0.5em;
    cursor: pointer;
    font-size: 0.9em;
    box-shadow: 0 0 .3em rgba(0, 0, 0, 0.1);
    transition: background-color 0.2s, box-shadow 0.2s;
}

.back-button:hover {
    background-color: var(--bg-hover, #f5f5f5);
    box-shadow: 0 0 .5em rgba(0, 0, 0, 0.15);
}

.directories-section h3 {
    margin-bottom: 0.5em;
    font-size: 1.1em;
}

.directory-list a {
    text-decoration: none;
    padding: 0.5em;
    display: inline-block;
    border-radius: 0.3em;
    transition: background-color 0.2s;
}

.directory-list a:hover {
    background-color: var(--bg-hover, #f5f5f5);
    text-decoration: underline;
}

.entity-type-link {
    text-decoration: none;
    transition: opacity 0.2s;
}

.entity-type-link:hover {
    text-decoration: underline;
    opacity: 0.8;
}

hr {
	border: 0;
	border-top: 1px solid var(--border-color, #ccc);
}

@media (prefers-color-scheme: dark) {
    .back-button {
        background-color: var(--bg, #111);
        border-color: var(--border-color, #333);
    }

    .back-button:hover {
        background-color: var(--bg-hover, #222);
    }

    .directories-section {
        border-top-color: var(--border-color, #333);
    }


    .directory-list a:hover {
        background-color: var(--bg-hover, #222);
    }
}
</style>
