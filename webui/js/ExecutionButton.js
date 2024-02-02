import { ExecutionDialog } from './ExecutionDialog.js'

class ExecutionButton extends window.HTMLElement {
  constructFromJson (json) {
    this.executionUuid = json
    this.ellapsed = 0

    this.appendChild(document.createElement('button'))
    this.isWaiting = true

    this.setAttribute('id', 'execution-' + json)

    this.btn = this.querySelector('button')
    this.btn.innerText = 'Executing...'
    this.btn.onclick = () => {
      this.show()
    }
  }

  show () {
    if (typeof (window.executionDialog) === 'undefined') {
      window.executionDialog = new ExecutionDialog()
    }

    window.executionDialog.show(this.executionUuid)
  }

  onFinished (LogEntry) {
    if (LogEntry.timedOut) {
      this.onActionResult('action-timeout', 'Timed out')
    } else if (LogEntry.blocked) {
      this.onActionResult('action-blocked', 'Blocked!')
    } else if (LogEntry.exitCode !== 0) {
      this.onActionResult('action-nonzero-exit', 'Exit code ' + LogEntry.exitCode)
    } else {
      console.log(LogEntry)
      this.ellapsed = Math.ceil(new Date(LogEntry.datetimeFinished) - new Date(LogEntry.datetimeStarted)) / 1000
      this.onActionResult('action-success', 'Success!')
    }
  }

  onActionResult (cssClass, temporaryStatusMessage) {
    this.temporaryStatusMessage = '[' + temporaryStatusMessage + ']'
    this.updateDom()
    this.btn.classList.add(cssClass)
  }

  onActionError (err) {
    console.error('callback error', err)
    this.isWaiting = false
    this.updateDom()
    this.btn.classList.add('action-failed')
  }

  updateDom () {
    if (this.temporaryStatusMessage != null) {
      this.btn.innerText = this.temporaryStatusMessage
      this.btn.classList.add('temporary-status-message')
      this.isWaiting = false

      setTimeout(() => {
        this.temporaryStatusMessage = null
        this.btn.classList.remove('temporary-status-message')
        this.updateDom()
      }, 2000)
    } else if (this.isWaiting) {
      this.btn.innerText = 'Waiting...'
    } else {
      this.btn.innerText = this.ellapsed + 's'
      this.btn.title = this.ellapsed + ' seconds'
    }
  }
}

window.customElements.define('execution-button', ExecutionButton)
