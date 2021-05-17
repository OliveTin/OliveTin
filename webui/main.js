'use strict'

import { marshalActionButtonsJsonToHtml } from './js/marshaller.js'

function showBigError (type, friendlyType, message) {
  console.error('Error ' + type + ': ', message)

  const domErr = document.createElement('div')
  domErr.classList.add('error')
  domErr.innerHTML = '<h1>Error ' + friendlyType + '</h1><p>' + message + "</p><p><a href = 'http://github.com/jamesread/OliveTin' target = 'blank'/>OliveTin Documentation</a></p>"

  document.getElementById('rootGroup').appendChild(domErr)
}

function onInitialLoad (res) {
  window.restBaseUrl = res.Rest

  window.fetch(window.restBaseUrl + 'GetButtons', {
    cors: 'cors'
    // No fetch options
  }).then(res => {
    return res.json()
  }).then(res => {
    marshalActionButtonsJsonToHtml(res)
  }).catch(err => {
    showBigError('fetch-initial-buttons', 'getting initial buttons', err, 'blat')
  })
}

window.fetch('webUiSettings.json').then(res => {
  return res.json()
}).then(res => {
  onInitialLoad(res)
}).catch(err => {
  showBigError('fetch-webui-settings', 'getting webui settings', err)
})
