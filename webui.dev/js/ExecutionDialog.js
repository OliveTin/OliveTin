import { ActionStatusDisplay } from './ActionStatusDisplay.js'

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

    this.domDuration = document.getElementById('execution-dialog-duration')
    this.domStatus = new ActionStatusDisplay(document.getElementById('execution-dialog-status'))

    this.domExecutionBasics = document.getElementById('execution-dialog-basics')
    this.domExecutionDetails = document.getElementById('execution-dialog-details')

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

    this.updateDuration(this.executionSeconds + ' seconds ago', '')
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

  updateDuration (started, finished) {
    if (finished === '') {
      this.domDuration.innerHTML = started
    } else {
      let delta = ''

      try {
        delta = (new Date(finished) - new Date(started)) / 1000
        delta = new Intl.RelativeTimeFormat().format(delta, 'seconds').replace('in ', '').replace('ago', '')
      } catch (e) {
        console.warn('Failed to calculate delta', e)
      }

      this.domDuration.innerHTML = started + ' &rarr; ' + finished

      if (delta !== '') {
        this.domDuration.innerHTML += ' (' + delta + ')'
      }
    }
  }

  renderExecutionResult (res) {
    this.res = res

    clearInterval(window.executionDialogTicker)

    this.domOutput.hidden = false

    if (!this.hideDetailsOnResult) {
      this.domExecutionDetails.hidden = false
    }

    this.executionTrackingId = res.logEntry.executionTrackingId

    this.domBtnKill.disabled = res.logEntry.executionFinished

    this.domStatus.update(res.logEntry)

    this.domIcon.innerHTML = res.logEntry.actionIcon
    this.domTitle.innerText = res.logEntry.actionTitle

    if (res.logEntry.executionFinished) {
      this.updateDuration(res.logEntry.datetimeStarted, res.logEntry.datetimeFinished)
    } else {
      this.updateDuration(res.logEntry.datetimeStarted, 'Still running...')
    }

    window.terminal.reset()
    window.terminal.write(res.logEntry.output, () => {
      window.terminal.fit()
    })
  }

  renderError (err) {
    this.dlg.querySelector('pre').innerText = JSON.stringify(err)
  }
}
