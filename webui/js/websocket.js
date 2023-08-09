export function setupWebsocket() {
  window.websocketAvailable = false
  
  window.ws = new WebSocket('ws://localhost:1337')

  ws.addEventListener('open', websocketOnOpen)
  ws.addEventListener('message', websocketOnMessage)
  ws.addEventListener('error', websocketOnError)
  ws.addEventListener('close', websocketOnClose)
}

function websocketOnOpen(evt) {
  window.websocketAvailable = true
  console.log("open")

  const foo = '{}'
  
  ws.send(foo)
}

function websocketOnMessage(msg) {
  console.log(msg)
}

function websocketOnError(err) {
  window.websocketAvailable = false
  console.log(err)
}

function websocketOnClose() {
  window.websocketAvailable = false
}
