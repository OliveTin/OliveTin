// This ExecutionDialog is NOT a custom HTML element, but rather just picks up
// the <dialog /> element out of index.html and just re-uses that - as only
// one dialog can be shown at a time.
export class ExecutionDialog {
  constructFromJson (json) {
    this.executionUuid = json

    this.dlg = document.querySelector('dialog#execution-results-popup')

    this.domIcon = this.dlg.querySelector('.icon')
    this.domTitle = this.dlg.querySelector('.title')
    this.domStdout = this.dlg.querySelector('.stdout')
    this.domStderr = this.dlg.querySelector('.stderr')
    this.domDatetimeStarted = this.dlg.querySelector('.datetimeStarted')
    this.domDatetimeFinished = this.dlg.querySelector('.datetimeFinished')
    this.domExitCode = this.dlg.querySelector('.exitCode')
    this.domStatus = this.dlg.querySelector('.status')
  }

  show () {
    this.dlg.showModal()
  }

  renderResult (res) {
    this.executionUuid = res.logEntry.executionUuid

    if (res.logEntry.executionFinished) {
      this.domStatus.innerText = 'Completed'
      this.domDatetimeFinished.innerText = res.logEntry.datetimeFinished

      if (res.logEntry.blocked) {
        this.domStatus.innerText = 'Blocked'
      }

      if (res.logEntry.timedOut) {
        this.domExitCode.innerText = 'Timed out'
        this.domStatus.innerText = 'Timed out'
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

    if (res.logEntry.stderr === '') {
      this.domStderr.parentElement.hidden = true
      this.domStderr.innerText = res.logEntry.stderr
    } else {
      this.domStderr.parentElement.hidden = false
      this.domStderr.innerText = res.logEntry.stderr
    }

    this.domDatetimeStarted.innerText = res.logEntry.datetimeStarted
  }

  renderError (err) {
    this.dlg.querySelector('pre').innerText = JSON.stringify(err)
  }
}
