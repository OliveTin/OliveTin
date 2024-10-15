
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
      this.domArgs.appendChild(this.createDomLabel(arg))
      this.domArgs.appendChild(this.createDomSuggestions(arg))
      this.domArgs.appendChild(this.createDomInput(arg))
      this.domArgs.appendChild(this.createDomDescription(arg))
    }
  }

  createDomLabel (arg) {
    const domLbl = document.createElement('label')

    const lastChar = arg.title.charAt(arg.title.length - 1)

    if (lastChar === '?' || lastChar === '.' || lastChar === ':') {
      domLbl.innerHTML = arg.title
    } else {
      domLbl.innerHTML = arg.title + ':'
    }

    domLbl.setAttribute('for', arg.name)

    return domLbl
  }

  createDomSuggestions (arg) {
    if (typeof arg.suggestions !== 'object' || arg.suggestions.length === 0) {
      return document.createElement('span')
    }

    const ret = document.createElement('datalist')
    ret.setAttribute('id', arg.name + '-choices')

    for (const suggestion of Object.keys(arg.suggestions)) {
      const opt = document.createElement('option')

      opt.setAttribute('value', suggestion)

      if (typeof arg.suggestions[suggestion] !== 'undefined' && arg.suggestions[suggestion].length > 0) {
        opt.innerText = arg.suggestions[suggestion]
      }

      ret.appendChild(opt)
    }

    return ret
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
    } else {
      switch (arg.type) {
        case 'confirmation':
          this.domBtnStart.disabled = true

          domEl = document.createElement('input')
          domEl.setAttribute('type', 'checkbox')
          domEl.onchange = () => {
            this.domBtnStart.disabled = false
            domEl.disabled = true
          }
          break
        case 'datetime':
          domEl = document.createElement('input')
          domEl.setAttribute('type', 'datetime-local')
          domEl.setAttribute('step', '1')
          break
        case 'password':
        case 'email':
          domEl = document.createElement('input')
          domEl.setAttribute('type', arg.type)
          break
        default:
          domEl = document.createElement('input')

          if (arg.type.startsWith('regex:')) {
            domEl.setAttribute('pattern', arg.type.replace('regex:', ''))
          }

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
    }

    domEl.name = arg.name
    domEl.value = arg.defaultValue

    if (typeof arg.suggestions === 'object' && Object.keys(arg.suggestions).length > 0) {
      domEl.setAttribute('list', arg.name + '-choices')
    }

    this.argInputs.push(domEl)

    return domEl
  }

  createDomDescription (arg) {
    const domArgumentDescription = document.createElement('span')
    domArgumentDescription.classList.add('argument-description')
    domArgumentDescription.innerHTML = arg.description

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
