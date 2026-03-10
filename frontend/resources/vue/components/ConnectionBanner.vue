<template>
    <span id="connection-banner" v-if="!connectionState.connected" class="inline-notification critical user-info-connection">
        <span class="connection-banner-sr-only" role="status">{{ staticAnnouncement }}</span>
        <span aria-hidden="true">
            <a :href="websocketDocsUrl" target="_blank" rel="noopener noreferrer" class="connection-banner-link">{{ linkText }}</a>{{ bannerSuffix }}
        </span>
    </span>
</template>

<script setup>
import { ref, computed, watch, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { connectionState } from '../stores/connectionState.js'

const { t } = useI18n()

const websocketDocsUrl = 'https://docs.olivetin.app/troubleshooting/err-websocket-connection.html'

const linkText = computed(() => t('disconnected-banner-link-text'))

function formatShortRelative(ms) {
  if (ms < 0) return '0s'
  const secs = Math.floor(ms / 1000)
  const mins = Math.floor(secs / 60)
  const hours = Math.floor(mins / 60)
  if (hours > 0) return `${hours}h`
  if (mins > 0) return `${mins}m`
  return `${secs}s`
}

function formatShortTime(ts) {
  if (ts == null) return '--:--'
  return new Date(ts).toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' })
}

const now = ref(Date.now())
let ticker = null
watch(() => connectionState.connected, (connected) => {
  if (ticker) {
    clearInterval(ticker)
    ticker = null
  }
  if (!connected) {
    now.value = Date.now()
    ticker = setInterval(() => { now.value = Date.now() }, 1000)
  }
}, { immediate: true })

onUnmounted(() => {
  if (ticker) {
    clearInterval(ticker)
    ticker = null
  }
})

const staticAnnouncement = computed(() => t('disconnected-banner-announcement'))

const bannerSuffix = computed(() => {
  const at = connectionState.disconnectedAt
  const next = connectionState.nextReconnectAt
  const n = now.value
  const disconnectedSince = formatShortTime(at)
  if (next != null && next > n) {
    const reconnectIn = formatShortRelative(next - n)
    return t('disconnected-banner-suffix', { disconnectedSince, reconnectIn })
  }
  return t('disconnected-banner-suffix-reconnecting', { disconnectedSince })
})
</script>

<style scoped>
#connection-banner.user-info-connection {
    font-weight: 500;
}
.inline-notification {
    border: 0;
    margin: 0;
}
.connection-banner-link {
    color: inherit;
    text-decoration: underline;
}
.connection-banner-sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
}
</style>
