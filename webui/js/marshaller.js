import './ActionButton.js' // To define action-button

export function marshalActionButtonsJsonToHtml (json) {
  const currentIterationTimestamp = Date.now()

  for (const jsonButton of json.actions) {
    var htmlButton = document.querySelector('#actionButton_' + jsonButton.id)

    if (htmlButton == null) {
      htmlButton = document.createElement('button', { is: 'action-button' })
      htmlButton.constructFromJson(jsonButton)

      document.getElementById('rootGroup').appendChild(htmlButton)
    } else {
      htmlButton.updateFromJson(jsonButton)
      htmlButton.updateHtml()
    }

    htmlButton.updateIterationTimestamp = currentIterationTimestamp;
  }

  for (const existingButton of document.querySelector('#contentActions').querySelectorAll('button')) {
    if (existingButton.updateIterationTimestamp != currentIterationTimestamp) {
      existingButton.remove();
    }
  }
}

export function marshalLogsJsonToHtml (json) {
  for (const logEntry of json.logs) {
    const tpl = document.getElementById('tplLogRow')
    const row = tpl.content.cloneNode(true)

    row.querySelector('.timestamp').innerText = logEntry.datetime
    row.querySelector('.content').innerText = logEntry.actionTitle
    row.querySelector('pre').innerText = logEntry.stdout

    document.querySelector('#logTableBody').prepend(row)
  }
}
