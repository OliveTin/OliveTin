<template>
  <div class="logs-view">
    <div class="toolbar">
      <label class="input-with-icons">
        <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24">
          <path fill="currentColor" d="m19.6 21l-6.3-6.3q-.75.6-1.725.95T9.5 16q-2.725 0-4.612-1.888T3 9.5t1.888-4.612T9.5 3t4.613 1.888T16 9.5q0 1.1-.35 2.075T14.7 13.3l6.3 6.3zM9.5 14q1.875 0 3.188-1.312T14 9.5t-1.312-3.187T9.5 5T6.313 6.313T5 9.5t1.313 3.188T9.5 14"/>
        </svg>
        <input 
          placeholder="Search for action name" 
          v-model="searchText"
          @input="handleSearch"
        />
        <button 
          title="Clear search filter" 
          :disabled="!searchText"
          @click="clearSearch"
        >
          <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24">
            <path fill="currentColor" d="M19 6.41L17.59 5L12 10.59L6.41 5L5 6.41L10.59 12L5 17.59L6.41 19L12 13.41L17.59 19L19 17.59L13.41 12z"/>
          </svg>
        </button>
      </label>
    </div>

    <table v-show="filteredLogs.length > 0" class="logs-table">
      <thead>
        <tr>
          <th>Timestamp</th>
          <th>Action</th>
          <th>Metadata</th>
          <th>Status</th>
        </tr>
      </thead>
      <tbody>
        <tr 
          v-for="log in filteredLogs" 
          :key="log.executionTrackingId"
          class="log-row"
          :title="log.actionTitle"
        >
          <td class="timestamp">{{ formatTimestamp(log.datetimeStarted) }}</td>
          <td>
            <span class="icon" v-html="log.actionIcon"></span>
            <a href="javascript:void(0)" class="content" @click="showLogDetails(log)">
              {{ log.actionTitle }}
            </a>
          </td>
          <td class="tags">
            <span v-if="log.tags && log.tags.length > 0" class="tag-list">
              <span v-for="tag in log.tags" :key="tag" class="tag">{{ tag }}</span>
            </span>
          </td>
          <td class="exit-code">
            <span :class="getStatusClass(log)">
              {{ getStatusText(log) }}
            </span>
          </td>
        </tr>
      </tbody>
    </table>

    <div v-show="filteredLogs.length === 0" class="empty-state">
      <p>There are no logs to display.</p>
      <router-link to="/">Return to index</router-link>
    </div>

    <p class="note">
      <strong>Note:</strong> The server is configured to only send 
      <strong>{{ pageSize }}</strong> log entries at a time. 
      The search box at the top of this page only searches this current page of logs.
    </p>
  </div>
</template>

<script>
export default {
  name: 'LogsView',
  data() {
    return {
      logs: [],
      searchText: '',
      pageSize: '?',
      loading: false
    }
  },
  computed: {
    filteredLogs() {
      if (!this.searchText) {
        return this.logs
      }
      
      const searchLower = this.searchText.toLowerCase()
      return this.logs.filter(log => 
        log.actionTitle.toLowerCase().includes(searchLower)
      )
    }
  },
  mounted() {
    this.fetchLogs()
    this.fetchPageSize()
  },
  methods: {
    async fetchLogs() {
      this.loading = true
      try {
        const response = await window.client.getLogs()
        this.logs = response.logEntries || []
      } catch (err) {
        console.error('Failed to fetch logs:', err)
        window.showBigError('fetch-logs', 'getting logs', err, false)
      } finally {
        this.loading = false
      }
    },
    
    async fetchPageSize() {
      try {
        const response = await fetch('webUiSettings.json')
        const settings = await response.json()
        this.pageSize = settings.LogsPageSize || '?'
      } catch (err) {
        console.warn('Failed to fetch page size:', err)
      }
    },
    
    handleSearch() {
      // Search is handled by computed property
    },
    
    clearSearch() {
      this.searchText = ''
    },
    
    formatTimestamp(timestamp) {
      if (!timestamp) return 'Unknown'
      
      try {
        const date = new Date(timestamp)
        return date.toLocaleString()
      } catch (err) {
        return timestamp
      }
    },
    
    getStatusClass(log) {
      if (log.timedOut) return 'status-timeout'
      if (log.blocked) return 'status-blocked'
      if (log.exitCode !== 0) return 'status-error'
      return 'status-success'
    },
    
    getStatusText(log) {
      if (log.timedOut) return 'Timed out'
      if (log.blocked) return 'Blocked'
      if (log.exitCode !== 0) return `Exit code ${log.exitCode}`
      return 'Success'
    },
    
    showLogDetails(log) {
      // Emit event to parent or use global execution dialog
      if (window.executionDialog) {
        window.executionDialog.reset()
        window.executionDialog.show()
        window.executionDialog.fetchExecutionResult(log.executionTrackingId)
      }
    }
  }
}
</script>

<style scoped>
.logs-view {
  padding: 1rem;
}

.toolbar {
  margin-bottom: 1rem;
}

.input-with-icons {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background: #fff;
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 0.5rem;
}

.input-with-icons input {
  border: none;
  outline: none;
  flex: 1;
  font-size: 1rem;
}

.input-with-icons button {
  background: none;
  border: none;
  cursor: pointer;
  padding: 0.25rem;
  border-radius: 3px;
}

.input-with-icons button:hover:not(:disabled) {
  background: #f5f5f5;
}

.input-with-icons button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.logs-table {
  width: 100%;
  border-collapse: collapse;
  background: #fff;
  border-radius: 4px;
  overflow: hidden;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.logs-table th {
  background: #f8f9fa;
  padding: 0.75rem;
  text-align: left;
  font-weight: 600;
  border-bottom: 1px solid #dee2e6;
}

.logs-table td {
  padding: 0.75rem;
  border-bottom: 1px solid #f1f3f4;
}

.logs-table tr:hover {
  background: #f8f9fa;
}

.timestamp {
  font-family: monospace;
  font-size: 0.875rem;
  color: #666;
}

.icon {
  margin-right: 0.5rem;
  font-size: 1.2em;
}

.content {
  color: #007bff;
  text-decoration: none;
  cursor: pointer;
}

.content:hover {
  text-decoration: underline;
}

.tag-list {
  display: flex;
  gap: 0.25rem;
  flex-wrap: wrap;
}

.tag {
  background: #e9ecef;
  color: #495057;
  padding: 0.125rem 0.5rem;
  border-radius: 12px;
  font-size: 0.75rem;
}

.status-success {
  color: #28a745;
  font-weight: 500;
}

.status-error {
  color: #dc3545;
  font-weight: 500;
}

.status-timeout {
  color: #ffc107;
  font-weight: 500;
}

.status-blocked {
  color: #6c757d;
  font-weight: 500;
}

.empty-state {
  text-align: center;
  padding: 2rem;
  color: #666;
}

.empty-state a {
  color: #007bff;
  text-decoration: none;
}

.empty-state a:hover {
  text-decoration: underline;
}

.note {
  margin-top: 1rem;
  padding: 1rem;
  background: #f8f9fa;
  border-radius: 4px;
  font-size: 0.875rem;
  color: #666;
}
</style> 