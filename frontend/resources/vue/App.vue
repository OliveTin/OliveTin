<template>
    <header>
        <img src="../../OliveTinLogo.png" alt="OliveTin logo" class="logo" />

        <h1 id="page-title">
            <router-link to="/">OliveTin</router-link>
        </h1>

        <button id="sidebar-toggler-button" aria-label="Open sidebar navigation" aria-pressed="false"
            aria-haspopup="menu" @click="toggleSidebar">&#9776;</button>


        <div class="fg1" />

        <div class="userinfo">
            <span id="link-login" hidden><router-link to="/login">Login</router-link> |</span>
            <span id="link-logout" hidden><a href="/api/Logout">Logout</a> |</span>
            <span id="username">&nbsp;</span>
            <svg xmlns="http://www.w3.org/2000/svg" width="1.5em" height="1.5em" viewBox="0 0 24 24">
                <g fill="none" fill-rule="evenodd">
                    <path
                        d="m12.594 23.258l-.012.002l-.071.035l-.02.004l-.014-.004l-.071-.036q-.016-.004-.024.006l-.004.01l-.017.428l.005.02l.01.013l.104.074l.015.004l.012-.004l.104-.074l.012-.016l.004-.017l-.017-.427q-.004-.016-.016-.018m.264-.113l-.014.002l-.184.093l-.01.01l-.003.011l.018.43l.005.012l.008.008l.201.092q.019.005.029-.008l.004-.014l-.034-.614q-.005-.019-.02-.022m-.715.002a.02.02 0 0 0-.027.006l-.006.014l-.034.614q.001.018.017.024l.015-.002l.201-.093l.01-.008l.003-.011l.018-.43l-.003-.012l-.01-.01z" />
                    <path fill="currentColor"
                        d="M12 2C6.477 2 2 6.477 2 12s4.477 10 10 10s10-4.477 10-10S17.523 2 12 2M8.5 9.5a3.5 3.5 0 1 1 7 0a3.5 3.5 0 0 1-7 0m9.758 7.484A7.99 7.99 0 0 1 12 20a7.99 7.99 0 0 1-6.258-3.016C7.363 15.821 9.575 15 12 15s4.637.821 6.258 1.984" />
                </g>
            </svg>
        </div>
    </header>

    <div id="layout">
        <Sidebar ref="sidebar" />

        <div id="content">
            <main title="Main content">
                <router-view />
            </main>

            <ExecutionDialog ref="executionDialog" />

            <footer title="footer">
                <p><img title="application icon" src="../../OliveTinLogo.png" alt="OliveTin logo" height="1em"
                        class="logo" />
                    OliveTin</p>
                <p>
                    <span>
                        <a href="https://docs.olivetin.app" target="_new">Documentation</a>
                    </span>

                    <span>
                        <a href="https://github.com/OliveTin/OliveTin/issues/new/choose" target="_new">Raise an issue on
                            GitHub</a>
                    </span>

                    <span id="currentVersion">?</span>

                    <span id="serverConnectionRest">REST</span>
                    <span id="serverConnectionWebSocket">WebSocket</span>
                </p>
                <p>
                    <a id="available-version" href="http://olivetin.app" target="_blank" hidden>?</a>
                </p>
            </footer>
        </div>
    </div>
</template>

<script setup>
import { ref } from 'vue';
import Sidebar from './components/Sidebar.vue';
import ExecutionDialog from './ExecutionDialog.vue';

import { provide } from 'vue';
const sidebar = ref(null);
const executionDialog = ref(null);

provide('executionDialog', executionDialog.value);

function toggleSidebar() {
    if (sidebar.value && typeof sidebar.value.isOpen !== 'undefined') {
        sidebar.value.isOpen = !sidebar.value.isOpen;
    }
}
</script>