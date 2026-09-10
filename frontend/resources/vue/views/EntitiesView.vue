<template>
  <Section
    v-if="!definitionsLoaded"
    title="Loading entity definitions..."
    :icon="CellsIcon"
  />
  <Section
    v-else-if="totalInstances === 0"
    title="There are no entities to show yet."
    :icon="CellsIcon"
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
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { CellsIcon } from '@hugeicons/core-free-icons'
import Section from 'picocrank/vue/components/Section.vue'
import EntityDefinitionSection from '../components/EntityDefinitionSection.vue'
const definitionsLoaded = ref(false)
const entityDefinitions = ref([])
let entityFetchGeneration = 0

const totalInstances = computed(() =>
  entityDefinitions.value.reduce(
    (sum, def) => sum + (def.instances?.length ?? 0),
    0
  )
)

async function fetchEntities () {
  const fetchGeneration = ++entityFetchGeneration
  try {
    const ret = await window.client.getEntities()
    if (fetchGeneration !== entityFetchGeneration) return
    entityDefinitions.value = ret.entityDefinitions ?? []
  } catch (err) {
    if (fetchGeneration !== entityFetchGeneration) return
    console.error('Failed to fetch entities:', err)
    window.showBigError('fetch-entities', 'getting entities', err, false)
    entityDefinitions.value = []
  } finally {
    if (fetchGeneration === entityFetchGeneration) {
      definitionsLoaded.value = true
    }
  }
}

onMounted(() => {
  fetchEntities()
  window.addEventListener('EventEntityChanged', fetchEntities)
  window.addEventListener('EventConfigChanged', fetchEntities)
})

onUnmounted(() => {
  window.removeEventListener('EventEntityChanged', fetchEntities)
  window.removeEventListener('EventConfigChanged', fetchEntities)
})
</script>
