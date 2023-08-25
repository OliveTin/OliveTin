import { marshalLogsJsonToHtml } from './marshaller.js'

window.ws = null

export function checkWebsocketConnection () {
  if (window.ws === null || window.ws.readyState === 3) {
    reconnectWebsocket()
  }
}

function reconnectWebsocket () {
  window.websocketAvailable = false

  let proto = 'ws:'

  if (window.location.protocol === 'https:') {
    proto = 'wss:'
  }

  const websocketConnectionUrl = proto + window.location.host + '/websocket'
  const ws = window.ws = new WebSocket(websocketConnectionUrl)

  ws.addEventListener('open', websocketOnOpen)
  ws.addEventListener('message', websocketOnMessage)
  ws.addEventListener('error', websocketOnError)
  ws.addEventListener('close', websocketOnClose)
}

function websocketOnOpen (evt) {
  window.websocketAvailable = true

  window.ws.send('monitor')

  window.refreshLoop()
}

function websocketOnMessage (msg) {
  // FIXME check msg status is OK
  const j = JSON.parse(msg.data)

  switch (j.type) {
    case 'ExecutionFinished':
      updatePageAfterFinished(j.payload)
      break
    default:
      window.showBigError('Unknown message type from server: ' + j.type)
  }
}

function updatePageAfterFinished (logEntry) {
  document.querySelector('execution-button#execution-' + logEntry.uuid).onFinished(logEntry)

  marshalLogsJsonToHtml({
    logs: [logEntry]
  })

  // If the current execution dialog is open, update that too
  if (window.executionDialog != null && window.executionDialog.dlg.open && window.executionDialog.executionUuid === logEntry.uuid) {
    window.executionDialog.renderResult({
      logEntry: logEntry
    })
  }
}

function websocketOnError (err) {
  window.websocketAvailable = false
  window.refreshLoop()
  console.error(err)
}

function websocketOnClose () {
  window.websocketAvailable = false
}
