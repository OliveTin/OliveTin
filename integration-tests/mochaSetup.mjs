import { Options } from 'selenium-webdriver/chrome.js'
import { Builder, Browser } from 'selenium-webdriver'
import getRunner from './runner.mjs'

export async function mochaGlobalSetup () {
  const options = new Options()
  options.addArguments('--headless')

  global.webdriver = await new Builder().forBrowser(Browser.CHROME).setChromeOptions(options).build()

  global.runner = getRunner()
}

export async function mochaGlobalTeardown () {
  await global.webdriver.quit()
}
