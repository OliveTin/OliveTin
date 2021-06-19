import './ActionButton.js' // To define action-button

export function marshalActionButtonsJsonToHtml (json) {
  for (const jsonButton of json.actions) {
    var htmlButton = document.querySelector('#actionButton_' + jsonButton.id)

    if (htmlButton == null) {
      htmlButton = document.createElement('button', { is: 'action-button' })
      htmlButton.constructFromJson(jsonButton)
      document.getElementById('rootGroup').appendChild(htmlButton)
    } else {
      htmlButton.updateFromJson(jsonButton)
    }
  }
}
