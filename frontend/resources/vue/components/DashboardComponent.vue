<template>
    <ActionButton v-if="component.type == 'link'" :actionData="component.action" :cssClass="component.cssClass" :key="component.title" />

    <DashboardComponentDirectory v-else-if="component.type == 'directory'" :component="component" />

    <DashboardComponentDisplay v-else-if="component.type == 'display'" :component="component" />

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
import DashboardComponentDirectory from './DashboardComponentDirectory.vue'
import DashboardComponentDisplay from './DashboardComponentDisplay.vue'

const props = defineProps({
    component: {
        type: Object,
        required: true
    }
})
</script>