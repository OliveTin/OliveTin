import { buttonResults } from '../resources/vue/stores/buttonResults.js'

export function initWebsocket () {
  window.addEventListener('EventOutputChunk', onOutputChunk)

  window.checkWebsocketConnection = checkWebsocketConnection
}

export function checkWebsocketConnection () {
  reconnectWebsocket()
}

window.websocketAvailable = false

async function reconnectWebsocket () {
  if (window.websocketAvailable) {
    return
  }

  try {
    window.websocketAvailable = true
    for await (const e of window.client.eventStream()) {
      handleEvent(e)
    }
  } catch (err) {
    console.error('Websocket connection failed: ', err)
  }

  window.websocketAvailable = false
  console.log('Reconnecting websocket...')
}

function handleEvent (msg) {
  const typeName = msg.event.value.$typeName.replace('olivetin.api.v1.', '')

  const j = new Event(typeName)
  j.payload = msg.event.value

  switch (typeName) {
    case 'EventOutputChunk':
    case 'EventConfigChanged':
    case 'EventEntityChanged':
      window.dispatchEvent(j)
      break
    case 'EventExecutionFinished':
    case 'EventExecutionStarted':
      buttonResults[msg.event.value.logEntry.executionTrackingId] = msg.event.value.logEntry
      window.dispatchEvent(j)
      break
    default:
      console.warn('Unhandled websocket message type from server: ', typeName)

      window.showBigError('ws-unhandled-message', 'handling websocket message', 'Unhandled websocket message type from server: ' + typeName, true)
  }
}

function onOutputChunk (evt) {
  const chunk = evt.payload

  if (window.terminal) {
    if (chunk.executionTrackingId === window.terminal.executionTrackingId) {
      window.terminal.write(chunk.output)
    }
  }
}
