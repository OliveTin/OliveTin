<template>
  <Section :title="t('diagnostics.get-support')">
    <p>{{ t('diagnostics.get-support-description') }}
    </p>
    <ul>
      <li>
        <a href = "https://docs.olivetin.app/troubleshooting/wheretofindhelp.html" target="_blank">{{ t('diagnostics.where-to-find-help') }}</a>
      </li>
    </ul>
  </Section>

  <Section :title="t('diagnostics.ssh')">
    <dl>
      <dt>{{ t('diagnostics.found-key') }}</dt>
      <dd>{{ diagnostics.sshFoundKey || '?' }}</dd>
      <dt>{{ t('diagnostics.found-config') }}</dt>
      <dd>{{ diagnostics.sshFoundConfig || '?' }}</dd>
    </dl>
  </Section>

  <Section :title="t('diagnostics.sos-report')">
    <p>{{ t('diagnostics.sos-report-description') }}</p>
    <p>
      <a href="https://docs.olivetin.app/troubleshooting/sosreport.html" target="_blank">{{ t('diagnostics.sos-report-docs') }}</a>
    </p>

    <div role="toolbar">
      <button @click="generateSosReport" :disabled="loading" class = "good">{{ t('diagnostics.generate-sos-report') }}</button>
    </div>

    <textarea v-model="sosReport" readonly style="flex: 1; min-height: 200px; resize: vertical; width: 100%; box-sizing: border-box;"></textarea>
  </Section>

  <Section :title="t('diagnostics.browser-info')">
    <p>{{ t('diagnostics.browser-info-description') }}</p>

    <div role="toolbar">
      <button @click="generateBrowserInfo" :disabled="loading" class = "good">{{ t('diagnostics.generate-browser-info') }}</button>
    </div>

    <textarea v-model="browserInfo" readonly style="flex: 1; min-height: 200px; resize: vertical; width: 100%; box-sizing: border-box;"></textarea>
  </Section>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import Section from 'picocrank/vue/components/Section.vue'
import { useI18n } from 'vue-i18n'

const { t, locale } = useI18n()

const diagnostics = ref({})
const loading = ref(false)
const sosReport = ref('')
const browserInfo = ref('')

async function fetchDiagnostics() {
  loading.value = true

  try {
    const response = await window.client.getDiagnostics();
    diagnostics.value = {
      sshFoundKey: response.SshFoundKey,
      sshFoundConfig: response.SshFoundConfig
    };
  } catch (err) {
    console.error('Failed to fetch diagnostics:', err);
    diagnostics.value = {
      sshFoundKey: t('diagnostics.unknown'),
      sshFoundConfig: t('diagnostics.unknown')
    }
  }
  loading.value = false
}

function formatKey(key) {
  return key
    .replace(/([A-Z])/g, ' $1')
    .replace(/^./, str => str.toUpperCase())
    .trim()
}

async function generateSosReport() {
  const response = await window.client.sosReport()
  console.log("response", response)
  sosReport.value = response.alert
}

