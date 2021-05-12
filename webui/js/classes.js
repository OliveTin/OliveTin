class ActionButton extends window.HTMLButtonElement {
  constructFromJson (json) {
    this.title = json.title
    this.states = []
    this.stateLabels = []
    this.currentState = 0
    this.isWaiting = false
    this.actionCallUrl = window.restBaseUrl + 'StartAction?actionName=' + this.title

    if (json.icon !== undefined) {
      this.unicodeIcon = unescape(json.icon)
    } else {
      this.unicodeIcon = '&#x1f4a9'
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

    window.fetch(this.actionCallUrl).then(res => {
      if (!res.ok) {
        return res.json()
      }
    }).then(json => {
      this.onActionResult()
    }).catch(err => {
      this.onActionError(err)
    })
  }

  onActionResult (json) {
    this.disabled = false
    this.isWaiting = false
    this.updateHtml()
    this.classList.add('actionSuccess')
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
    if (this.isWaiting) {
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
