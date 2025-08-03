'use strict'

import { createClient } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'

import { OliveTinApiService } from './resources/scripts/gen/olivetin/api/v1/olivetin_pb'

import { createApp } from 'vue'
import router from './resources/vue/router.js';
import App from './resources/vue/App.vue';

import {
  initMarshaller,
  setupSectionNavigation,
  marshalDashboardComponentsJsonToHtml,
  refreshServerConnectionLabel
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


function refreshLoop () {
  checkWebsocketConnection()
//  fetchGetDashboardComponents()
//  fetchGetLogs()
  refreshServerConnectionLabel()
}

async function fetchGetDashboardComponents () {
  try {
    const res = await window.client.getDashboardComponents()

    marshalDashboardComponentsJsonToHtml(res)

    refreshServerConnectionLabel() // in-case it changed, update the label quicker
  } catch(err) {
    window.showBigError('fetch-buttons', 'getting buttons', err, false)
  }
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

  processAdditionalLinks(settings.AdditionalLinks)

  window.settings = settings
}

function processAdditionalLinks (links) {
  if (links === null) {
    return
  }

  if (links.length > 0) {
    for (const link of links) {
      const linkA = document.createElement('a')
      linkA.href = link.Url
      linkA.innerText = link.Title

      if (link.Target === '') {
        linkA.target = '_blank'
      } else {
        linkA.target = link.Target
      }

      const linkLi = document.createElement('li')
      linkLi.appendChild(linkA)

      document.getElementById('supplemental-links').prepend(linkLi)
    }
  }
}

function initClient () {
  const transport = createConnectTransport({
    baseUrl: window.location.protocol + '//' + window.location.host + '/api/',

  })

  window.client = createClient(OliveTinApiService, transport)
}

function setupVue () {
  const app = createApp(App)

  app.use(router);
  app.mount('#app')
}

function main () {
  setupVue();

  initClient() 

  initMarshaller()

  window.addEventListener('EventConfigChanged', fetchGetDashboardComponents)
  window.addEventListener('EventEntityChanged', fetchGetDashboardComponents)

  window.fetch('webUiSettings.json').then(res => {
    return res.json()
  }).then(res => {
    processWebuiSettingsJson(res)

    fetchGetDashboardComponents()

    window.restAvailable = true
    window.refreshLoop = refreshLoop
    window.refreshLoop()

    setInterval(refreshLoop, 3000)
  }).catch(err => {
    window.showBigError('fetch-webui-settings', 'getting webui settings', err)
  })
}

main() // call self