async function generateBrowserInfo() {
  loading.value = true
  try {
    let userAgentData = 'N/A'
    if (navigator.userAgentData) {
      try {
        const uaData = await navigator.userAgentData.getHighEntropyValues([
          'platform',
          'platformVersion',
          'architecture',
          'model',
          'uaFullVersion',
          'bitness',
          'fullVersionList'
        ])
        userAgentData = JSON.stringify(uaData, null, 2)
      } catch (err) {
        userAgentData = `${t('diagnostics.useragent-data-error')}: ${err.message}`
      }
    }

    const info = {
      userAgent: navigator.userAgent,
      platform: navigator.platform,
      language: navigator.language,
      languages: navigator.languages?.join(', ') || 'N/A',
      cookieEnabled: navigator.cookieEnabled,
      onLine: navigator.onLine,
      screenWidth: screen.width,
      screenHeight: screen.height,
      screenColorDepth: screen.colorDepth,
      screenPixelDepth: screen.pixelDepth,
      viewportWidth: window.innerWidth,
      viewportHeight: window.innerHeight,
      devicePixelRatio: window.devicePixelRatio || 'N/A',
      timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
      timezoneOffset: new Date().getTimezoneOffset(),
      localStorageEnabled: (() => {
        try {
          localStorage.setItem('test', 'test')
          localStorage.removeItem('test')
          return true
        } catch {
          return false
        }
      })(),
      sessionStorageEnabled: (() => {
        try {
          sessionStorage.setItem('test', 'test')
          sessionStorage.removeItem('test')
          return true
        } catch {
          return false
        }
      })(),
      hardwareConcurrency: navigator.hardwareConcurrency || 'N/A',
      maxTouchPoints: navigator.maxTouchPoints || 'N/A',
      vendor: navigator.vendor || 'N/A',
      appName: navigator.appName,
      appVersion: navigator.appVersion,
      product: navigator.product,
      userAgentData: userAgentData
    }

    const olivetinVersion = window.initResponse?.currentVersion || t('diagnostics.unknown')
    const currentLanguage = locale.value || t('diagnostics.unknown')

    let output = '### BROWSER INFO START (copy all text to BROWSER INFO END)\n'
    output += `# OliveTin Information\n`
    output += `olivetinVersion: ${olivetinVersion}\n`
    output += `currentLanguage: ${currentLanguage}\n`
    output += `\n# Browser Information\n`
    output += `userAgent: ${info.userAgent}\n`
    output += `platform: ${info.platform}\n`
    output += `language: ${info.language}\n`
    output += `languages: ${info.languages}\n`
    output += `vendor: ${info.vendor}\n`
    output += `appName: ${info.appName}\n`
    output += `appVersion: ${info.appVersion}\n`
    output += `product: ${info.product}\n`
    output += `\n# User Agent Data\n`
    output += `userAgentData:\n${info.userAgentData}\n`
    output += `\n# Display Information\n`
    output += `screenWidth: ${info.screenWidth}\n`
    output += `screenHeight: ${info.screenHeight}\n`
    output += `screenColorDepth: ${info.screenColorDepth}\n`
    output += `screenPixelDepth: ${info.screenPixelDepth}\n`
    output += `viewportWidth: ${info.viewportWidth}\n`
    output += `viewportHeight: ${info.viewportHeight}\n`
    output += `devicePixelRatio: ${info.devicePixelRatio}\n`
    output += `\n# Feature Support\n`
    output += `cookieEnabled: ${info.cookieEnabled}\n`
    output += `localStorageEnabled: ${info.localStorageEnabled}\n`
    output += `sessionStorageEnabled: ${info.sessionStorageEnabled}\n`
    output += `onLine: ${info.onLine}\n`
    output += `hardwareConcurrency: ${info.hardwareConcurrency}\n`
    output += `maxTouchPoints: ${info.maxTouchPoints}\n`
    output += `\n# Location & Time\n`
    output += `timezone: ${info.timezone}\n`
    output += `timezoneOffset: ${info.timezoneOffset}\n`
    output += `\n### BROWSER INFO END (copy all text from BROWSER INFO START)\n`

    browserInfo.value = output
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchDiagnostics()
})
</script>

<style scoped>
.diagnostics-view {
  padding: 1rem;
}

.diagnostics-content {
  max-width: 800px;
  margin: 0 auto;
}

.note {
  background: #f8f9fa;
  border-left: 4px solid #007bff;
  padding: 1rem;
  margin-bottom: 1rem;
  border-radius: 0 4px 4px 0;
  font-size: 0.875rem;
  color: #495057;
}

.note a {
  color: #007bff;
  text-decoration: none;
}

.note a:hover {
  text-decoration: underline;
}

.diagnostics-table {
  width: 100%;
  border-collapse: collapse;
}

.diagnostics-table td {
  padding: 0.75rem 1rem;
  border-bottom: 1px solid #f1f3f4;
}

.diagnostics-table td:first-child {
  font-weight: 500;
  color: #495057;
  background: #f8f9fa;
}

.diagnostics-table tr:last-child td {
  border-bottom: none;
}

.error-list {
  padding: 1rem;
}

.error-item {
  background: #f8d7da;
  color: #721c24;
  padding: 0.75rem;
  margin-bottom: 0.5rem;
  border-radius: 4px;
  border-left: 4px solid #dc3545;
  font-family: monospace;
  font-size: 0.875rem;
}

.error-item:last-child {
  margin-bottom: 0;
}

.flex-col {
  display: flex;
  flex-direction: column;
}

.section-content {
  display: flex;
  flex-direction: column;
  gap: 1em;
}
</style>