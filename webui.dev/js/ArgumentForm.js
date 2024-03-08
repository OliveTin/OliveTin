
class ArgumentForm extends window.HTMLElement {
  setup (json, callback) {
    this.setAttribute('class', 'action-arguments')

    this.constructTemplate()
    this.domTitle.innerText = json.title
    this.domIcon.innerHTML = json.icon
    this.createDomFormArguments(json.arguments)

    this.domBtnStart.onclick = () => {
      for (const arg of this.argInputs) {
        if (!arg.validity.valid) {
          return
        }
      }

      const argvs = this.getArgumentValues()

      callback(argvs)

      this.remove()
    }

    this.domBtnCancel.onclick = () => {
      this.remove()
    }
  }

  getArgumentValues () {
    const ret = []

    for (const arg of this.argInputs) {
      ret.push({
        name: arg.name,
        value: arg.value
      })
    }

    return ret
  }

  constructTemplate () {
    const tpl = document.getElementById('tplArgumentForm')
    const content = tpl.content.cloneNode(true)

    this.appendChild(content)

    this.domTitle = this.querySelector('h2')
    this.domIcon = this.querySelector('span.icon')
    this.domWrapper = this.querySelector('.wrapper')

    this.domArgs = this.querySelector('.arguments')

    this.domBtnStart = this.querySelector('[name=start]')
    this.domBtnCancel = this.querySelector('[name=cancel]')
  }

  createDomFormArguments (args) {
    this.argInputs = []

    for (const arg of args) {
      const domArgumentWrapper = document.createElement('p')
      domArgumentWrapper.classList.add('argument-wrapper')

      domArgumentWrapper.appendChild(this.createDomLabel(arg))
      domArgumentWrapper.appendChild(this.createDomInput(arg))
      domArgumentWrapper.appendChild(this.createDomDescription(arg))

      this.domArgs.appendChild(domArgumentWrapper)
    }
  }

  createDomLabel (arg) {
    const domLbl = document.createElement('label')
    domLbl.innerText = arg.title + ':'
    domLbl.setAttribute('for', arg.name)

    return domLbl
  }

  createDomInput (arg) {
    let domEl = null

    if (arg.choices.length > 0) {
      domEl = document.createElement('select')

      // select/choice elements don't get an onchange/validation because theoretically
      // the user should only select from a dropdown of valid options. The choices are
      // riggeriously checked on StartAction anyway. ValidateArgumentType is only
      // meant for showing simple warnings in the UI before running.

      for (const choice of arg.choices) {
        domEl.appendChild(this.createSelectOption(choice))
      }
    } else if (arg.type === 'confirmation') {
      this.domBtnStart.disabled = true

      domEl = document.createElement('input')
      domEl.setAttribute('type', 'checkbox')
      domEl.onchange = () => {
        this.domBtnStart.disabled = false
        domEl.disabled = true
      }
    } else if (arg.type === 'datetime') {
      domEl = document.createElement('input')
      domEl.setAttribute('type', 'datetime-local')
      domEl.setAttribute('step', '1')
    } else {
      domEl = document.createElement('input')
      domEl.onchange = () => {
        const validateArgumentTypeArgs = {
          value: domEl.value,
          type: arg.type
        }

        window.fetch(window.restBaseUrl + 'ValidateArgumentType', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify(validateArgumentTypeArgs)
        }).then((res) => {
          if (res.ok) {
            return res.json()
          } else {
            throw new Error(res.statusText)
          }
        }).then((json) => {
          if (json.valid) {
            domEl.setCustomValidity('')
          } else {
            domEl.setCustomValidity(json.description)
          }
        })
      }
    }

    domEl.name = arg.name
    domEl.value = arg.defaultValue

    this.argInputs.push(domEl)

    return domEl
  }

  createDomDescription (arg) {
    const domArgumentDescription = document.createElement('span')
    domArgumentDescription.classList.add('argument-description')
    domArgumentDescription.innerText = arg.description

    return domArgumentDescription
  }

  createSelectOption (choice) {
    const domEl = document.createElement('option')

    domEl.setAttribute('value', choice.value)
    domEl.innerText = choice.title

    return domEl
  }
}

window.customElements.define('argument-form', ArgumentForm)
