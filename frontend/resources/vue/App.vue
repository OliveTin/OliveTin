<template>
    <Header title="OliveTin" :logoUrl="logoUrl" @toggleSidebar="toggleSidebar" :sidebarEnabled="showNavigation">
        <template #toolbar>
            <div id="banner" v-if="bannerMessage" :style="bannerCss">
                <p>{{ bannerMessage }}</p>
            </div>
        </template>

        <template #user-info>
            <div class="flex-row user-info" style="gap: .5em;">
                <span id="link-login" v-if="!isLoggedIn && showLoginLink"><router-link to="/login">{{ t('login-button') }}</router-link></span>
                <router-link v-else to="/user" class="user-link" v-if="isLoggedIn">
                    <span id="username-text">{{ username }}</span>
                </router-link>
                <HugeiconsIcon :icon="UserCircle02Icon" width = "1.5em" height = "1.5em" v-if="isLoggedIn" />
            </div>

        </template>
    </Header>

    <div id="layout">
        <Sidebar ref="sidebar" id = "mainnav" v-if="showNavigation" />

		<div id="content" initial-martial-complete="{{ hasLoaded }}">
            <main title="Main content">
                <router-view :key="$route.fullPath" />
            </main>

            <footer title="footer" v-if="showFooter">
                <p>
                    <img title="application icon" :src="logoUrl" alt="OliveTin logo" style="height: 1em;" class="logo" />
                    OliveTin {{ currentVersion }}
                </p>
                <p>
                    <span>
                        <a href="https://docs.olivetin.app" target="_new">{{ t('docs') }}</a>
                    </span>

                    <span>
                        <a href="https://github.com/OliveTin/OliveTin/issues/new/choose" target="_new">{{ t('raise-issue') }}</a>
                    </span>

                    <span>{{ t('connected') }}</span>

                    <span>
                        <a href="#" @click.prevent="openLanguageDialog">{{ currentLanguageName }}</a>
                    </span>
                </p>
                <p>
                    <a id="available-version" href="http://olivetin.app" target="_blank" hidden>?</a>
                </p>
            </footer>
        </div>
    </div>

    <dialog ref="languageDialog" class="language-dialog" @click="handleDialogClick">
        <div class="dialog-content" @click.stop>
            <h2>{{ t('language-dialog.title') }}</h2>
            <select v-model="selectedLanguage" @change="changeLanguage" class="language-select">
                <option v-for="(name, code) in availableLanguages" :key="code" :value="code">
                    {{ name }}
                </option>
            </select>
            <p class="browser-languages">
                {{ t('language-dialog.browser-languages') }}: 
                <span v-if="browserLanguages.length > 0">{{ browserLanguages.join(', ') }}</span>
                <span v-else>{{ t('language-dialog.not-available') }}</span>
            </p>
            <div class="dialog-buttons">
                <button @click="closeLanguageDialog">{{ t('language-dialog.close') }}</button>
            </div>
        </div>
    </dialog>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue';
import { useRouter } from 'vue-router';
import Sidebar from 'picocrank/vue/components/Sidebar.vue';
import Header from 'picocrank/vue/components/Header.vue';
import { HugeiconsIcon } from '@hugeicons/vue'
import { Menu01Icon } from '@hugeicons/core-free-icons'
import { UserCircle02Icon } from '@hugeicons/core-free-icons'
import { DashboardSquare01Icon } from '@hugeicons/core-free-icons'
import logoUrl from '../../OliveTinLogo.png';
import { useI18n } from 'vue-i18n';

const { t, locale } = useI18n();

const router = useRouter();

const sidebar = ref(null);
const username = ref('notset');
const isLoggedIn = ref(false);
const serverConnection = ref(true);
const currentVersion = ref('?');
const bannerMessage = ref('');
const bannerCss = ref('');
const hasLoaded = ref(false);
const showFooter = ref(true)
const showNavigation = ref(true)
const showLogs = ref(true)
const showDiagnostics = ref(true)
const showLoginLink = ref(true)

const languageDialog = ref(null)
const browserLanguages = ref([])

const initialLanguagePreference = typeof window !== 'undefined' ? localStorage.getItem('olivetin-language') : null
const languagePreference = ref(initialLanguagePreference || 'auto')
const selectedLanguage = ref(languagePreference.value)

// Available languages with display names
const availableLanguages = {
    'auto': 'Browser Language',
    'en': 'English',
    'de-DE': 'Deutsch',
    'es-ES': 'Español',
    'it-IT': 'Italiano',
    'zh-Hans-CN': '简体中文'
}

// Computed property to get current language display name
const currentLanguageName = computed(() => {
    if (languagePreference.value === 'auto') {
        return availableLanguages['auto']
    }

    return availableLanguages[languagePreference.value] || languagePreference.value
})

