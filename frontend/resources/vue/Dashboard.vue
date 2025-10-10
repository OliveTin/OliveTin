<template>
    <section v-if="!dashboard && !initError" style = "text-align: center; padding: 2em;">
        <HugeiconsIcon :icon="Loading03Icon" width="3em" height="3em" style="animation: spin 1s linear infinite;" />
        <p>Loading dashboard...</p>
        <p style="color: var(--fg2);">{{ loadingTime }}s</p>
    </section>
    <section v-if="initError" style="text-align: center; padding: 2em;" class = "bad">
        <h2 style="color: var(--error);">Initialization Failed</h2>
        <p>{{ initError }}</p>
        <p style="color: var(--fg2);">Please check your configuration and try again.</p>
    </section>
    <div v-else-if="dashboard">
        <section v-if="dashboard.contents.length == 0">
            <legend>{{ dashboard.title }}</legend>
            <p style = "text-align: center" class = "padding">This dashboard is empty.</p>
        </section>

        <section class="transparent" v-else>
            <div v-for="component in dashboard.contents" :key="component.title">
                <fieldset>
                    <legend v-if = "dashboard.title != 'Default'">{{ component.title }}</legend>

                    <template v-for="subcomponent in component.contents">
                        <DashboardComponent :component="subcomponent" />
                    </template>
                </fieldset>
            </div>
        </section>
    </div>
</template>

<script setup>
import DashboardComponent from './components/DashboardComponent.vue'
import { onMounted, onUnmounted, ref } from 'vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import { Loading03Icon } from '@hugeicons/core-free-icons'

const props = defineProps({
    title: {
        type: String,
        required: false
    }
})

const dashboard = ref(null)
const loadingTime = ref(0)
const initError = ref(null)
let loadingTimer = null
let checkInitInterval = null

async function getDashboard() {
    let title = props.title

    // If no specific title was provided or it's the placeholder 'default',
    // prefer the first configured root dashboard (e.g., "Test").
    if ((!title || title === 'default') && window.initResponse.rootDashboards && window.initResponse.rootDashboards.length > 0) {
        title = window.initResponse.rootDashboards[0]
    }

    try {
        const ret = await window.client.getDashboard({
            title: title,
        })

        if (!ret || !ret.dashboard) {
            throw new Error('No dashboard found')
        }

        dashboard.value = ret.dashboard 
        document.title = ret.dashboard.title + ' - OliveTin'
        
        // Clear any previous init error since we successfully loaded
        initError.value = null
        
        // Stop the loading timer once dashboard is loaded
        if (loadingTimer) {
            clearInterval(loadingTimer)
            loadingTimer = null
        }
        
        // Set attribute to indicate dashboard is loaded successfully
        document.body.setAttribute('loaded-dashboard', title || 'default')
    } catch (e) {
        // On error, provide a safe fallback state
        console.error('Failed to load dashboard', e)
        dashboard.value = { title: title || 'Default', contents: [] }
        document.title = 'Error - OliveTin'
        
        // Stop the loading timer on error
        if (loadingTimer) {
            clearInterval(loadingTimer)
            loadingTimer = null
        }
        
        // Set attribute even on error so tests can proceed
        document.body.setAttribute('loaded-dashboard', title || 'error')
    }
}

function waitForInitAndLoadDashboard() {
    // Start the loading timer
    loadingTime.value = 0
    loadingTimer = setInterval(() => {
        loadingTime.value++
    }, 1000)
    
    // Check if init has completed successfully
    if (window.initCompleted && window.initResponse) {
        getDashboard()
    } else if (window.initError) {
        // Init failed, show error immediately
        initError.value = window.initErrorMessage || 'Initialization failed. Please check your configuration and try again.'
        // Stop the loading timer since we're showing an error
        if (loadingTimer) {
            clearInterval(loadingTimer)
            loadingTimer = null
        }
    } else {
        // Init hasn't completed yet, poll for completion
        checkInitInterval = setInterval(() => {
            if (window.initCompleted && window.initResponse) {
                clearInterval(checkInitInterval)
                checkInitInterval = null
                getDashboard()
            } else if (window.initError) {
                clearInterval(checkInitInterval)
                checkInitInterval = null
                initError.value = window.initErrorMessage || 'Initialization failed. Please check your configuration and try again.'
                // Stop the loading timer since we're showing an error
                if (loadingTimer) {
                    clearInterval(loadingTimer)
                    loadingTimer = null
                }
            }
        }, 100) // Check every 100ms
    }
}

onMounted(() => {
    waitForInitAndLoadDashboard()
})

onUnmounted(() => {
    // Clean up the timers when component is unmounted
    if (loadingTimer) {
        clearInterval(loadingTimer)
        loadingTimer = null
    }
    if (checkInitInterval) {
        clearInterval(checkInitInterval)
        checkInitInterval = null
    }
})

</script>

<style>
fieldset {
    display: grid;
    grid-template-columns: repeat(auto-fit, 180px);
    grid-auto-rows: 1fr;
    justify-content: center;
    place-items: stretch;
}

@keyframes spin {
    from {
        transform: rotate(0deg);
    }
    to {
        transform: rotate(360deg);
    }
}
</style>