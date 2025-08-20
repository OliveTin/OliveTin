<template>
	<section class = "with-header-and-content" v-if="entityDefinitions.length === 0">
		<div class = "section-header">
			<h2 class="loading-message">
				Loading entity definitions...
			</h2>
		</div>
	</section>
	<template v-else>
		<section v-for="def in entityDefinitions" :key="def.name" class="with-header-and-content">
			<div class = "section-header">
				<h2>Entity: {{ def.title }}</h2>
			</div>

			<div class = "section-content">
				<p>{{ def.instances.length }} instances.</p>

				<ul>
					<li v-for="inst in def.instances" :key="inst.id">
						<router-link :to="{ name: 'EntityDetails', params: { entityType: inst.type, entityKey: inst.uniqueKey } }">
							{{ inst.title }}
						</router-link>
					</li>
				</ul>

				<h3>Used on Dashboards:</h3>
				<ul>
					<li v-for="dash in def.usedOnDashboards">
						<router-link :to="{ name: 'Dashboard', params: { title: dash } }">
							{{ dash }}
						</router-link>
					</li>
				</ul>
			</div>
		</section>
	</template>
</template>

<script setup>
	import { ref, onMounted } from 'vue'

	const entityDefinitions = ref([])

	async function fetchEntities() {
	    const ret = await window.client.getEntities()

        entityDefinitions.value = ret.entityDefinitions
	}

    onMounted(() => {
        fetchEntities()
	})
</script>
