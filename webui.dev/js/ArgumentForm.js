
class ArgumentForm extends window.HTMLElement {
  getQueryParams () {
    return new URLSearchParams(window.location.search.substring(1))
  }

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
      this.clearBookmark()
      this.remove()
    }
  }

  getArgumentValues () {
    const ret = []

    for (const arg of this.argInputs) {
      if (arg.type === 'checkbox') {
        if (arg.checked) {
          arg.value = "1"
        } else {
          arg.value = "0"
        }
      }

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

    if (arg.choices.length > 0 && (arg.type === "select" || arg.type === "")) {
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
        case 'html':
          domEl = document.createElement('div')
          domEl.innerHTML = arg.defaultValue

          return domEl
        case 'confirmation':
          this.domBtnStart.disabled = true

          domEl = document.createElement('input')
          domEl.setAttribute('type', 'checkbox')
          domEl.onchange = () => {
            this.domBtnStart.disabled = false
            domEl.disabled = true
          }
          break
        case 'raw_string_multiline':
          domEl = document.createElement('textarea')
          domEl.setAttribute('rows', '5')
          domEl.style.resize = 'vertical'
          break
        case 'datetime':
          domEl = document.createElement('input')
          domEl.setAttribute('type', 'datetime-local')
          domEl.setAttribute('step', '1')
          break
        case 'checkbox':
          domEl = document.createElement('input')
          domEl.setAttribute('type', 'checkbox')
          domEl.setAttribute('name', arg.name)
          domEl.setAttribute('value', "1")
          
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

    // Use query parameter value if available
    const params = this.getQueryParams()
    const paramValue = params.get(arg.name)

    if (paramValue !== null) {
      domEl.value = paramValue
    } else {
      domEl.value = arg.defaultValue
    }

    // update the URL when a parameter is changed
    domEl.addEventListener('change', this.updateUrlWithArg)

    if (typeof arg.suggestions === 'object' && Object.keys(arg.suggestions).length > 0) {
      domEl.setAttribute('list', arg.name + '-choices')
    }

    this.argInputs.push(domEl)

    return domEl
  }

  updateUrlWithArg (ev) {
    if (!ev.target.name) {
      return
    }

    const url = new URL(window.location.href)

    if (ev.target.type === 'password') {
      return
    }

    // copy the parameter value
    url.searchParams.set(ev.target.name, ev.target.value)

    // Update the URL without reloading the page
    window.history.replaceState({}, '', url.toString())
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

  clearBookmark () {
    // remove the action from the URL
    window.history.replaceState({
      path: window.location.pathname
    }, '', window.location.pathname)
  }
}

window.customElements.define('argument-form', ArgumentForm)
