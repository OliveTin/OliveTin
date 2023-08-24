export function setupWebsocket () {
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
  console.log('open')

  const foo = '{}'

  ws.send(foo)
}

function websocketOnMessage (msg) {
  console.log(msg)
}

function websocketOnError (err) {
  window.websocketAvailable = false
  console.log(err)
}

function websocketOnClose () {
  window.websocketAvailable = false
}
