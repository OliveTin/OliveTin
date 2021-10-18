
class ArgumentForm extends window.HTMLFormElement {
  setup(json, callback) {
    this.setAttribute('class', 'actionArguments')
    this.title = document.createElement("h1")
    this.title.innerHTML = "Action Arguments"

    this.appendChild(this.title);

    let a = document.createElement("span")
    a.innerText = "Hi"
    frm.appendChild(a)


    console.log(json)
  }
}

window.customElements.define('argument-form', ArgumentForm, { extends: 'form' })
