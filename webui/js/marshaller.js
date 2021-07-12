import './ActionButton.js' // To define action-button

export function marshalActionButtonsJsonToHtml (json) {
  const currentIterationTimestamp = Date.now()

  for (const jsonButton of json.actions) {
    var htmlButton = document.querySelector('#actionButton_' + jsonButton.id)

    if (htmlButton == null) {
      htmlButton = document.createElement('button', { is: 'action-button' })
      htmlButton.constructFromJson(jsonButton)

      document.getElementById('rootGroup').appendChild(htmlButton)
    } else {
      htmlButton.updateFromJson(jsonButton)
      htmlButton.updateHtml()
    }

    console.log("action", jsonButton.title)
    htmlButton.updateIterationTimestamp = currentIterationTimestamp;
  }

  for (const existingButton of document.querySelectorAll('button')) {
    if (existingButton.updateIterationTimestamp != currentIterationTimestamp) {
      existingButton.remove();
    }
  }
}
