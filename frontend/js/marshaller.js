export function initMarshaller () {
  window.addEventListener('EventOutputChunk', onOutputChunk)
}

function onOutputChunk (evt) {
  const chunk = evt.payload

  if (window.terminal) {
    if (chunk.executionTrackingId === window.terminal.executionTrackingId) {
      window.terminal.write(chunk.output)
    }
  }
}