function getBrowserLanguage() {
    if (navigator.languages && navigator.languages.length > 0) {
        return navigator.languages[0]
    }

    if (navigator.language) {
        return navigator.language
    }

    return 'en'
}

function toggleSidebar() {
    if (sidebar.value && showNavigation.value) {
        sidebar.value.toggle()
    }
}

function updateHeaderFromInit() {
    if (!window.initResponse) {
        return
    }

    username.value = window.initResponse.authenticatedUser
    isLoggedIn.value = window.initResponse.authenticatedUser !== '' && window.initResponse.authenticatedUser !== 'guest'
    currentVersion.value = window.initResponse.currentVersion
    bannerMessage.value = window.initResponse.bannerMessage || ''
    bannerCss.value = window.initResponse.bannerCss || ''
    showFooter.value = window.initResponse.showFooter
    showNavigation.value = window.initResponse.showNavigation
    showLogs.value = window.initResponse.showLogList
    showDiagnostics.value = window.initResponse.showDiagnostics

    if (!window.initResponse.authLocalLogin && window.initResponse.oAuth2Providers.length === 0) {
        showLoginLink.value = false
    }

    renderSidebar()

    if (window.checkWebsocketConnection) {
        window.checkWebsocketConnection()
    }

    if (window.initResponse.loginRequired) {
        router.push('/login')
        return
    }
}

function renderSidebar() {
    if (!sidebar.value) {
        return
    }

    if (typeof sidebar.value.clear === 'function') {
        sidebar.value.clear()
    }

    for (const rootDashboard of window.initResponse.rootDashboards) {
        sidebar.value.addNavigationLink({
            id: rootDashboard,
            name: rootDashboard,
            title: rootDashboard,
            path: rootDashboard === 'Actions' ? '/' : `/dashboards/${rootDashboard}`,
            icon: DashboardSquare01Icon,
        })
    }

    sidebar.value.addSeparator()
    sidebar.value.addRouterLink('Entities', t('nav.entities'))

    if (showLogs.value) {
        sidebar.value.addRouterLink('Logs', t('nav.logs'))
    }

    if (showDiagnostics.value) {
        sidebar.value.addRouterLink('Diagnostics', t('nav.diagnostics'))
    }
}

function openLanguageDialog() {
    selectedLanguage.value = languagePreference.value
    
    if (typeof navigator !== 'undefined' && Array.isArray(navigator.languages)) {
        browserLanguages.value = navigator.languages
    } else {
        browserLanguages.value = []
    }

    if (languageDialog.value) {
        languageDialog.value.showModal()
    }
}

function closeLanguageDialog() {
    if (languageDialog.value) {
        languageDialog.value.close()
    }
}

function changeLanguage() {
    if (!window.i18n || !selectedLanguage.value) {
        return
    }

    if (selectedLanguage.value === 'auto') {
        localStorage.removeItem('olivetin-language')
        languagePreference.value = 'auto'
        window.i18n.locale.value = getBrowserLanguage()
    } else {
        window.i18n.locale.value = selectedLanguage.value
        localStorage.setItem('olivetin-language', selectedLanguage.value)
        languagePreference.value = selectedLanguage.value
    }

    // Update sidebar with new translations
    if (sidebar.value) {
        renderSidebar()
    }

    closeLanguageDialog()
}

function handleDialogClick(event) {
    // Close dialog when clicking on the backdrop
    if (event.target === languageDialog.value) {
        closeLanguageDialog()
    }
}

window.updateHeaderFromInit = updateHeaderFromInit

onMounted(() => {
    serverConnection.value = true;
    updateHeaderFromInit()
    
    // Initialize selected language from stored preference
    selectedLanguage.value = languagePreference.value

    if (typeof navigator !== 'undefined' && Array.isArray(navigator.languages)) {
        browserLanguages.value = navigator.languages
    }
})
</script>

<style scoped>
.user-info span {
    margin-left: 1em;
}

.user-link {
    text-decoration: none;
    color: inherit;
}

.user-link:hover {
    text-decoration: underline;
}

.language-dialog {
    border: 1px solid var(--border-color, #ccc);
    border-radius: 0.5rem;
    padding: 0;
    max-width: 400px;
    width: 90%;
}

.language-dialog::backdrop {
    background-color: rgba(0, 0, 0, 0.5);
}

.dialog-content {
    padding: 1.5rem;
}

.dialog-content h2 {
    margin-top: 0;
    margin-bottom: 1rem;
}

.language-select {
    width: 100%;
    padding: 0.5rem;
    margin-bottom: 1rem;
    font-size: 1rem;
    border: 1px solid var(--border-color, #ccc);
    border-radius: 0.25rem;
}

.dialog-buttons {
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
}

.dialog-buttons button {
    padding: 0.5rem 1rem;
    cursor: pointer;
}

.browser-languages {
    font-size: 0.875rem;
    color: var(--fg2, #555);
    margin-bottom: 1rem;
}
</style>