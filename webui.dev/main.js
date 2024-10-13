'use strict'

import {
  initMarshaller,
  setupSectionNavigation,
  marshalDashboardComponentsJsonToHtml,
  marshalLogsJsonToHtml,
  refreshServerConnectionLabel,
  refreshDiagnostics
} from './js/marshaller.js'
import { checkWebsocketConnection } from './js/websocket.js'

function searchLogs (e) {
  document.getElementById('searchLogsClear').disabled = false

  const searchText = e.target.value.toLowerCase()

  for (const row of document.querySelectorAll('tr.log-row')) {
    const actionTitle = row.getAttribute('title').toLowerCase()

    row.hidden = !actionTitle.includes(searchText)
  }
}

function searchLogsClear () {
  for (const row of document.querySelectorAll('tr.log-row')) {
    row.hidden = false
  }

  document.getElementById('searchLogsClear').disabled = true
  document.getElementById('logSearchBox').value = ''
}

function setupLogSearchBox () {
  document.getElementById('logSearchBox').oninput = searchLogs
  document.getElementById('searchLogsClear').onclick = searchLogsClear
}

function refreshLoop () {
  if (window.websocketAvailable) {
    // Websocket updates are streamed live, not updated on a loop.
  } else if (window.restAvailable) {
    // Fallback to rest, but try to reconnect the websocket anyway.

    fetchGetDashboardComponents()
    fetchGetLogs()

    checkWebsocketConnection()
  } else {
    // Still try to fetch the dashboard, if successfull window.restAvailable = true
    fetchGetDashboardComponents()
  }

  refreshServerConnectionLabel()
}

function fetchGetDashboardComponents () {
  window.fetch(window.restBaseUrl + 'GetDashboardComponents', {
    cors: 'cors'
  }).then(res => {
    if (!res.ok && res.status === 401) {
      window.location.href = window.settings.AuthLoginUrl
    }
    return res.json()
  }).then(res => {
    if (!window.restAvailable) {
      window.clearBigErrors()
    }

    window.restAvailable = true
    marshalDashboardComponentsJsonToHtml(res)

    refreshServerConnectionLabel() // in-case it changed, update the label quicker
  }).catch((err) => { // err is 1st arg
    window.restAvailable = false
    window.showBigError('fetch-buttons', 'getting buttons', err, false)
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
    window.showBigError('fetch-buttons', 'getting buttons', err, false)
  })
}

function processWebuiSettingsJson (settings) {
  setupSectionNavigation(settings.SectionNavigationStyle)

  window.restBaseUrl = settings.Rest

  document.querySelector('#currentVersion').innerText = settings.CurrentVersion

  if (settings.ShowNewVersions && settings.AvailableVersion !== 'none') {
    document.querySelector('#available-version').innerText = 'New Version Available: ' + settings.AvailableVersion
    document.querySelector('#available-version').hidden = false
  }

  if (!settings.ShowNavigation) {
    document.querySelector('header').style.display = 'none'
  }

  if (!settings.ShowFooter) {
    document.querySelector('footer[title="footer"]').style.display = 'none'
  }

  if (settings.EnableCustomJs) {
    const script = document.createElement('script')
    script.src = './custom-webui/custom.js'
    document.head.appendChild(script)
  }

  window.pageTitle = 'OliveTin'

  if (settings.PageTitle) {
    window.pageTitle = settings.PageTitle

    document.title = window.pageTitle

    const titleElem = document.querySelector('#page-title')
    if (titleElem) titleElem.innerText = window.pageTitle
  }

  processAdditionaLinks(settings.AdditionalLinks)

  window.settings = settings

  refreshDiagnostics()
}

function processAdditionaLinks (links) {
  if (links === null) {
    return
  }

  for (const link of links) {
    const linkA = document.createElement('a')
    linkA.href = link.Url
    linkA.innerText = link.Title
    linkA.target = '_blank'

    const linkLi = document.createElement('li')
    linkLi.appendChild(linkA)

    document.getElementById('supplemental-links').prepend(linkLi)
  }
}

function main () {
  initMarshaller()

  setupLogSearchBox()

  window.addEventListener('EventConfigChanged', fetchGetDashboardComponents)
  window.addEventListener('EventEntityChanged', fetchGetDashboardComponents)

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
