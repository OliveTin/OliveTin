'use strict'

import 'femtocrank/style.css'
import 'femtocrank/dark.css'
import './style.css'

import 'iconify-icon'

import { createClient } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'

import { OliveTinApiService } from './resources/scripts/gen/olivetin/api/v1/olivetin_pb'

import { createApp } from 'vue'
import { createI18n } from 'vue-i18n'

import router from './resources/vue/router.js'
import App from './resources/vue/App.vue'

import combinedTranslations from '../lang/combined_output.json'

import {
  initMarshaller
} from './js/marshaller.js'

import { checkWebsocketConnection } from './js/websocket.js'

function getSelectedLanguage() {
  const storedLanguage = localStorage.getItem('olivetin-language');

  if (storedLanguage && storedLanguage !== 'auto') {
    return storedLanguage;
  }

  if (storedLanguage === 'auto') {
    localStorage.removeItem('olivetin-language');
  }

  if (navigator.languages && navigator.languages.length > 0) {
    return navigator.languages[0];
  }

  return 'en';
}

async function initClient () {
  const transport = createConnectTransport({
    baseUrl: window.location.protocol + '//' + window.location.host + '/api/'
  })

  window.client = createClient(OliveTinApiService, transport)
  window.initResponse = await window.client.init({})
  
  const i18nSettings = createI18n({
    legacy: false,
    locale: getSelectedLanguage(),
    fallbackLocale: 'en',
    messages: combinedTranslations.messages,
    postTranslation: (translated) => {
      const params = new URLSearchParams(window.location.search)

      if (params.has('debug-translations')) {
        return '____'
      } else {
        return translated
      }
    }
  })

  return i18nSettings
}

function setupVue (i18nSettings) {
  const app = createApp(App)

  app.use(router)
  app.use(i18nSettings)
  
  // Make i18n instance accessible globally for language switching
  window.i18n = i18nSettings.global
  
  app.mount('#app')
}

async function main () {
  window.checkWebsocketConnection = checkWebsocketConnection

  const i18nSettings = await initClient()

  setupVue(i18nSettings)

  initMarshaller()
}

main() // call self
