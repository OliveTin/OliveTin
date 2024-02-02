// This ExecutionDialog is NOT a custom HTML element, but rather just picks up
// the <dialog /> element out of index.html and just re-uses that - as only
// one dialog can be shown at a time.
export class ExecutionDialog {
  show (json) {
    this.executionUuid = json

    this.dlg = document.querySelector('dialog#execution-results-popup')

    this.domIcon = document.getElementById('execution-dialog-icon')
    this.domTitle = document.getElementById('execution-dialog-title')
    this.domStdout = document.getElementById('execution-dialog-stdout')
    this.domStderr = document.getElementById('execution-dialog-stderr')
    this.domDatetimeStarted = document.getElementById('execution-dialog-datetime-started')
    this.domDatetimeFinished = document.getElementById('execution-dialog-datetime-finished')
    this.domExitCode = document.getElementById('execution-dialog-exit-code')
    this.domStatus = document.getElementById('execution-dialog-status')

    this.domTitle.innerText = 'Loading...'
    this.domExitCode.innerText = '?'
    this.domStatus.className = ''
    this.domDatetimeStarted.innerText = ''
    this.domDatetimeFinished.innerText = ''
    this.domStdout.innerText = ''
    this.domStderr.innerText = ''

    this.dlg.showModal()

    this.fetchExecutionResult()
  }

  fetchExecutionResult () {
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
      this.renderResult(json)
    }).catch(err => {
      this.renderError(err)
    })
  }

  renderResult (res) {
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

    if (res.logEntry.stderr === '') {
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
