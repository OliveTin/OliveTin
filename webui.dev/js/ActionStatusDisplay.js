export class ActionStatusDisplay {
  constructor (parentElement) {
    this.exitCodeElement = document.createElement('span')
    this.statusElement = document.createElement('span')

    parentElement.innerText = ''
    parentElement.appendChild(this.statusElement)
    parentElement.appendChild(this.exitCodeElement)
  }

  getText () {
    return this.statusElement.innerText
  }

  update (logEntry) {
    this.statusElement.classList.remove(...this.statusElement.classList)

    if (logEntry.executionFinished) {
      this.statusElement.innerText = 'Completed'
      this.exitCodeElement.innerText = ', Exit code: ' + logEntry.exitCode

      if (logEntry.exitCode === 0) {
        this.statusElement.classList.add('action-success')
      } else if (logEntry.blocked) {
        this.statusElement.innerText = 'Blocked'
        this.statusElement.classList.add('action-blocked')
        this.exitCodeElement.innerText = ''
      } else if (logEntry.timedOut) {
        this.statusElement.innerText = 'Timed out'
        this.statusElement.classList.add('action-timeout')
        this.exitCodeElement.innerText = ''
      } else {
        this.statusElement.classList.add('action-nonzero-exit')
      }
    } else {
      this.statusElement.innerText = 'Still running...'
      this.exitCodeElement.innerText = ''
    }
  }
}
