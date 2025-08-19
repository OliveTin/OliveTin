/**
 * This is a weird function that just sets some globals.
 */
export function initMarshaller () {
  window.logEntries = new Map()

  window.addEventListener('EventExecutionStarted', onExecutionStarted)
  window.addEventListener('EventExecutionFinished', onExecutionFinished)
  window.addEventListener('EventOutputChunk', onOutputChunk)
}

function onOutputChunk (evt) {
  const chunk = evt.payload

  return;
  if (chunk.executionTrackingId === window.executionDialog.executionTrackingId) {
    window.terminal.write(chunk.output)
  }
}

function onExecutionStarted (evt) {
  const logEntry = evt.payload.logEntry

  // marshalLogsJsonToHtml({
  //   logs: [logEntry]
  // })
}

function onExecutionFinished (evt) {
  const logEntry = evt.payload.logEntry

  window.logEntries.set(logEntry.executionTrackingId, logEntry)

  return;

  const executionButton = document.querySelector('execution-button#execution-' + logEntry.executionTrackingId)
  let feedbackButton = actionButton

  switch (actionButton.popupOnStart) {
    case 'execution-button':
      if (executionButton != null) {
        feedbackButton = executionButton
      }

      break
    case 'execution-dialog-output-html':
    case 'execution-dialog-stdout-only':
    case 'execution-dialog':
      // We don't need to fetch the logEntry for the dialog because we already
      // have it, so we open the dialog and it will get updated below.

      window.executionDialog.show()
      window.executionDialog.executionTrackingId = logEntry.uuid

      break
  }

  feedbackButton.onExecutionFinished(logEntry)

  // marshalLogsJsonToHtml({
  //   logs: [logEntry]
  // })

  // If the current execution dialog is open, update that too
  if (window.executionDialog.dlg.open && window.executionDialog.executionUuid === logEntry.uuid) {
    window.executionDialog.renderExecutionResult({
      logEntry: logEntry,
      type: actionButton.popupOnStart
    })
  }
}
