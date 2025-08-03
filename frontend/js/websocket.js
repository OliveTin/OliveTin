import {
  refreshServerConnectionLabel
} from './marshaller.js'

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
    for await (let e of window.client.eventStream()) {
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

  console.log("Websocket event receved: ", typeName)

  const j = new Event(typeName)
  j.payload = msg.event.value

  switch (typeName) {
    case 'EventOutputChunk':
    case 'EventConfigChanged':
    case 'EventEntityChanged':
    case 'EventExecutionFinished':
    case 'EventExecutionStarted':
      window.dispatchEvent(j)
      break
    default:
      console.warn('Unhandled websocket message type from server: ', typeName)

      window.showBigError('ws-unhandled-message', 'handling websocket message', 'Unhandled websocket message type from server: ' + typeName, true)
  }
}
