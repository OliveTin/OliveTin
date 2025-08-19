<template>
    <header>
        <div id="sidebar-button" class="flex-row" @click="toggleSidebar">
            <img src="../../OliveTinLogo.png" alt="OliveTin logo" class="logo" />

            <h1 id="page-title">OliveTin</h1>

            <div class="fg1" />
            <button id="sidebar-toggler-button" aria-label="Open sidebar navigation" aria-pressed="false" aria-haspopup="menu" class="neutral">
                <HugeiconsIcon :icon="Menu01Icon" width = "1em" height = "1em" />
            </button>
        </div>

        <div class="fg1">
            <Breadcrumbs />
        </div>

		<div id="banner" v-if="bannerMessage" :style="bannerCss">
			<p>{{ bannerMessage }}</p>
		</div>

        <div class="flex-row" style="gap: .5em;">
            <span id="link-login" v-if="!isLoggedIn"><router-link to="/login">Login</router-link></span>
            <span id="link-logout" v-if="isLoggedIn"><a href="/api/Logout">Logout</a></span>
            <span id="username-text" :title="'Provider: ' + userProvider">{{ username }}</span>
            <HugeiconsIcon :icon="UserCircle02Icon" width = "1.5em" height = "1.5em" />
        </div>
    </header>

    <div id="layout">
        <Sidebar ref="sidebar" />

        <div id="content">
            <main title="Main content">
                <router-view :key="$route.fullPath" />
            </main>

            <footer title="footer">
                <p>
                    <img title="application icon" src="../../OliveTinLogo.png" alt="OliveTin logo" height="1em"
                        class="logo" />
                    OliveTin 3000!
                </p>
                <p>
                    <span>
                        <a href="https://docs.olivetin.app" target="_new">Documentation</a>
                    </span>

                    <span>
                        <a href="https://github.com/OliveTin/OliveTin/issues/new/choose" target="_new">Raise an issue on
                            GitHub</a>
                    </span>

                    <span>{{ currentVersion }}</span>

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
import Sidebar from './components/Sidebar.vue';
import { HugeiconsIcon } from '@hugeicons/vue'
import { Menu01Icon } from '@hugeicons/core-free-icons'
import { UserCircle02Icon } from '@hugeicons/core-free-icons'

const sidebar = ref(null);
const username = ref('guest');
const userProvider = ref('system');
const isLoggedIn = ref(false);
const serverConnection = ref('Connected');
const currentVersion = ref('?');
const bannerMessage = ref('');
const bannerCss = ref('');

function toggleSidebar() {
    sidebar.value.toggle()
}

async function requestInit() {
    try {
        const initResponse = await window.client.init({})

        console.log("init response", initResponse)

        username.value = initResponse.authenticatedUser
        currentVersion.value = initResponse.currentVersion
		bannerMessage.value = initResponse.bannerMessage || '';
		bannerCss.value = initResponse.bannerCss || '';

        for (const rootDashboard of initResponse.rootDashboards) {
            sidebar.value.addNavigationLink({
                id: rootDashboard,
                title: rootDashboard,
                path: `/dashboards/${rootDashboard}`,
                icon: '📊'
            })
        }
    } catch (error) {
        console.error("Error initializing client", error)
    }
}

onMounted(() => {
    serverConnection.value = 'Connected';
    requestInit()
})
</script>
