import {
  refreshServerConnectionLabel
} from './marshaller.js'

window.ws = null

export function checkWebsocketConnection () {
  if (window.ws === null || window.ws.readyState === 3) {
    reconnectWebsocket()
  }
}

function reconnectWebsocket () {
  window.websocketAvailable = false

  const websocketConnectionUrl = new URL(window.location.toString())
  websocketConnectionUrl.hash = ''
  websocketConnectionUrl.pathname = '/websocket'

  if (window.location.protocol === 'https:') {
    websocketConnectionUrl.protocol = 'wss'
  } else {
    websocketConnectionUrl.protocol = 'ws'
  }

  window.websocketConnectionUrl = websocketConnectionUrl

  const ws = window.ws = new WebSocket(websocketConnectionUrl.toString())

  ws.addEventListener('open', websocketOnOpen)
  ws.addEventListener('message', websocketOnMessage)
  ws.addEventListener('error', websocketOnError)
  ws.addEventListener('close', websocketOnClose)
}

function websocketOnOpen (evt) {
  window.websocketAvailable = true

  window.ws.send('monitor')

  refreshServerConnectionLabel()

  window.refreshLoop()
}

function websocketOnMessage (msg) {
  // FIXME check msg status is OK
  const j = JSON.parse(msg.data)

  const e = new Event(j.type)
  e.payload = j.payload

  switch (j.type) {
    case 'EventOutputChunk':
    case 'EventConfigChanged':
    case 'EventExecutionFinished':
    case 'EventEntityChanged':
      window.dispatchEvent(e)
      break
    default:
      window.showBigError('ws-unhandled-message', 'handling websocket message', 'Unhandled websocket message type from server: ' + j.type, true)
  }
}

function websocketOnError (err) {
  window.websocketAvailable = false
  window.refreshLoop()

  console.log('Websocket error is: ', err)

  window.showBigError('websocket-connection', 'connecting to the websocket', 'This often means the connection was closed, sometimes this can happen due to reverse proxy timeouts. Sometimes your web browser can provide helpful diagnostic information in the web developer console. The reason given by your browser is:' + err, true)

  refreshServerConnectionLabel()
}

function websocketOnClose () {
  window.websocketAvailable = false

  refreshServerConnectionLabel()
}
