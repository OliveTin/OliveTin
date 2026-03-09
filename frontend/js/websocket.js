import { buttonResults } from '../resources/vue/stores/buttonResults.js'
import { rateLimits } from '../resources/vue/stores/rateLimits.js'
import { connectionState } from '../resources/vue/stores/connectionState.js'

const RECONNECT_DELAY_MS = 10000

export function initWebsocket () {
  window.addEventListener('EventOutputChunk', onOutputChunk)
  window.addEventListener('EventExecutionStarted', onExecutionChanged)
  window.addEventListener('EventExecutionFinished', onExecutionChanged)

  reconnectWebsocket()
}

window.websocketAvailable = false

async function reconnectWebsocket () {
  if (window.websocketAvailable) {
    return
  }

  connectionState.reconnecting = true
  connectionState.connected = false
  if (connectionState.disconnectedAt == null) {
    connectionState.disconnectedAt = Date.now()
  }
  connectionState.nextReconnectAt = null

  try {
    window.websocketAvailable = true
    const stream = window.client.eventStream()
    connectionState.connected = true
    connectionState.reconnecting = false
    connectionState.nextReconnectAt = null
    for await (const e of stream) {
      connectionState.disconnectedAt = null
      handleEvent(e)
    }
  } catch (err) {
    console.error('Websocket connection failed: ', err)
  }

  window.websocketAvailable = false
  connectionState.connected = false
  connectionState.disconnectedAt = connectionState.disconnectedAt ?? Date.now()
  connectionState.nextReconnectAt = Date.now() + RECONNECT_DELAY_MS
  console.log('Reconnecting websocket in ' + RECONNECT_DELAY_MS + 'ms...')
  setTimeout(() => {
    reconnectWebsocket()
  }, RECONNECT_DELAY_MS)
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

function onExecutionChanged (evt) {
  buttonResults[evt.payload.logEntry.executionTrackingId] = evt.payload.logEntry

  const logEntry = evt.payload.logEntry

  // Update rate limit store from logEntry if rate limit expiry datetime is provided
  if (logEntry && logEntry.datetimeRateLimitExpires && logEntry.bindingId) {
    // Parse datetime string "2006-01-02 15:04:05" and convert to Unix timestamp
    const date = new Date(logEntry.datetimeRateLimitExpires.replace(' ', 'T') + 'Z')
    rateLimits[logEntry.bindingId] = date.getTime() / 1000
  } else if (logEntry && logEntry.bindingId) {
    // Clear rate limit if not set
    rateLimits[logEntry.bindingId] = 0
  }
}
