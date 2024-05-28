// This ExecutionDialog is NOT a custom HTML element, but rather just picks up
// the <dialog /> element out of index.html and just re-uses that - as only
// one dialog can be shown at a time.
export class ExecutionDialog {
  constructor () {
    this.dlg = document.querySelector('dialog#execution-results-popup')

    this.domIcon = document.getElementById('execution-dialog-icon')
    this.domTitle = document.getElementById('execution-dialog-title')
    this.domOutput = document.getElementById('execution-dialog-xterm')
    this.domOutputToggleBig = document.getElementById('execution-dialog-toggle-size')
    this.domOutputToggleBig.onclick = () => {
      this.toggleSize()
    }

    this.domBtnKill = document.getElementById('execution-dialog-kill-action')

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
      this.domOutput.parentElement.open = false
    } else {
      this.dlg.classList.add('big')
      this.domOutput.parentElement.open = true
    }
  }

  reset () {
    this.executionSeconds = 0
    this.executionTrackingId = 'notset'

    this.dlg.classList.remove('big')
    this.dlg.style.maxWidth = 'calc(100vw - 2em)'
    this.dlg.style.width = ''
    this.dlg.style.height = ''
    this.dlg.style.border = ''

    this.domOutputToggleBig.hidden = false

    this.domIcon.innerText = ''
    this.domTitle.innerText = 'Waiting for result... '
    this.domExitCode.innerText = '?'
    this.domStatus.className = ''
    this.domDatetimeStarted.innerText = ''
    this.domDatetimeFinished.innerText = ''
    this.domOutput.innerText = ''

    //    window.terminal.close()

    this.domBtnKill.disabled = true
    this.domBtnKill.onclick = () => {}

    this.hideDetailsOnResult = false
    this.domExecutionBasics.hidden = false

    this.domExecutionDetails.hidden = true
    this.domOutput.parentElement.open = false

    this.domExecutionOutput.hidden = true
  }

  show (actionButton) {
    if (typeof actionButton !== 'undefined' && actionButton != null) {
      this.domIcon.innerText = actionButton.domIcon.innerText
    }

    this.domBtnKill.disabled = false
    this.domBtnKill.onclick = () => {
      this.killAction()
    }

    clearInterval(window.executionDialogTicker)
    this.executionSeconds = 0
    this.executionTick()
    window.executionDialogTicker = setInterval(() => {
      this.executionTick()
    }, 1000)

    window.terminal.open(this.domOutput)

    if (this.dlg.open) {
      this.dlg.close()
    }

    this.dlg.showModal()
  }

  killAction () {
    const killActionArgs = {
      executionTrackingId: this.executionTrackingId
    }

    window.fetch(window.restBaseUrl + 'KillAction', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(killActionArgs)
    }).then((res) => {
      console.log(res)
    }).catch(err => {
      throw err
    })
  }

  executionTick () {
    this.executionSeconds++

    this.domDatetimeStarted.innerText = this.executionSeconds + ' seconds ago'
  }

  hideEverythingApartFromOutput () {
    this.hideDetailsOnResult = true
    this.domExecutionBasics.hidden = true
  }

  fetchExecutionResult (executionTrackingId) {
    this.executionTrackingId = executionTrackingId

    const executionStatusArgs = {
      executionTrackingId: this.executionTrackingId
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
      this.domOutput.parentElement.open = true
    }

    this.executionTrackingId = res.logEntry.executionTrackingId

    this.domBtnKill.disabled = res.logEntry.executionFinished

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

    this.domDatetimeStarted.innerText = res.logEntry.datetimeStarted

    window.terminal.clear()
    window.terminal.write(res.logEntry.stdout)
  }

  renderError (err) {
    this.dlg.querySelector('pre').innerText = JSON.stringify(err)
  }
}
