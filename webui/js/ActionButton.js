import './ArgumentForm.js'
import './ExecutionButton.js'

class ActionButton extends window.HTMLElement {
  constructFromJson (json) {
    this.updateIterationTimestamp = 0

    this.constructDomFromTemplate()

    // Class attributes
    this.temporaryStatusMessage = null
    this.isWaiting = false
    this.actionCallUrl = window.restBaseUrl + 'StartAction'

    this.updateFromJson(json)

    // DOM Attributes
    this.setAttribute('role', 'none')
    this.btn.title = json.title
    this.btn.onclick = () => {
      if (json.arguments.length > 0) {
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

    this.updateFromJson(json)

    this.domTitle.innerText = this.btn.title
    this.domIcon.innerHTML = this.unicodeIcon

    this.setAttribute('id', 'actionButton_' + json.id)
  }

  updateFromJson (json) {
    // Fields that should not be updated
    //
    // title - as the callback URL relies on it
    // actionCallbackUrl - as it's based on the title
    // temporaryStatusMessage - as the button might be "waiting" on execution to finish while it's being updated.

    if (json.icon === '') {
      this.unicodeIcon = '&#x1f4a9'
    } else {
      this.unicodeIcon = unescape(json.icon)
    }
  }

  getUniqueId () {
    if (window.isSecureContext) {
      return window.crypto.randomUUID()
    } else {
      return Date.now().toString()
    }
  }

  startAction (actionArgs) {
    //    this.btn.disabled = true
    //    this.isWaiting = true
    //    this.updateDom()
    this.btn.classList = [] // Removes old animation classes

    if (actionArgs === undefined) {
      actionArgs = []
    }

    // UUIDs are create client side, so that we can setup a "execution-button"
    // to track the execution before we send the request to the server.
    const startActionArgs = {
      actionName: this.btn.title,
      arguments: actionArgs,
      uuid: this.getUniqueId()
    }

    const btnExecution = document.createElement('execution-button')
    btnExecution.constructFromJson(startActionArgs.uuid)
    this.querySelector('div.executions').appendChild(btnExecution)

    window.fetch(this.actionCallUrl, {
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
      this.onActionError(err)
    })
  }

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
}

window.customElements.define('action-button', ActionButton)
