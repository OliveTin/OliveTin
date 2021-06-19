'use strict'

import { marshalActionButtonsJsonToHtml } from './js/marshaller.js'

function showBigError (type, friendlyType, message) {
  clearInterval(window.buttonInterval)

  console.error('Error ' + type + ': ', message)

  const domErr = document.createElement('div')
  domErr.classList.add('error')
  domErr.innerHTML = '<h1>Error ' + friendlyType + '</h1><p>' + message + "</p><p><a href = 'http://olivetin.app/_errors_troubleshooting.html' target = 'blank'/>OliveTin Documentation</a></p>"

  document.getElementById('rootGroup').appendChild(domErr)
}

function fetchGetButtons() {
 window.fetch(window.restBaseUrl + 'GetButtons', {
    cors: 'cors'
    // No fetch options
  }).then(res => {
    return res.json()
  }).then(res => {
    marshalActionButtonsJsonToHtml(res)
  }).catch(err => {
    showBigError('fetch-buttons', 'getting buttons', err, 'blat')
  })
}

function onInitialLoad (res) {
  window.restBaseUrl = res.Rest

  window.buttonInterval = setInterval(fetchGetButtons, 3000);
  fetchGetButtons()
}

window.fetch('webUiSettings.json').then(res => {
  return res.json()
}).then(res => {
  onInitialLoad(res)
}).catch(err => {
  showBigError('fetch-webui-settings', 'getting webui settings', err)
})
