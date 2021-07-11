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

function processWebuiSettingsJson (settings) {
  window.restBaseUrl = settings.Rest

  if (settings.ThemeName) {
    var themeCss = document.createElement('link')
    themeCss.setAttribute('rel', 'stylesheet');
    themeCss.setAttribute('type', 'text/css')
    themeCss.setAttribute('href', '/themes/' + settings.ThemeName + '/theme.css');

    document.head.appendChild(themeCss);
  }
}

window.fetch('webUiSettings.json').then(res => {
  return res.json()
}).then(res => {
  processWebuiSettingsJson(res)

  fetchGetButtons()

  window.buttonInterval = setInterval(fetchGetButtons, 3000);
}).catch(err => {
  showBigError('fetch-webui-settings', 'getting webui settings', err)
})
