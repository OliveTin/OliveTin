<template>
    <Header title="OliveTin" :logoUrl="logoUrl" @toggleSidebar="toggleSidebar">
        <template #toolbar>
            <div id="banner" v-if="bannerMessage" :style="bannerCss">
                <p>{{ bannerMessage }}</p>
            </div>
        </template>

        <template #user-info>
            <div class="flex-row" style="gap: .5em;">
                <span id="link-login" v-if="!isLoggedIn"><router-link to="/login">Login</router-link></span>
                <div v-else>
                    <span id="username-text" :title="'Provider: ' + userProvider">{{ username }}</span>
                    <span id="link-logout" v-if="isLoggedIn"><a href="/api/Logout">Logout</a></span>
                </div>
                <HugeiconsIcon :icon="UserCircle02Icon" width = "1.5em" height = "1.5em" />
            </div>

        </template>
    </Header>

    <div id="layout">
        <Sidebar ref="sidebar" id = "mainnav" v-if="showNavigation && !initError" />

		<div id="content" initial-martial-complete="{{ hasLoaded }}">
            <main title="Main content">
                <section v-if="initError" class="error-container error" style="text-align: center; padding: 2em;">
                    <h2>Failed to Initialize OliveTin</h2>
                    <p><strong>Error Message:</strong> {{ initErrorMessage }}</p>
                    <p>Please check the your browser console first, and then the server logs for more details.</p>
                    <button @click="retryInit" class="bad">Retry</button>
                </section>
                <router-view v-else :key="$route.fullPath" />
            </main>

            <footer title="footer" v-if="showFooter && !initError">
                <p>
                    <img title="application icon" :src="logoUrl" alt="OliveTin logo" style="height: 1em;" class="logo" />
                    OliveTin {{ currentVersion }}
                </p>
                <p>
                    <span>
                        <a href="https://docs.olivetin.app" target="_new">Documentation</a>
                    </span>

                    <span>
                        <a href="https://github.com/OliveTin/OliveTin/issues/new/choose" target="_new">Raise an issue on
                            GitHub</a>
                    </span>

                    <span>{{ serverConnection }}</span>
                </p>
                <p>
                    <a id="available-version" href="http://olivetin.app" target="_blank" hidden>?</a>
                </p>
            </footer>
        </div>
    </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import Sidebar from 'picocrank/vue/components/Sidebar.vue';
import Header from 'picocrank/vue/components/Header.vue';
import { HugeiconsIcon } from '@hugeicons/vue'
import { Menu01Icon } from '@hugeicons/core-free-icons'
import { UserCircle02Icon } from '@hugeicons/core-free-icons'
import { DashboardSquare01Icon } from '@hugeicons/core-free-icons'
import logoUrl from '../../OliveTinLogo.png';

const sidebar = ref(null);
const username = ref('guest');
const userProvider = ref('system');
const isLoggedIn = ref(false);
const serverConnection = ref('Connected');
const currentVersion = ref('?');
const bannerMessage = ref('');
const bannerCss = ref('');
const hasLoaded = ref(false);
const showFooter = ref(true)
const showNavigation = ref(true)
const showLogs = ref(true)
const showDiagnostics = ref(true)
const initError = ref(false)
const initErrorMessage = ref('')

function toggleSidebar() {
    sidebar.value.toggle()
}

async function requestInit() {
    try {
        const initResponse = await window.client.init({})

        window.initResponse = initResponse
        window.initError = false
        window.initErrorMessage = ''
        window.initCompleted = true

        username.value = initResponse.authenticatedUser
        isLoggedIn.value = initResponse.authenticatedUser !== '' && initResponse.authenticatedUser !== 'guest'
        currentVersion.value = initResponse.currentVersion
		bannerMessage.value = initResponse.bannerMessage || '';
		bannerCss.value = initResponse.bannerCss || '';
		showFooter.value = initResponse.showFooter
        showNavigation.value = initResponse.showNavigation
        showLogs.value = initResponse.showLogList
        showDiagnostics.value = initResponse.showDiagnostics

        for (const rootDashboard of initResponse.rootDashboards) {
            sidebar.value.addNavigationLink({
                id: rootDashboard,
                name: rootDashboard,
                title: rootDashboard,
                path: rootDashboard === 'Actions' ? '/' : `/dashboards/${rootDashboard}`,
                icon: DashboardSquare01Icon,
            })
        }

        sidebar.value.addSeparator()
        sidebar.value.addRouterLink('Entities')

        if (showLogs.value) {
            sidebar.value.addRouterLink('Logs')
        }

        if (showDiagnostics.value) {
            sidebar.value.addRouterLink('Diagnostics')
        }

        hasLoaded.value = true;
        initError.value = false;
        
        // Only start websocket connection after successful init
        if (window.checkWebsocketConnection) {
            window.checkWebsocketConnection()
        }
    } catch (error) {
        console.error("Error initializing client", error)
        initError.value = true
        initErrorMessage.value = error.message || 'Failed to connect to OliveTin server'
        window.initError = true
        window.initErrorMessage = error.message || 'Failed to connect to OliveTin server'
        window.initCompleted = false
        serverConnection.value = 'Disconnected'
    }
}

function retryInit() {
    initError.value = false
    initErrorMessage.value = ''
    window.initError = false
    window.initErrorMessage = ''
    window.initCompleted = false
    requestInit()
}

onMounted(() => {
    serverConnection.value = 'Connected';
    // Initialize global state
    window.initError = false
    window.initErrorMessage = ''
    window.initCompleted = false
    requestInit()
})
</script>
