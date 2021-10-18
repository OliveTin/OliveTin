'use strict'

import { marshalActionButtonsJsonToHtml, marshalLogsJsonToHtml } from './js/marshaller.js'

function showBigError (type, friendlyType, message) {
  clearInterval(window.buttonInterval)

  console.error('Error ' + type + ': ', message)

  const domErr = document.createElement('div')
  domErr.classList.add('error')
  domErr.innerHTML = '<h1>Error ' + friendlyType + '</h1><p>' + message + "</p><p><a href = 'http://docs.olivetin.app/troubleshooting.html' target = 'blank'/>OliveTin Documentation</a></p>"

  document.getElementById('rootGroup').appendChild(domErr)
}

function showSection (name) {
  for (const otherName of ['Actions', 'Logs']) {
    document.getElementById('show' + otherName).classList.remove('activeSection')
    document.getElementById('content' + otherName).hidden = true
  }

  document.getElementById('show' + name).classList.add('activeSection')
  document.getElementById('content' + name).hidden = false
}

function setupSections () {
  document.getElementById('showActions').onclick = () => { showSection('Actions') }
  document.getElementById('showLogs').onclick = () => { showSection('Logs') }

  showSection('Actions')
}

function fetchGetButtons () {
  window.fetch(window.restBaseUrl + 'GetButtons', {
    cors: 'cors'
  }).then(res => {
    return res.json()
  }).then(res => {
    marshalActionButtonsJsonToHtml(res)
  }).catch(err => {
    showBigError('fetch-buttons', 'getting buttons', err, 'blat')
  })
}

function fetchGetLogs () {
  window.fetch(window.restBaseUrl + 'GetLogs', {
    cors: 'cors'
  }).then(res => {
    return res.json()
  }).then(res => {
    marshalLogsJsonToHtml(res)
  }).catch(err => {
    showBigError('fetch-buttons', 'getting buttons', err, 'blat')
  })
}

function processWebuiSettingsJson (settings) {
  window.restBaseUrl = settings.Rest

  if (settings.ThemeName) {
    const themeCss = document.createElement('link')
    themeCss.setAttribute('rel', 'stylesheet')
    themeCss.setAttribute('type', 'text/css')
    themeCss.setAttribute('href', '/themes/' + settings.ThemeName + '/theme.css')

    document.head.appendChild(themeCss)
  }
}

setupSections()

window.fetch('webUiSettings.json').then(res => {
  return res.json()
}).then(res => {
  processWebuiSettingsJson(res)

  fetchGetButtons()
  fetchGetLogs()

  window.buttonInterval = setInterval(fetchGetButtons, 3000)
}).catch(err => {
  showBigError('fetch-webui-settings', 'getting webui settings', err)
})
