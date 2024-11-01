import { ActionStatusDisplay } from './ActionStatusDisplay.js'
import { OutputTerminal } from './OutputTerminal.js'

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

    this.domBtnRerun = document.getElementById('execution-dialog-rerun-action')
    this.domBtnKill = document.getElementById('execution-dialog-kill-action')

    this.domDuration = document.getElementById('execution-dialog-duration')
    this.domStatus = new ActionStatusDisplay(document.getElementById('execution-dialog-status'))

    this.domExecutionBasics = document.getElementById('execution-dialog-basics')
    this.domExecutionDetails = document.getElementById('execution-dialog-details')

    window.terminal = new OutputTerminal()
    window.terminal.open(this.domOutput)
  }

  showOutput () {
    this.domOutput.hidden = false
    this.domOutput.hidden = false
  }

  toggleSize () {
    if (this.dlg.classList.contains('big')) {
      this.dlg.classList.remove('big')
    } else {
      this.dlg.classList.add('big')
    }

    window.terminal.fit()
  }

  reset () {
    this.executionSeconds = 0
    this.executionTrackingId = 'notset'

    this.dlg.classList.remove('big')

    this.domOutputToggleBig.hidden = false

    this.domIcon.innerText = ''
    this.domTitle.innerText = 'Waiting for result... '
    this.domDuration.innerText = ''

    //    window.terminal.close()

    this.domBtnRerun.disabled = true
    this.domBtnRerun.onclick = () => {}
    this.domBtnKill.disabled = true
    this.domBtnKill.onclick = () => {}

    this.hideDetailsOnResult = false
    this.domExecutionBasics.hidden = false

    this.domExecutionDetails.hidden = true

    window.terminal.reset()
    window.terminal.fit()
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

    if (this.dlg.open) {
      this.dlg.close()
    }

    this.dlg.showModal()
  }

  rerunAction (actionId) {
    const actionButton = document.getElementById('actionButton-' + actionId)

    if (actionButton !== undefined) {
      actionButton.btn.click()
    }

    this.dlg.close()
  }

  killAction () {
    const killActionArgs = {
      executionTrackingId: this.executionTrackingId
    }

    window.fetch(window.restBaseUrl + 'KillAction', {
      cors: 'cors',
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(killActionArgs)
    }).then((res) => {
      return res.json() // This isn't used by anything. UI is updated by OnExecutionFinished like normal.
    }).catch(err => {
      throw err
    })
  }

  executionTick () {
    this.executionSeconds++

    this.updateDuration(null)
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
      cors: 'cors',
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(executionStatusArgs)
    }).then((res) => {
      if (res.ok) {
        return res.json()
      } else if (res.status === 404) {
        throw new Error('Execution not found: ' + executionTrackingId)
      } else {
        throw new Error(res.statusText)
      }
    }
    ).then((json) => {
      this.renderExecutionResult(json)
    }).catch(err => {
      console.log(err)
      this.renderError(err)
    })
  }

  updateDuration (logEntry) {
    if (logEntry == null) {
      this.domDuration.innerHTML = this.executionSeconds + ' seconds'
    } else if (!logEntry.executionStarted) {
      this.domDuration.innerHTML = logEntry.datetimeStarted + ' (request time). Not executed.'
    } else if (logEntry.executionStarted && !logEntry.executionFinished) {
      this.domDuration.innerHTML = logEntry.datetimeStarted
    } else {
      let delta = ''

      try {
        delta = (new Date(logEntry.datetimeStarted) - new Date(logEntry.datetimeStarted)) / 1000
        delta = new Intl.RelativeTimeFormat().format(delta, 'seconds').replace('in ', '').replace('ago', '')
      } catch (e) {
        console.warn('Failed to calculate delta', e)
      }

      this.domDuration.innerHTML = logEntry.datetimeStarted + ' &rarr; ' + logEntry.datetimeFinished

      if (delta !== '') {
        this.domDuration.innerHTML += ' (' + delta + ')'
      }
    }
  }

  renderExecutionResult (res) {
    this.res = res

    clearInterval(window.executionDialogTicker)

    this.domOutput.hidden = false

    if (this.hideDetailsOnResult) {
      this.domExecutionDetails.hidden = true
    }

    this.executionTrackingId = res.logEntry.executionTrackingId

    this.domBtnRerun.disabled = !res.logEntry.executionFinished
    this.domBtnRerun.onclick = () => { this.rerunAction(res.logEntry.actionId) }

    this.domBtnKill.disabled = res.logEntry.executionFinished

    this.domStatus.update(res.logEntry)

    this.domIcon.innerHTML = res.logEntry.actionIcon
    this.domTitle.innerText = res.logEntry.actionTitle
    this.domTitle.title = 'Action ID: ' + res.logEntry.actionId + '\nExecution ID: ' + res.logEntry.executionTrackingId

    this.updateDuration(res.logEntry)

    window.terminal.reset()
    window.terminal.write(res.logEntry.output, () => {
      window.terminal.fit()
    })
  }

  renderError (err) {
    window.showBigError('execution-dlg-err', 'in the execution dialog', 'Failed to fetch execution result. ' + err, false)
  }
}
