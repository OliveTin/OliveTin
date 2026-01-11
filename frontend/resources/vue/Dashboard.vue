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
    <template v-else-if="dashboard">
        <section v-if="dashboard.contents.length == 0">
            <div class="back-button-container" v-if="isDirectory">
                <button @click="goBack" class="back-button">
                    <HugeiconsIcon :icon="ArrowLeftIcon" width="1.2em" height="1.2em" />
                    <span>Back</span>
                </button>
            </div>
            <h2>{{ dashboard.title }}</h2>
            <p style = "text-align: center" class = "padding">This dashboard is empty.</p>
        </section>

        <section class="transparent" v-else>
            <div class="back-button-container" v-if="isDirectory">
                <button @click="goBack" class="back-button">
                    <HugeiconsIcon :icon="ArrowLeftIcon" width="1.2em" height="1.2em" />
                    <span>Back</span>
                </button>
            </div>
            <div class = "dashboard-row" v-for="component in dashboard.contents" :key="component.title">
                <h2 v-if = "dashboard.title != 'Default'">
                    <router-link 
                        v-if="component.entityType && component.entityKey" 
                        :to="{ 
                            name: 'EntityDetails', 
                            params: { 
                                entityType: component.entityType,
                                entityKey: component.entityKey
                            }
                        }"
                        class="entity-link">
                        {{ component.title }}
                    </router-link>
                    <span v-else>{{ component.title }}</span>
                </h2>

                <fieldset :class="component.cssClass">
                    <template v-for="subcomponent in component.contents">
                        <DashboardComponent :component="subcomponent" />
                    </template>
                </fieldset>
            </div>
        </section>
    </template>
</template>

<script setup>
import DashboardComponent from './components/DashboardComponent.vue'
import { onMounted, onUnmounted, ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { HugeiconsIcon } from '@hugeicons/vue'
import { Loading03Icon, ArrowLeftIcon } from '@hugeicons/core-free-icons'

const props = defineProps({
    title: {
        type: String,
        required: false
    },
    entityType: {
        type: String,
        required: false
    },
    entityKey: {
        type: String,
        required: false
    }
})

const router = useRouter()
const dashboard = ref(null)
const loadingTime = ref(0)
const initError = ref(null)
let loadingTimer = null
let checkInitInterval = null

const isDirectory = computed(() => {
    if (!dashboard.value || !window.initResponse) {
        return false
    }
    const rootDashboards = window.initResponse.rootDashboards || []
    return !rootDashboards.includes(dashboard.value.title) && dashboard.value.title !== 'Actions'
})

function goBack() {
    if (window.history.length > 1) {
        router.back()
    } else {
        const rootDashboards = window.initResponse?.rootDashboards || []
        if (rootDashboards.length > 0) {
            router.push({ name: 'Dashboard', params: { title: rootDashboards[0] } })
        } else {
            router.push({ name: 'Actions' })
        }
    }
}

async function getDashboard() {
    let title = props.title

    // If no specific title was provided or it's the placeholder 'default',
    // prefer the first configured root dashboard (e.g., "Test").
    if ((!title || title === 'default') && window.initResponse.rootDashboards && window.initResponse.rootDashboards.length > 0) {
        title = window.initResponse.rootDashboards[0]
    }

    try {
        const request = {
            title: title,
        }
        
        if (props.entityType && props.entityKey) {
            request.entityType = props.entityType
            request.entityKey = props.entityKey
        }
        
        const ret = await window.client.getDashboard(request)

        if (!ret || !ret.dashboard) {
            throw new Error('No dashboard found')
        }

        dashboard.value = ret.dashboard 
        const pageTitle = window.initResponse?.pageTitle || 'OliveTin'
        document.title = ret.dashboard.title + ' - ' + pageTitle
        
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
        const pageTitle = window.initResponse?.pageTitle || 'OliveTin'
        document.title = 'Error - ' + pageTitle
        
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
    if (window.initResponse) {
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
            if (window.initResponse) {
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

<style scoped>

h2 {
	font-weight: bold;
	text-align: center;
	padding: 1em;
	padding-top: 1.5em;
    grid-column: 1 / -1;
}

h2 .entity-link {
	color: inherit;
	text-decoration: none;
	transition: opacity 0.2s;
}

h2 .entity-link:hover {
	opacity: 0.7;
	text-decoration: underline;
}

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

.back-button-container {
    display: flex;
    justify-content: flex-start;
    padding: 1em;
    padding-bottom: 0;
}

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

@media (prefers-color-scheme: dark) {
    .back-button {
        background-color: var(--bg, #111);
        border-color: var(--border-color, #333);
    }

    .back-button:hover {
        background-color: var(--bg-hover, #222);
    }
}
</style>