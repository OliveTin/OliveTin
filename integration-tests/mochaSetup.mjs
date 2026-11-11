import { Options } from 'selenium-webdriver/chrome.js'
import { Builder, Browser } from 'selenium-webdriver'
import getRunner from './runner.mjs'

export async function mochaGlobalSetup () {
  const options = new Options()
  options.addArguments('--headless', '--disable-gpu', '--disable-dev-shm-usage')

  // Logging in with a test password can raise Chrome's save-password or
  // breached-password prompt, which swallows clicks for every later test.
  options.setUserPreferences({
    credentials_enable_service: false,
    'profile.password_manager_enabled': false,
    'profile.password_manager_leak_detection': false,
  })

  global.webdriver = await new Builder().forBrowser(Browser.CHROME).setChromeOptions(options).build()
  await global.webdriver.manage().setTimeouts({ implicit: 0, pageLoad: 30000, script: 30000 })

  global.runner = getRunner()

  console.log('Runner constructor: ' + global.runner.constructor.name)
}

export async function mochaGlobalTeardown () {
  await global.webdriver.quit()
}
