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

  <div
    v-if="totalInstances > 0"
    class="padding"
  >
    <Pagination
      v-model:page="currentPageModel"
      v-model:page-size="pageSizeModel"
      :total="totalInstances"
      item-title="entities"
    />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import Table from 'picocrank/vue/components/Table.vue'
import Pagination from 'picocrank/vue/components/Pagination.vue'
import { entityDetailsRoute } from '../utils/entityRoutes.js'

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
    label: property.title
  }))

  return [
    { key: 'title', label: 'Name' },
    ...propertyHeaders
  ]
})

const tableRows = computed(() =>
  props.instances.map(instance => ({
    ...instance.fields,
    title: instance.title,
    type: instance.type,
    uniqueKey: instance.uniqueKey
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

</script>

<style scoped>
a {
	text-decoration: none;
}

a:hover {
	text-decoration: underline;
}
</style>
