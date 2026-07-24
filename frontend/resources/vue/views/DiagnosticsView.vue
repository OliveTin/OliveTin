<template>
  <Section :title="t('diagnostics.config-issues')">
    <p>{{ t('diagnostics.config-issues-description') }}</p>

    <p v-if="!loading && configIssues.length === 0">
      {{ t('diagnostics.config-issues-none') }}
    </p>

    <Table
      v-else
      :data="configIssueRows"
      :headers="configIssueHeaders"
      :show-pagination="false"
    >
      <template #cell-severity="{ value }">
        <div
          class="tag"
          :class="value === 'error' ? 'fg-bad' : 'fg-warning'"
        >
          {{ value }}
        </div>
      </template>
    </Table>
  </Section>

  <Section :title="t('diagnostics.get-support')">
    <p>
      {{ t('diagnostics.get-support-description') }}
    </p>
    <ul>
      <li>
        <a
          href="https://docs.olivetin.app/troubleshooting/wheretofindhelp.html"
          target="_blank"
        >{{ t('diagnostics.where-to-find-help') }}</a>
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

  <Section :title="t('diagnostics.server-diagnostics')">
    <p>{{ t('diagnostics.server-diagnostics-description') }}</p>
    <p>
      <a
        href="https://docs.olivetin.app/troubleshooting/server-diagnostics.html"
        target="_blank"
      >{{ t('diagnostics.server-diagnostics-docs') }}</a>
    </p>

    <div role="toolbar">
      <button
        :disabled="loading"
        class="good"
        @click="generateServerDiagnostics"
      >
        {{ t('diagnostics.generate-server-diagnostics') }}
      </button>
      <button
        :disabled="!serverDiagnostics || loading"
        :class="serverDiagnosticsCopied ? 'good' : ''"
        @click="copyServerDiagnostics"
      >
        {{ serverDiagnosticsCopied ? t('diagnostics.copied') : t('diagnostics.copy-to-clipboard') }}
      </button>
    </div>

    <textarea
      v-model="serverDiagnostics"
      readonly
      style="flex: 1; min-height: 200px; resize: vertical; width: 100%; box-sizing: border-box;"
    />
  </Section>

  <Section :title="t('diagnostics.browser-info')">
    <p>{{ t('diagnostics.browser-info-description') }}</p>

    <div role="toolbar">
      <button
        :disabled="loading"
        class="good"
        @click="generateBrowserInfo"
      >
        {{ t('diagnostics.generate-browser-info') }}
      </button>
      <button
        :disabled="!browserInfo || loading"
        :class="browserInfoCopied ? 'good' : ''"
        @click="copyBrowserInfo"
      >
        {{ browserInfoCopied ? t('diagnostics.copied') : t('diagnostics.copy-to-clipboard') }}
      </button>
    </div>

    <textarea
      v-model="browserInfo"
      readonly
      style="flex: 1; min-height: 200px; resize: vertical; width: 100%; box-sizing: border-box;"
    />
  </Section>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import Section from 'picocrank/vue/components/Section.vue'
import Table from 'picocrank/vue/components/Table.vue'
import { useI18n } from 'vue-i18n'

const { t, locale } = useI18n()

const diagnostics = ref({})
const configIssues = ref([])
const loading = ref(false)
const serverDiagnostics = ref('')
const browserInfo = ref('')
const serverDiagnosticsCopied = ref(false)
const browserInfoCopied = ref(false)

const configIssueHeaders = computed(() => [
  { key: 'severity', label: t('diagnostics.config-issue-severity'), sortable: true, width: '7rem' },
  { key: 'code', label: t('diagnostics.config-issue-code'), sortable: true, width: '12rem' },
  { key: 'message', label: t('diagnostics.config-issue-message'), sortable: false },
  { key: 'actionTitle', label: t('diagnostics.config-issue-action'), sortable: true, width: '10rem' },
  { key: 'argumentName', label: t('diagnostics.config-issue-argument'), sortable: true, width: '8rem' },
  { key: 'configFile', label: t('diagnostics.config-issue-config-file'), sortable: true, width: '14rem' },
  { key: 'source', label: t('diagnostics.config-issue-detail'), sortable: false, width: '12rem' }
])

