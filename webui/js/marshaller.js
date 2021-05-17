import './ActionButton.js' // To define action-button

export function marshalActionButtonsJsonToHtml (json) {
  for (const jsonButton of json.actions) {
    const a = document.createElement('button', { is: 'action-button' })
    a.constructFromJson(jsonButton)

    document.getElementById('rootGroup').appendChild(a)
  }
}
