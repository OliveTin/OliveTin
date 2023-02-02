import './ActionButton.js' // To define action-button

export function marshalActionButtonsJsonToHtml (json) {
  const currentIterationTimestamp = Date.now()

  for (const jsonButton of json.actions) {
    let htmlButton = document.querySelector('#actionButton_' + jsonButton.id)

    if (htmlButton == null) {
      htmlButton = document.createElement('action-button')
      htmlButton.constructFromJson(jsonButton)

      document.getElementById('root-group').appendChild(htmlButton)
    } else {
      htmlButton.updateFromJson(jsonButton)
      htmlButton.updateDom()
    }

    htmlButton.updateIterationTimestamp = currentIterationTimestamp
  }

  // Remove existing, but stale buttons (that were not updated in this round)
  for (const existingButton of document.querySelector('#contentActions').querySelectorAll('action-button')) {
    if (existingButton.updateIterationTimestamp !== currentIterationTimestamp) {
      existingButton.remove()
    }
  }
}

export function marshalLogsJsonToHtml (json) {
  for (const logEntry of json.logs) {
    const tpl = document.getElementById('tplLogRow')
    const row = tpl.content.cloneNode(true)

    if (logEntry.stdout.length === 0) {
      logEntry.stdout = '(empty)'
    }

    if (logEntry.stderr.length === 0) {
      logEntry.stderr = '(empty)'
    }

    let logTableExitCode = logEntry.exitCode

    if (logEntry.exitCode === 0) {
      logTableExitCode = 'OK'
    }

    if (logEntry.timedOut) {
      logTableExitCode += ' (timed out)'
    }

    row.querySelector('.timestamp').innerText = logEntry.datetime
    row.querySelector('.content').innerText = logEntry.actionTitle
    row.querySelector('.icon').innerHTML = logEntry.actionIcon
    row.querySelector('pre.stdout').innerText = logEntry.stdout
    row.querySelector('pre.stderr').innerText = logEntry.stderr
    row.querySelector('.exit-code').innerText = logTableExitCode

    for (const tag of logEntry.tags) {
      const domTag = document.createElement('span')
      domTag.classList.add('tag')
      domTag.innerText = tag

      row.querySelector('.tags').append(domTag)
    }

    document.querySelector('#logTableBody').prepend(row)
  }
}
