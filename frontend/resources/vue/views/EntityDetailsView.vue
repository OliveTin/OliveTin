<template>
  <Section>
    <template #title>
      <span class="section-title-with-icon">
        Entity Details:
        <ActionIconGlyph
          v-if="entityIcon"
          class="entity-title-icon"
          :glyph="entityIcon"
        />
        <span v-if="entityDetails?.title">{{ entityDetails.title }}</span>
      </span>
    </template>
    <template #toolbar>
      <button
        class="back-button"
        @click="goBack"
      >
        <HugeiconsIcon
          :icon="ArrowLeftIcon"
          width="1.2em"
          height="1.2em"
        />
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
          <router-link
            :to="{ name: 'Entities' }"
            class="entity-type-link"
          >
            {{ entityType }}
          </router-link>
        </dd>
        <dt v-if="entityDetails.title">
          Title
        </dt>
        <dd v-if="entityDetails.title">
          {{ entityDetails.title }}
        </dd>
        <template v-if="entityDetails.fields">
          <template
            v-for="(value, key) in entityDetails.fields"
            :key="key"
          >
            <dt>{{ key }}</dt>
            <dd>{{ value }}</dd>
          </template>
        </template>
      </dl>
      <p v-if="!entityDetails.title && (!entityDetails.fields || Object.keys(entityDetails.fields).length === 0)">
        No details available for this entity.
      </p>
    </template>
  </Section>

  <Section
    v-if="entityDetails"
    title="Dashboard Entity Directories"
  >
    <div
      v-if="filteredDirectories.length > 0"
      class="directories-section"
    >
      <ul class="directory-list">
        <li
          v-for="(directory, idx) in filteredDirectories"
          :key="idx"
        >
          <router-link
            :to="{
              name: 'Dashboard',
              params: {
                title: directory,
                entityType: entityType,
                entityKey: entityKey
              }
            }"
          >
            {{ directory }}
          </router-link>
        </li>
      </ul>
    </div>
    <p v-else>
      No directories found for this entity.
      <a
        href="https://docs.olivetin.app/dashboards/entity-directories.html"
        target="_blank"
        rel="noopener noreferrer"
      >Learn more</a>
    </p>
  </Section>

  <section
    v-if="entityDetails && relatedActions.length > 0"
    class="transparent"
  >
    <div class="dashboard-row">
      <fieldset>
        <legend class="visually-hidden">
          Related actions
        </legend>
        <template
          v-for="(related, idx) in relatedActions"
          :key="related.action?.bindingId || idx"
        >
          <ActionButton
            v-if="related.action"
            :action-data="related.action"
            :prefilled-arguments="related.prefilledArguments"
          />
        </template>
      </fieldset>
    </div>
  </section>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { HugeiconsIcon } from '@hugeicons/vue'
import { ArrowLeftIcon } from '@hugeicons/core-free-icons'
import Section from 'picocrank/vue/components/Section.vue'
import ActionButton from '../ActionButton.vue'
import ActionIconGlyph from '../components/ActionIconGlyph.vue'

const router = useRouter()
const entityDetails = ref(null)

const props = defineProps({
  entityType: String,
  entityKey: String
})

const filteredDirectories = computed(() => {
  if (!entityDetails.value?.directories) {
    return []
  }
  return entityDetails.value.directories.filter(d => d)
})

const relatedActions = computed(() => entityDetails.value?.relatedActions ?? [])

const entityIcon = computed(() => entityDetails.value?.icon ?? '')

function goBack () {
  router.push({ name: 'Entities' })
}

async function fetchEntityDetails () {
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

.section-title-with-icon {
	display: inline-flex;
	align-items: center;
	gap: 0.5em;
}

.entity-title-icon {
	font-size: 1.2em;
}

fieldset {
	display: grid;
	grid-template-columns: repeat(auto-fit, 180px);
	grid-auto-rows: 1fr;
	justify-content: center;
	place-items: stretch;
}

@media (prefers-color-scheme: dark) {
    .back-button {
        background-color: var(--bg, #111);
        border-color: var(--border-color, #333);
    }

    .back-button:hover {
        background-color: var(--bg-hover, #222);
    }

    .directory-list a:hover {
        background-color: var(--bg-hover, #222);
    }
}
</style>
