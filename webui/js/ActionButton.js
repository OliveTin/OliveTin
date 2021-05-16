class ActionButton extends window.HTMLButtonElement {
  constructFromJson (json) {
    this.title = json.title
    this.states = []
    this.stateLabels = []
    this.temporaryStatusMessage = null
    this.currentState = 0
    this.isWaiting = false
    this.actionCallUrl = window.restBaseUrl + 'StartAction?actionName=' + this.title

    if (json.icon == "") {
      this.unicodeIcon = '&#x1f4a9'
    } else {
      this.unicodeIcon = unescape(json.icon)
    }

    this.onclick = () => { this.startAction() }

    this.constructTemplate()
    this.updateHtml()
  }

  startAction () {
    this.disabled = true
    this.isWaiting = true
    this.updateHtml()
    this.classList = [] // Removes old animation classes

    window.fetch(this.actionCallUrl).then(res => res.json()
    ).then((json) => {
      if (json.timedOut) {
        this.onActionResult('actionTimeout', "Timed out")
      } else if (json.exitCode != 0) {
        this.onActionResult('actionNonZeroExit', "Exit code " + json.exitCode)
      } else {
        this.onActionResult('actionSuccess', "Success!")
      }
    }).catch(err => {
      this.onActionError(err)
    })
  }

  onActionResult (cssClass, temporaryStatusMessage) {
    this.temporaryStatusMessage = '[ ' + temporaryStatusMessage + ' ]'
    this.updateHtml()
    this.classList.add(cssClass)
  }

  onActionError (err) {
    console.log('callback error', err)
    this.disabled = false
    this.isWaiting = false
    this.updateHtml()
    this.classList.add('actionFailed')
  }

  constructTemplate () {
    const tpl = document.getElementById('tplActionButton')
    const content = tpl.content.cloneNode(true)

    /*
     * FIXME: Should probably be using a shadowdom here, but seem to
     * get an error when combined with custom elements.
     */

    this.appendChild(content)

    this.domTitle = this.querySelector('.title')
    this.domIcon = this.querySelector('.icon')
  }

  updateHtml () {
    if (this.temporaryStatusMessage != null) {
      this.domTitle.innerText = this.temporaryStatusMessage
      this.domTitle.classList.add('temporaryStatusMessage')
      this.isWaiting = false
      this.disabled = false

      setTimeout(() => { 
        this.temporaryStatusMessage = null
        this.domTitle.classList.remove('temporaryStatusMessage')
        this.updateHtml() 
      }, 2000)
    } else if (this.isWaiting) {
      this.domTitle.innerText = 'Waiting...'
    } else {
      this.domTitle.innerText = this.title
    }

    this.domIcon.innerHTML = this.unicodeIcon
  }

  getCurrentStateLabel (useLabels = true) {
    if (useLabels) {
      return this.stateLabels[this.currentState]
    } else {
      return this.states[this.currentState]
    }
  }

  getNextStateLabel () {
    return this.stateLabels[this.currentState + 1]
  }
}

window.customElements.define('action-button', ActionButton, { extends: 'button' })
