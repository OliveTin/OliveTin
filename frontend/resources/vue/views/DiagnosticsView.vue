<template>
  <Section title = "Get support">
    <p>If you are having problems with OliveTin and want to raise a support request, it would be very helpful to include a sosreport from this page.
    </p>
    <ul>
      <li>
        <a href="https://docs.olivetin.app/sosreport.html" target="_blank">sosreport Documentation</a>
      </li>
      <li>
        <a href = "https://docs.olivetin.app/troubleshooting/wheretofindhelp.html" target="_blank">Where to find help</a>
      </li>
    </ul>
  </Section>

  <Section title = "SSH">
    <dl>
      <dt>Found Key</dt>
      <dd>{{ diagnostics.sshFoundKey || '?' }}</dd>
      <dt>Found Config</dt>
      <dd>{{ diagnostics.sshFoundConfig || '?' }}</dd>
    </dl>
  </Section>

  <Section title = "SOS Report">
    <p>This section allows you to generate a detailed report of your configuration and environment. It is a good idea to include this when raising a support request.</p>

    <div role="toolbar">
      <button @click="generateSosReport" :disabled="loading" class = "good">Generate SOS Report</button>
    </div>

    <textarea v-model="sosReport" readonly style="flex: 1; min-height: 200px; resize: vertical;"></textarea>
  </Section>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import Section from 'picocrank/vue/components/Section.vue'

const diagnostics = ref({})
const loading = ref(false)
const sosReport = ref('Waiting to start...')

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
      sshFoundKey: 'Unknown',
      sshFoundConfig: 'Unknown'
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