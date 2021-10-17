
class ArgumentForm extends window.HTMLFormElement {
  setup(json, callback) {
    this.setAttribute('class', 'actionArguments')

    console.log(json)

    this.domWrapper = document.createElement('div')
    this.domWrapper.classList += 'wrapper'
    this.appendChild(this.domWrapper)

    this.domTitle = document.createElement('h2')
    this.domTitle.innerText = json.title + ": Arguments"
    this.domWrapper.appendChild(this.domTitle);

    this.domIcon = document.createElement('span');
    this.domIcon.classList += 'icon'
    this.domIcon.setAttribute('role', 'img')
    this.domIcon.innerHTML = json.icon
    this.domTitle.prepend(this.domIcon)

    let a = document.createElement("span")
    a.innerText = "This is test version of the form."
    this.domWrapper.appendChild(a)

    this.createDomFormArguments(json.arguments)
    this.domWrapper.appendChild(this.createDomSubmit())

    console.log(json)
  }

  createDomSubmit() {
    let el = document.createElement('button')
    el.setAttribute('action', 'submit')
    el.innerText = "Run"

    return el
  }

  createDomFormArguments(args) {
    for (let arg of args) {
      let domFieldWrapper = document.createElement('p');

      domFieldWrapper.appendChild(this.createDomLabel(arg))
      domFieldWrapper.appendChild(this.createDomInput(arg))

      this.domWrapper.appendChild(domFieldWrapper)
    }
  }

  createDomLabel(arg) {
    let domLbl = document.createElement('label')
    domLbl.innerText = arg.label + ':';
    domLbl.setAttribute('for', arg.name)

    return domLbl;
  }

  createDomInput(arg) {
    let domEl = null;

    if (arg.choices.length > 0) {
      domEl = document.createElement('select')

      for (let choice of arg.choices) {
        domEl.appendChild(this.createSelectOption(choice))
      }
    } else {
      domEl = document.createElement('input')
    }

    domEl.setAttribute('id', arg.name)
    domEl.value = arg.defaultValue

    return domEl;
  }

  createSelectOption(choice) {
    let domEl = document.createElement('option')

    domEl.setAttribute('value', choice.value)
    domEl.innerText = choice.label

    return domEl
  }
}

window.customElements.define('argument-form', ArgumentForm, { extends: 'form' })
