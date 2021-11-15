'use strict'

import { marshalActionButtonsJsonToHtml, marshalLogsJsonToHtml } from './js/marshaller.js'

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

function fetchGetDashboardComponents () {
  window.fetch(window.restBaseUrl + 'GetDashboardComponents', {
    cors: 'cors'
  }).then(res => {
    return res.json()
  }).then(res => {
    marshalActionButtonsJsonToHtml(res)
  }).catch(err => {
    window.showBigError('fetch-buttons', 'getting buttons', err, 'blat')
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
    window.showBigError('fetch-buttons', 'getting buttons', err, 'blat')
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

  document.querySelector('#currentVersion').innerText = 'Version: ' + settings.CurrentVersion

  if (settings.ShowNewVersions && settings.AvailableVersion !== 'none') {
    document.querySelector('#availableVersion').innerText = 'New Version Available: ' + settings.AvailableVersion
    document.querySelector('#availableVersion').hidden = false
  }

  document.querySelector('#sectionSwitcher').hidden = settings.HideNavigation
}

function main () {
  setupSections()

  window.fetch('webUiSettings.json').then(res => {
    return res.json()
  }).then(res => {
    processWebuiSettingsJson(res)

    fetchGetDashboardComponents()
    fetchGetLogs()

    window.buttonInterval = setInterval(fetchGetDashboardComponents, 3000)
  }).catch(err => {
    window.showBigError('fetch-webui-settings', 'getting webui settings', err)
  })
}

main() // call self
