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

import { initWebsocket } from './js/websocket.js'
import combinedTranslations from '../lang/combined_output.json'

function getSelectedLanguage() {
  const storedLanguage = localStorage.getItem('olivetin-language');

  if (storedLanguage && storedLanguage !== 'auto') {
    return storedLanguage;
  }

  if (storedLanguage === 'auto') {
    localStorage.removeItem('olivetin-language');
  }

  if (navigator.languages && navigator.languages.length > 0) {
    const available = Object.keys(combinedTranslations.messages || {})

    for (const candidate of navigator.languages) {
      const lowerCandidate = candidate.toLowerCase()
      const exact = available.find(locale => locale.toLowerCase() === lowerCandidate)

      if (exact) {
        return exact
      }

      const prefix = available.find(locale => locale.toLowerCase().startsWith(lowerCandidate.split('-')[0] + '-'))

      if (prefix) {
        return prefix
      }
    }
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
  
  window.i18n = i18nSettings.global
  
  app.mount('#app')
}

async function main () {
  const i18nSettings = await initClient()

  initWebsocket()

  setupVue(i18nSettings)
}

main()
