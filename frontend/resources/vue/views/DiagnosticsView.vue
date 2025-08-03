<template>
  <div class="diagnostics-view">
    <div class="diagnostics-content">
      <p class="note">
        <strong>Note:</strong> Diagnostics are only generated on OliveTin startup - they are not updated in real-time or
        when you refresh this page.
        They are intended as a "quick reference" to help you.
      </p>

      <p class="note">
        If you are having problems with OliveTin and want to raise a support request, please don't take a screenshot or
        copy text from this page,
        but instead it is highly recommended to include a
        <a href="https://docs.olivetin.app/sosreport.html" target="_blank">sosreport</a>
        which is more detailed, and makes it easier to help you.
      </p>

      <div class="diagnostics-section">
        <h3>SSH</h3>
        <table class="diagnostics-table">
          <tbody>
            <tr>
              <td width="10%">Found Key</td>
              <td>{{ diagnostics.sshFoundKey || '?' }}</td>
            </tr>
            <tr>
              <td>Found Config</td>
              <td>{{ diagnostics.sshFoundConfig || '?' }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div v-if="diagnostics.system" class="diagnostics-section">
        <h3>System</h3>
        <table class="diagnostics-table">
          <tbody>
            <tr v-for="(value, key) in diagnostics.system" :key="key">
              <td width="10%">{{ formatKey(key) }}</td>
              <td>{{ value }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div v-if="diagnostics.network" class="diagnostics-section">
        <h3>Network</h3>
        <table class="diagnostics-table">
          <tbody>
            <tr v-for="(value, key) in diagnostics.network" :key="key">
              <td width="10%">{{ formatKey(key) }}</td>
              <td>{{ value }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div v-if="diagnostics.storage" class="diagnostics-section">
        <h3>Storage</h3>
        <table class="diagnostics-table">
          <tbody>
            <tr v-for="(value, key) in diagnostics.storage" :key="key">
              <td width="10%">{{ formatKey(key) }}</td>
              <td>{{ value }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div v-if="diagnostics.services" class="diagnostics-section">
        <h3>Services</h3>
        <table class="diagnostics-table">
          <tbody>
            <tr v-for="(value, key) in diagnostics.services" :key="key">
              <td width="10%">{{ formatKey(key) }}</td>
              <td>{{ value }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div v-if="diagnostics.errors && diagnostics.errors.length > 0" class="diagnostics-section">
        <h3>Errors</h3>
        <div class="error-list">
          <div v-for="(error, index) in diagnostics.errors" :key="index" class="error-item">
            {{ error }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'

const diagnostics = ref({})
const loading = ref(false)

async function fetchDiagnostics() {
  loading.value = true

  try {
    const response = await window.client.getDiagnostics();
    diagnostics.value = {
      sshFoundKey: response.sshFoundKey,
      sshFoundConfig: response.sshFoundConfig
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

.diagnostics-section {
  margin-bottom: 2rem;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.diagnostics-section h3 {
  margin: 0;
  padding: 1rem;
  background: #f8f9fa;
  border-bottom: 1px solid #dee2e6;
  font-size: 1.1rem;
  font-weight: 600;
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
</style>