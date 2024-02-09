export class ExecutionFeedbackButton extends window.HTMLElement {
  onExecutionFinished (LogEntry) {
    if (LogEntry.timedOut) {
      this.renderExecutionResult('action-timeout', 'Timed out')
    } else if (LogEntry.blocked) {
      this.renderExecutionResult('action-blocked', 'Blocked!')
    } else if (LogEntry.exitCode !== 0) {
      this.renderExecutionResult('action-nonzero-exit', 'Exit code ' + LogEntry.exitCode)
    } else {
      this.ellapsed = Math.ceil(new Date(LogEntry.datetimeFinished) - new Date(LogEntry.datetimeStarted)) / 1000
      this.renderExecutionResult('action-success', 'Success!')
    }
  }

  renderExecutionResult (resultCssClass, temporaryStatusMessage) {
    this.updateDom(resultCssClass, '[' + temporaryStatusMessage + ']')
    this.onExecStatusChanged()
  }

  updateDom (resultCssClass, title) {
    if (resultCssClass == null) {
      this.btn.className = ''
    } else {
      this.btn.classList.add(resultCssClass)
    }

    this.domTitle.innerText = title
  }
}
