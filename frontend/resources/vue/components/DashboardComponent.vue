<template>
    <ActionButton v-if="component.type == 'link'" :actionData="component.action" :key="component.title" />

    <div v-else-if="component.type == 'directory'">
        <router-link :to="{ name: 'Dashboard', params: { title: component.title } }" class="dashboard-link">
            <button>
                {{ component.title }}
            </button>
        </router-link>
    </div>

    <div v-else-if="component.type == 'display'" class="display">
        <div v-html="component.title" />
    </div>

    <DashboardComponentMostRecentExecution v-else-if="component.type == 'stdout-most-recent-execution'" :component="component" />

    <template v-else-if="component.type == 'fieldset'">
        <template v-for="subcomponent in component.contents" :key="subcomponent.title">
            <DashboardComponent :component="subcomponent" />
        </template>
    </template>

    <div v-else>
        OTHER: {{ component.type }}
        {{ component }}
    </div>

</template>

<script setup>
import ActionButton from '../ActionButton.vue'
import DashboardComponentMostRecentExecution from './DashboardComponentMostRecentExecution.vue'

const props = defineProps({
    component: {
        type: Object,
        required: true
    }
})
</script>