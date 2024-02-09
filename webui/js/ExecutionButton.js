import { ExecutionFeedbackButton } from './ExecutionFeedbackButton.js'

class ExecutionButton extends ExecutionFeedbackButton {
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

    this.domTitle = this.btn
  }

  show () {
    window.executionDialog.reset()
    window.executionDialog.show()
    window.executionDialog.fetchExecutionResult(this.executionUuid)
  }

  onExecStatusChanged () {
    this.domTitle.innerText = this.ellapsed + 's'
    this.btn.title = this.ellapsed + ' seconds'
  }
}

window.customElements.define('execution-button', ExecutionButton)
