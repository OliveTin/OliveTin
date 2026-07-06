<template>
	<Table
		:data="tableRows"
		:headers="headers"
		:show-pagination="false"
	>
		<template #cell-title="{ row, value }">
			<router-link :to="entityDetailsRoute(row)">
				{{ value }}
			</router-link>
		</template>
	</Table>

	<div v-if="totalInstances > 0" class="padding">
		<Pagination
			:total="totalInstances"
			v-model:page="currentPageModel"
			v-model:page-size="pageSizeModel"
			item-title="entities"
		/>
	</div>
</template>

<script setup>
	import { computed } from 'vue'
	import Table from 'picocrank/vue/components/Table.vue'
	import Pagination from 'picocrank/vue/components/Pagination.vue'

	const props = defineProps({
		instances: {
			type: Array,
			required: true
		},
		properties: {
			type: Array,
			required: true
		},
		totalInstances: {
			type: Number,
			default: 0
		},
		page: {
			type: Number,
			default: 1
		},
		pageSize: {
			type: Number,
			default: 10
		}
	})

	const emit = defineEmits(['update:page', 'update:pageSize'])

	const headers = computed(() => {
		const propertyHeaders = props.properties.map(property => ({
			key: property.name,
			label: property.title,
			sortable: true
		}))

		return [
			{ key: 'title', label: 'Name', sortable: true },
			...propertyHeaders
		]
	})

	const tableRows = computed(() =>
		props.instances.map(instance => ({
			...instance,
			...instance.fields
		}))
	)

	const currentPageModel = computed({
		get: () => props.page,
		set: value => emit('update:page', value)
	})

	const pageSizeModel = computed({
		get: () => props.pageSize,
		set: value => emit('update:pageSize', value)
	})

	function entityDetailsRoute(row) {
		return {
			name: 'EntityDetails',
			params: {
				entityType: row.type,
				entityKey: row.uniqueKey
			}
		}
	}
</script>

<style scoped>
a {
	text-decoration: none;
}

a:hover {
	text-decoration: underline;
}
</style>
