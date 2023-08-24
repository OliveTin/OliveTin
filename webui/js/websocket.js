function websocketInit() {
  let proto = "ws:"

  if (window.location.protocol == "https:") {
    proto = "wss:"
  }

  let wsUrl = proto + window.location.host + '/websocket'

  console.log(wsUrl)

  window.ws = new WebSocket(wsUrl)

  window.ws.addEventListener("open", () => {
    console.log("socket opened!")
  window.ws.send("Hi!")
  })

  window.ws.addEventListener("error", () => {
    console.error("ws error")
  })

  window.ws.addEventListener("message", (msg) => {
    console.log("ws msg", msg)
  })

}

websocketInit()
