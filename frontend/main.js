'use strict'

import 'picocrank/styles.css'
import 'femtocrank/dark.css'
import './style.css'

import 'iconify-icon'

import { createClient } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'

import { OliveTinApiService } from './resources/scripts/gen/olivetin/api/v1/olivetin_pb'

import { createApp, h } from 'vue'
import { createI18n } from 'vue-i18n'

import router from './resources/vue/router.js'
import App from './resources/vue/App.vue'

import { initWebsocket } from './js/websocket.js'
import { getSelectedLocale, loadInitialMessages } from './resources/vue/i18n.js'
import { applyThemeStyles, getStoredThemePreference } from './resources/vue/utils/themeLoader.js'

async function initClient () {
  const transport = createConnectTransport({
    baseUrl: window.location.protocol + '//' + window.location.host + '/api/'
  })
  const locale = getSelectedLocale()

  window.client = createClient(OliveTinApiService, transport)
  const [initResponse, messages] = await Promise.all([
    window.client.init({}),
    loadInitialMessages(locale)
  ])
  window.initResponse = initResponse

  if (window.initResponse.enableCustomJs) {
    const script = document.createElement('script')
    script.src = '/custom-webui/custom.js'
    script.async = true
    script.id = 'olivetin-custom-js'
    document.head.appendChild(script)
  }

  const i18nSettings = createI18n({
    legacy: false,
    locale,
    fallbackLocale: 'en',
    messages,
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

function setupErrorDisplay (errorMessage) {
  const ErrorApp = {
    render () {
      return h('section', { class: 'bad', style: 'padding: 2em; text-align: center; margin: 2em auto;' }, [
        h('h2', 'OliveTin Init Failed'),
        h('p', errorMessage),
        h('p', 'Please check your browser console for more details.')
      ])
    }
  }

  const app = createApp(ErrorApp)
  app.mount('#app')
}

async function main () {
  try {
    const i18nSettings = await initClient()

    try {
      await applyThemeStyles(getStoredThemePreference())
    } catch (err) {
      console.warn('Failed to load theme CSS:', err)
    }

    initWebsocket()

    setupVue(i18nSettings)
  } catch (err) {
    const errorMessage = err.message || 'Failed to initialize. Please check your configuration and try again.'
    console.error('Init failed:', err)
    setupErrorDisplay(errorMessage)
  }
}

main()
