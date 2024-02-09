// This ExecutionDialog is NOT a custom HTML element, but rather just picks up
// the <dialog /> element out of index.html and just re-uses that - as only
// one dialog can be shown at a time.
export class ExecutionDialog {
  constructor () {
    this.dlg = document.querySelector('dialog#execution-results-popup')

    this.domIcon = document.getElementById('execution-dialog-icon')
    this.domTitle = document.getElementById('execution-dialog-title')
    this.domStdout = document.getElementById('execution-dialog-stdout')
    this.domStderr = document.getElementById('execution-dialog-stderr')
    this.domStdoutToggleBig = document.getElementById('execution-dialog-toggle-size')
    this.domStdoutToggleBig.onclick = () => {
      this.toggleSize()
    }

    this.domDatetimeStarted = document.getElementById('execution-dialog-datetime-started')
    this.domDatetimeFinished = document.getElementById('execution-dialog-datetime-finished')
    this.domExitCode = document.getElementById('execution-dialog-exit-code')
    this.domStatus = document.getElementById('execution-dialog-status')

    this.domExecutionBasics = document.getElementById('execution-dialog-basics')
    this.domExecutionDetails = document.getElementById('execution-dialog-details')
    this.domExecutionOutput = document.getElementById('execution-dialog-output')
  }

  toggleSize () {
    if (this.dlg.classList.contains('big')) {
      this.dlg.classList.remove('big')
      this.domStdout.parentElement.open = false
    } else {
      this.dlg.classList.add('big')
      this.domStdout.parentElement.open = true
    }
  }

  reset () {
    this.executionSeconds = 0

    this.dlg.classList.remove('big')
    this.dlg.style.maxWidth = 'calc(100vw - 2em)'
    this.dlg.style.width = ''
    this.dlg.style.height = ''
    this.dlg.style.border = ''

    this.domStdoutToggleBig.hidden = false

    this.domIcon.innerText = ''
    this.domTitle.innerText = 'Waiting for result... '
    this.domExitCode.innerText = '?'
    this.domStatus.className = ''
    this.domDatetimeStarted.innerText = ''
    this.domDatetimeFinished.innerText = ''
    this.domStdout.innerText = ''
    this.domStderr.innerText = ''

    this.hideDetailsOnResult = false
    this.domExecutionBasics.hidden = false

    this.domExecutionDetails.hidden = true
    this.domStdout.parentElement.open = false

    this.domExecutionOutput.hidden = true
  }

  show (actionButton) {
    if (typeof actionButton !== 'undefined' && actionButton != null) {
      this.domIcon.innerText = actionButton.domIcon.innerText
    }

    clearInterval(window.executionDialogTicker)
    this.executionSeconds = 0
    this.executionTick()
    window.executionDialogTicker = setInterval(() => {
      this.executionTick()
    }, 1000)

    this.dlg.showModal()
  }

  executionTick () {
    this.executionSeconds++

    this.domDatetimeStarted.innerText = this.executionSeconds + ' seconds ago'
  }

  hideEverythingApartFromOutput () {
    this.hideDetailsOnResult = true
    this.domExecutionBasics.hidden = true
  }

  fetchExecutionResult (uuid) {
    this.executionUuid = uuid

    const executionStatusArgs = {
      executionUuid: this.executionUuid
    }

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
      this.renderExecutionResult(json)
    }).catch(err => {
      this.renderError(err)
    })
  }

  renderExecutionResult (res) {
    clearInterval(window.executionDialogTicker)

    this.domExecutionOutput.hidden = false

    if (!this.hideDetailsOnResult) {
      this.domExecutionDetails.hidden = false
    } else {
      this.domStdout.parentElement.open = true
    }

    this.executionUuid = res.logEntry.executionUuid

    if (res.logEntry.executionFinished) {
      this.domStatus.innerText = 'Completed'
      this.domStatus.classList.add('action-success')
      this.domDatetimeFinished.innerText = res.logEntry.datetimeFinished

      if (res.logEntry.timedOut) {
        this.domExitCode.innerText = 'Timed out'
        this.domStatus.classList.add('action-timeout')
      } else if (res.logEntry.blocked) {
        this.domStatus.innerText = 'Blocked'
        this.domStatus.classList.add('action-blocked')
      } else if (res.logEntry.exitCode !== 0) {
        this.domStatus.innerText = 'Non-Zero Exit'
        this.domStatus.classList.add('action-nonzero-exit')
      } else {
        this.domExitCode.innerText = res.logEntry.exitCode
      }
    } else {
      this.domDatetimeFinished.innerText = 'Still running...'
      this.domExitCode.innerText = 'Still running...'
      this.domStatus.innerText = 'Still running...'
    }

    this.domIcon.innerHTML = res.logEntry.actionIcon
    this.domTitle.innerText = res.logEntry.actionTitle

    this.domStdout.innerText = res.logEntry.stdout
    this.domStdout.innerText = res.logEntry.stdout

    if (res.logEntry.stderr === '(empty)') {
      this.domStderr.parentElement.style.display = 'none'
      this.domStderr.innerText = res.logEntry.stderr
    } else {
      this.domStderr.parentElement.style.display = 'block'
      this.domStderr.innerText = res.logEntry.stderr
    }

    this.domDatetimeStarted.innerText = res.logEntry.datetimeStarted
  }

  renderError (err) {
    this.dlg.querySelector('pre').innerText = JSON.stringify(err)
  }
}