const configIssueRows = computed(() => configIssues.value.map((issue) => ({
  severity: issue.severity || '',
  code: issue.code || '',
  message: issue.message || '',
  actionTitle: issue.actionTitle || '',
  argumentName: issue.argumentName || '',
  configFile: issue.configFile || '',
  source: issue.source || ''
})))

async function fetchDiagnostics () {
  loading.value = true

  try {
    const response = await window.client.getDiagnostics()
    diagnostics.value = {
      sshFoundKey: response.SshFoundKey,
      sshFoundConfig: response.SshFoundConfig
    }
    configIssues.value = response.configIssues || []
  } catch (err) {
    console.error('Failed to fetch diagnostics:', err)
    diagnostics.value = {
      sshFoundKey: t('diagnostics.unknown'),
      sshFoundConfig: t('diagnostics.unknown')
    }
    configIssues.value = []
  }
  loading.value = false
}

async function generateServerDiagnostics () {
  loading.value = true

  try {
    const response = await window.client.serverDiagnostics()
    console.log('response', response)
    serverDiagnostics.value = `\`\`\`\n${response.alert}\n\`\`\`\n`
  } catch (err) {
    console.error('Failed to generate server diagnostics:', err)
    serverDiagnostics.value = ''
  } finally {
    loading.value = false
  }
}

async function copyServerDiagnostics () {
  try {
    await navigator.clipboard.writeText(serverDiagnostics.value)
    serverDiagnosticsCopied.value = true
    setTimeout(() => {
      serverDiagnosticsCopied.value = false
    }, 2000)
  } catch (err) {
    console.error('Failed to copy Server Diagnostics to clipboard:', err)
  }
}

async function generateBrowserInfo () {
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
      userAgentData
    }

    const showVersionNumber = window.initResponse?.effectivePolicy?.showVersionNumber ?? true
    const olivetinVersion = showVersionNumber
      ? (window.initResponse?.currentVersion || t('diagnostics.unknown'))
      : '[hidden]'
    const currentLanguage = locale.value || t('diagnostics.unknown')

    let output = ''
    output += '```\n'
    output += '### BROWSER INFO START (copy all text to BROWSER INFO END)\n'
    output += '# OliveTin Information\n'
    output += `olivetinVersion: ${olivetinVersion}\n`
    output += `currentLanguage: ${currentLanguage}\n`
    output += '\n# Browser Information\n'
    output += `userAgent: ${info.userAgent}\n`
    output += `platform: ${info.platform}\n`
    output += `language: ${info.language}\n`
    output += `languages: ${info.languages}\n`
    output += '\n# User Agent Data\n'
    output += `userAgentData:\n${info.userAgentData}\n`
    output += '\n# Display Information\n'
    output += `screenWidth: ${info.screenWidth}\n`
    output += `screenHeight: ${info.screenHeight}\n`
    output += `screenColorDepth: ${info.screenColorDepth}\n`
    output += `screenPixelDepth: ${info.screenPixelDepth}\n`
    output += `viewportWidth: ${info.viewportWidth}\n`
    output += `viewportHeight: ${info.viewportHeight}\n`
    output += `devicePixelRatio: ${info.devicePixelRatio}\n`
    output += '\n# Feature Support\n'
    output += `cookieEnabled: ${info.cookieEnabled}\n`
    output += `localStorageEnabled: ${info.localStorageEnabled}\n`
    output += `sessionStorageEnabled: ${info.sessionStorageEnabled}\n`
    output += `onLine: ${info.onLine}\n`
    output += `hardwareConcurrency: ${info.hardwareConcurrency}\n`
    output += `maxTouchPoints: ${info.maxTouchPoints}\n`
    output += '\n# Location & Time\n'
    output += `timezone: ${info.timezone}\n`
    output += `timezoneOffset: ${info.timezoneOffset}\n`
    output += '\n### BROWSER INFO END (copy all text from BROWSER INFO START)'
    output += '\n```\n'

    browserInfo.value = output
  } finally {
    loading.value = false
  }
}

async function copyBrowserInfo () {
  try {
    await navigator.clipboard.writeText(browserInfo.value)
    browserInfoCopied.value = true
    setTimeout(() => {
      browserInfoCopied.value = false
    }, 2000)
  } catch (err) {
    console.error('Failed to copy browser info to clipboard:', err)
  }
}

onMounted(() => {
  fetchDiagnostics()
})
</script>
