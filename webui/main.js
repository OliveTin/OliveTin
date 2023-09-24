'use strict'

import { marshalActionButtonsJsonToHtml, marshalLogsJsonToHtml } from './js/marshaller.js'
import { checkWebsocketConnection } from './js/websocket.js'

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

function refreshLoop () {
  if (window.websocketAvailable) {
    document.querySelector('#serverConnection').classList.remove('error')
    document.querySelector('#serverConnection').innerText = 'websocket'

    // Websocket updates are streamed live, not updated on a loop.
  } else if (window.restAvailable) {
    // Fallback to rest, but try to reconnect the websocket anyway.

    fetchGetDashboardComponents()
    checkWebsocketConnection()

    document.querySelector('#serverConnection').classList.remove('error')
    document.querySelector('#serverConnection').innerText = 'rest'

    fetchGetLogs()
  } else {
    document.querySelector('#serverConnection').innerText = 'disconnected, trying to reconnect...'
    document.querySelector('#serverConnection').classList.add('error')

    // Still try to fetch the dashboard, if successfull window.restAvailable = true
    fetchGetDashboardComponents()
  }
}

function fetchGetDashboardComponents () {
  window.fetch(window.restBaseUrl + 'GetDashboardComponents', {
    cors: 'cors'
  }).then(res => {
    return res.json()
  }).then(res => {
    window.restAvailable = true
    marshalActionButtonsJsonToHtml(res)
  }).catch(() => { // err is 1st arg
    window.restAvailable = false
    //    window.showBigError('fetch-buttons', 'getting buttons', err, 'blat')
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

  document.querySelector('#currentVersion').innerText = settings.CurrentVersion

  if (settings.ShowNewVersions && settings.AvailableVersion !== 'none') {
    document.querySelector('#available-version').innerText = 'New Version Available: ' + settings.AvailableVersion
    document.querySelector('#available-version').hidden = false
  }

  document.querySelector('#perma-widget').hidden = !settings.ShowNavigation
  document.querySelector('footer[title="footer"]').hidden = !settings.ShowFooter

  if (settings.PageTitle) {
    document.title = settings.PageTitle
    const titleElem = document.querySelector('#page-title')
    if (titleElem) titleElem.innerText = settings.PageTitle
  }
}

function main () {
  setupSections()

  window.fetch('webUiSettings.json').then(res => {
    return res.json()
  }).then(res => {
    processWebuiSettingsJson(res)

    window.restAvailable = true
    window.refreshLoop = refreshLoop
    window.refreshLoop()

    setInterval(refreshLoop, 3000)
  }).catch(err => {
    window.showBigError('fetch-webui-settings', 'getting webui settings', err)
  })
}

main() // call self
