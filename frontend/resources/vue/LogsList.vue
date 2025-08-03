<template>
    <section title="Logs" class="">
        <div class="toolbar">
            <label class="input-with-icons">
                <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24">
                    <path fill="currentColor"
                        d="m19.6 21l-6.3-6.3q-.75.6-1.725.95T9.5 16q-2.725 0-4.612-1.888T3 9.5t1.888-4.612T9.5 3t4.613 1.888T16 9.5q0 1.1-.35 2.075T14.7 13.3l6.3 6.3zM9.5 14q1.875 0 3.188-1.312T14 9.5t-1.312-3.187T9.5 5T6.313 6.313T5 9.5t1.313 3.188T9.5 14" />
                </svg>
                <input placeholder="Search for action name" id="logSearchBox" />
                <button id="searchLogsClear" title="Clear search filter" disabled>
                    <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24">
                        <path fill="currentColor"
                            d="M19 6.41L17.59 5L12 10.59L6.41 5L5 6.41L10.59 12L5 17.59L6.41 19L12 13.41L17.59 19L19 17.59L13.41 12z" />
                    </svg>
                </button>
            </label>
        </div>
        <table id="logsTable" title="Logs" hidden>
            <thead>
                <tr title="untitled">
                    <th>Timestamp</th>
                    <th>Action</th>
                    <th>Metadata</th>
                    <th>Status</th>
                </tr>
            </thead>
            <tbody>
                <tr v-for="logEntry in logEntries" :key="logEntry.executionTrackingId"></tr>
                    <td>{{ logEntry.datetimeStarted }}</td>
                    <td>{{ logEntry.actionTitle }}</td>
                    <td>{{ logEntry.actionIcon }}</td>
                    <td>{{ logEntry.tags }}</td>
                    <td>{{ logEntry.user }}</td>
                </tr>
            </tbody>
        </table>

        <p id="logsTableEmpty">There are no logs to display. <a href="/">Return to index</a></p>

        <p><strong>Note:</strong> The server is configured to only send <strong id="logs-server-page-size">?</strong>
            log entries at a time. The search box at the top of this page only searches this current page of logs.</p>
    </section>
</template>

<script setup>
import { onMounted } from 'vue'

function setupLogSearchBox () {
  document.getElementById('logSearchBox').oninput = searchLogs
  document.getElementById('searchLogsClear').onclick = searchLogsClear
}

function marshalLogsJsonToHtml (json) {
  // This function is called internally with a "fake" server response, that does
  // not have pageSize set. So we need to check if it's set before trying to use it.
  if (json.pageSize !== undefined) {
    document.getElementById('logs-server-page-size').innerText = json.pageSize
  }

  if (json.logs != null && json.logs.length > 0) {
    document.getElementById('logsTable').hidden = false
    document.getElementById('logsTableEmpty').hidden = true
  } else {
    return
  }

  for (const logEntry of json.logs) {
    let row = document.getElementById('log-' + logEntry.executionTrackingId)

    if (row == null) {
      const tpl = document.getElementById('tplLogRow')
      row = tpl.content.querySelector('tr').cloneNode(true)
      row.id = 'log-' + logEntry.executionTrackingId

      row.querySelector('.content').onclick = () => {
        window.executionDialog.reset()
        window.executionDialog.show()
        window.executionDialog.renderExecutionResult({
          logEntry: window.logEntries.get(logEntry.executionTrackingId)
        })
        pushNewNavigationPath('/logs/' + logEntry.executionTrackingId)
      }

      logEntry.dom = row

      window.logEntries.set(logEntry.executionTrackingId, logEntry)

      document.querySelector('#logTableBody').prepend(row)
    }

    row.querySelector('.timestamp').innerText = logEntry.datetimeStarted
    row.querySelector('.content').innerText = logEntry.actionTitle
    row.querySelector('.icon').innerHTML = logEntry.actionIcon
    row.setAttribute('title', logEntry.actionTitle)

    row.exitCodeDisplay.update(logEntry)

    row.querySelector('.tags').innerHTML = ''

    for (const tag of logEntry.tags) {
      row.querySelector('.tags').append(createTag(tag))
    }

    row.querySelector('.tags').append(createAnnotation('user', logEntry.user))
  }
}

onMounted(() => {
    setupLogSearchBox()
})
</script>