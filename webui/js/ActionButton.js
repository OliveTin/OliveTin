import { marshalLogsJsonToHtml } from './marshaller.js'
import './ArgumentForm.js'

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
    this.btn.title = json.title
    this.btn.onclick = () => {
      console.log(json.arguments)
      if (json.arguments.length > 0) {
        const frm = document.createElement('argument-form')
        frm.setup(json, (args) => {
          this.startAction(args)
        })

        document.body.appendChild(frm)
      } else {
        this.startAction()
      }
    }

    this.updateFromJson(json)

    this.updateDom()

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

  startAction (actionArgs) {
    this.btn.disabled = true
    this.isWaiting = true
    this.updateDom()
    this.btn.classList = [] // Removes old animation classes

    if (actionArgs === undefined) {
      actionArgs = []
    }

    const startActionArgs = {
      actionName: this.btn.title,
      arguments: actionArgs
    }

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
      marshalLogsJsonToHtml({ logs: [json.logEntry] })

      if (json.logEntry.timedOut) {
        this.onActionResult('actionTimeout', 'Timed out')
      } else if (json.logEntry.exitCode === -1337) {
        this.onActionError('Error')
      } else if (json.logEntry.exitCode !== 0) {
        this.onActionResult('actionNonZeroExit', 'Exit code ' + json.logEntry.exitCode)
      } else {
        this.onActionResult('actionSuccess', 'Success!')
      }
    }).catch(err => {
      this.onActionError(err)
    })
  }

  onActionResult (cssClass, temporaryStatusMessage) {
    this.btn.disabled = false
    this.temporaryStatusMessage = '[ ' + temporaryStatusMessage + ' ]'
    this.updateDom()
    this.btn.classList.add(cssClass)

    setTimeout(() => {
      this.btn.classList.remove(cssClass)
    }, 1000)
  }

  onActionError (err) {
    console.error('callback error', err)
    this.btn.disabled = false
    this.isWaiting = false
    this.updateDom()
    this.btn.classList.add('actionFailed')

    setTimeout(() => {
      this.btn.classList.remove('actionFailed')
    }, 1000)
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

  updateDom () {
    if (this.temporaryStatusMessage != null) {
      this.domTitle.innerText = this.temporaryStatusMessage
      this.domTitle.classList.add('temporaryStatusMessage')
      this.isWaiting = false
      this.disabled = false

      setTimeout(() => {
        this.temporaryStatusMessage = null
        this.domTitle.classList.remove('temporaryStatusMessage')
        this.updateDom()
      }, 2000)
    } else if (this.isWaiting) {
      this.domTitle.innerText = 'Waiting...'
    } else {
      this.domTitle.innerText = this.btn.title
    }

    this.domIcon.innerHTML = this.unicodeIcon
  }
}

window.customElements.define('action-button', ActionButton)
