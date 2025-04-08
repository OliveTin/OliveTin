import './ExecutionButton.js'
import './ArgumentForm.js'
import { ExecutionFeedbackButton } from './ExecutionFeedbackButton.js'

class ActionButton extends ExecutionFeedbackButton {
  constructDomFromTemplate () {
    const tpl = document.getElementById('tplActionButton')
    const content = tpl.content.cloneNode(true)

    /*
     * FIXME: Should probably be using a shadowdom here, but seem to
     * get an error when combined with custom elements.
     */

    this.appendChild(content)

    this.btn = this.querySelector('button')
    this.domTitle = this.btn.querySelector('.title')
    this.domIcon = this.btn.querySelector('.icon')
  }

  constructFromJson (json) {
    this.updateIterationTimestamp = 0

    this.constructDomFromTemplate()

    // Class attributes
    this.updateFromJson(json)

    this.actionId = json.id

    // DOM Attributes
    this.setAttribute('role', 'none')
    this.setAttribute('id', 'actionButton-' + this.actionId)

    if (!json.canExec) {
      this.btn.disabled = true
    }

    this.btn.setAttribute('id', 'actionButtonInner-' + this.actionId)
    this.btn.title = json.title
    this.btn.onclick = () => {
      if (json.arguments.length > 0) {
        for (const oldArgumentForm of document.querySelectorAll('argument-form')) {
          oldArgumentForm.remove()
        }

        this.updateUrlWithAction()

        const frm = document.createElement('argument-form')
        frm.setup(json, (args) => {
          this.startAction(args)
        })

        document.body.appendChild(frm)
        frm.querySelector('dialog').showModal()
      } else {
        this.startAction()
      }
    }

    this.popupOnStart = json.popupOnStart

    this.updateFromJson(json)

    this.domTitle.innerText = this.btn.title
    this.domIcon.innerHTML = this.unicodeIcon
  }

  updateFromJson (json) {
    // Fields that should not be updated
    //
    // title - as the callback URL relies on it

    if (json.icon === '') {
      this.unicodeIcon = '&#x1f4a9'
    } else {
      this.unicodeIcon = unescape(json.icon)
    }

    this.domIcon.innerHTML = this.unicodeIcon
  }

  onExecStatusChanged () {
    this.btn.disabled = false

    setTimeout(() => {
      this.updateDom(null, this.btn.title)
    }, 2000)
  }

  getUniqueId () {
    if (window.isSecureContext) {
      return window.crypto.randomUUID()
    } else {
      return Date.now().toString()
    }
  }

  startAction (actionArgs) {
    this.btn.classList = [] // Removes old animation classes

    if (actionArgs === undefined) {
      actionArgs = []
    }

    // UUIDs are create client side, so that we can setup a "execution-button"
    // to track the execution before we send the request to the server.
    const startActionArgs = {
      actionId: this.actionId,
      arguments: actionArgs,
      uniqueTrackingId: this.getUniqueId()
    }

    this.onActionStarted(startActionArgs.uniqueTrackingId)

    window.fetch(window.restBaseUrl + 'StartAction', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(startActionArgs)
    }).then((res) => {
      if (res.ok) {
        return res.json()
      } else {
        throw new Error(res.statusText)
      }
    }
    ).then((json) => {
      // The button used to wait for the action to finish, but now it is fire & forget
    }).catch(err => {
      throw err // We used to flash buttons red, but now hand to the global error handler
    })
  }

  updateUrlWithAction () {
    // Get the current URL and create a new URL object
    const url = new URL(window.location.href)

    // Set the action parameter
    url.searchParams.set('action', this.btn.title)

    // Update the URL without reloading the page
    window.history.replaceState({}, '', url.toString())
  }

  onActionStarted (executionTrackingId) {
    if (this.popupOnStart === 'execution-button') {
      const btnExecution = document.createElement('execution-button')
      btnExecution.constructFromJson(executionTrackingId)
      this.querySelector('.action-button-footer').hidden = false
      this.querySelector('.action-button-footer').style.display = 'flex'
      this.querySelector('.action-button-footer').prepend(btnExecution)

      return
    }

    if (this.popupOnStart.includes('execution-dialog')) {
      window.executionDialog.reset()

      if (this.popupOnStart === 'execution-dialog-stdout-only') {
        window.executionDialog.hideEverythingApartFromOutput()
      }

      window.executionDialog.executionTrackingId = executionTrackingId
      window.executionDialog.show(this)
    }

    this.btn.disabled = true
  }
}

window.customElements.define('action-button', ActionButton)
