import { ExecutionDialog } from './ExecutionDialog.js'

class ExecutionButton extends window.HTMLElement {
  constructFromJson (json) {
    this.executionUuid = json

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
    if (window.executionDialog === undefined) {
      window.executionDialog = new ExecutionDialog()
    }

    const executionStatusArgs = {
      executionUuid: this.executionUuid
    }

    window.executionDialog.constructFromJson(this.executionUuid)
    window.executionDialog.show()

    window.fetch(window.restBaseUrl + 'ExecutionStatus', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(executionStatusArgs)
    }).then((res) => {
      if (res.ok) {
        return res.json()
      } else {
        throw new Error(res.statusText)
      }
    }
    ).then((json) => {
      window.executionDialog.renderResult(json)
    }).catch(err => {
      window.executionDialog.renderError(err)
    })
  }

  onFinished (LogEntry) {
    if (LogEntry.timedOut) {
      this.onActionResult('action-timeout', 'Timed out')
    } else if (LogEntry.exitCode === -1337) {
      this.onActionError('Error')
    } else if (LogEntry.exitCode !== 0) {
      this.onActionResult('action-nonzero-exit', 'Exit code ' + LogEntry.exitCode)
    } else {
      this.onActionResult('action-success', 'Success!')
    }
  }

  onActionResult (cssClass, temporaryStatusMessage) {
    this.btn.disabled = false
    this.temporaryStatusMessage = '[ ' + temporaryStatusMessage + ' ]'
    this.updateDom()
    this.btn.classList.add(cssClass)

    setTimeout(() => {
      this.btn.classList.remove(cssClass)
    }, 1000)
  }

  onActionError (err) {
    console.error('callback error', err)
    this.btn.disabled = false
    this.isWaiting = false
    this.updateDom()
    this.btn.classList.add('action-failed')

    setTimeout(() => {
      this.btn.classList.remove('action-failed')
    }, 1000)
  }

  updateDom () {
    if (this.temporaryStatusMessage != null) {
      this.btn.innerText = this.temporaryStatusMessage
      this.btn.classList.add('temporary-status-message')
      this.isWaiting = false
      this.disabled = false

      setTimeout(() => {
        this.temporaryStatusMessage = null
        this.btn.classList.remove('temporary-status-message')
        this.updateDom()
      }, 2000)
    } else if (this.isWaiting) {
      this.btn.innerText = 'Waiting...'
    } else {
      this.btn.innerText = 'Finished'
    }
  }
}

window.customElements.define('execution-button', ExecutionButton)
