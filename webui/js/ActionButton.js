import { marshalLogsJsonToHtml } from './marshaller.js'
import './ArgumentForm.js'

class ActionButton extends window.HTMLButtonElement {
  constructFromJson (json) {
    this.updateIterationTimestamp = 0

    this.title = json.title
    this.temporaryStatusMessage = null
    this.isWaiting = false
    this.actionCallUrl = window.restBaseUrl + 'StartAction'

    this.updateFromJson(json)

    this.onclick = () => {
      if (json.arguments.length > 0) {
        const frm = document.createElement('form', { is: 'argument-form' })
        frm.setup(json, (args) => {
          this.startAction(args)
        })

        document.body.appendChild(frm)
      } else {
        this.startAction()
      }
    }

    this.constructTemplate()

    this.updateHtml()

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
    this.disabled = true
    this.isWaiting = true
    this.updateHtml()
    this.classList = [] // Removes old animation classes

    if (actionArgs === undefined) {
      actionArgs = []
    }

    const startActionArgs = {
      actionName: this.title,
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
    this.temporaryStatusMessage = '[ ' + temporaryStatusMessage + ' ]'
    this.updateHtml()
    this.classList.add(cssClass)

    setTimeout(() => {
      this.classList.remove(cssClass)
    }, 1000)
  }

  onActionError (err) {
    console.log('callback error', err)
    this.disabled = false
    this.isWaiting = false
    this.updateHtml()
    this.classList.add('actionFailed')

    setTimeout(() => {
      this.classList.remove('actionFailed')
    }, 1000)
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
}

window.customElements.define('action-button', ActionButton, { extends: 'button' })
