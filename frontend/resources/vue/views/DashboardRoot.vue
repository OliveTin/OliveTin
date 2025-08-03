<template>
    <div v-for="dashboard in dashboards" :key="dashboard.id">       
        <Dashboard :dashboard="dashboard" />
    </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import Dashboard from '../Dashboard.vue'

const dashboards = ref([])

async function refreshActions() {
    const ret = await window.client.getDashboardComponents();

    console.log(ret.dashboards)
    dashboards.value = ret.dashboards
}

onMounted(() => {
    refreshActions()
})
</script>