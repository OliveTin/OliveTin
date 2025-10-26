'use strict'

import 'femtocrank/style.css'
import 'femtocrank/dark.css'
import './style.css'

import 'iconify-icon'

import { createClient } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'

import { OliveTinApiService } from './resources/scripts/gen/olivetin/api/v1/olivetin_pb'

import { createApp } from 'vue'
import router from './resources/vue/router.js'
import App from './resources/vue/App.vue'

import {
  initMarshaller
} from './js/marshaller.js'

import { checkWebsocketConnection } from './js/websocket.js'

function initClient () {
  const transport = createConnectTransport({
    baseUrl: window.location.protocol + '//' + window.location.host + '/api/'

  })

  window.client = createClient(OliveTinApiService, transport)
}

function setupVue () {
  const app = createApp(App)

  app.use(router)
  app.mount('#app')
}

function main () {
  initClient()

  // Expose websocket connection function globally so App.vue can call it after successful init
  window.checkWebsocketConnection = checkWebsocketConnection

  setupVue()

  initMarshaller()

//  window.addEventListener('EventConfigChanged', fetchGetDashboardComponents)
//  window.addEventListener('EventEntityChanged', fetchGetDashboardComponents)
}

main() // call self
