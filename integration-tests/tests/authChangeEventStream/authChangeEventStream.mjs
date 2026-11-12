import { By, until } from 'selenium-webdriver'
import {
  getRootAndWait,
  takeScreenshotOnFailure,
} from '../../lib/elements.js'

const UI_WAIT_MS = 5000

function actionButtonLocator (title) {
  return By.xpath(`//*[contains(@class, 'action-button')]//button[contains(., '${title}')]`)
}

async function clickActionAndWaitForSuccess (title) {
  const button = await webdriver.wait(until.elementLocated(actionButtonLocator(title)), UI_WAIT_MS)
  await button.click()

  await webdriver.wait(async () => {
    const classes = await button.getAttribute('class')
    return classes.includes('action-success')
  }, UI_WAIT_MS)
}

describe('config: authChangeEventStream', function () {
  before(async function () {
    await runner.start('authChangeEventStream')
  })

  after(async () => {
    await webdriver.manage().deleteAllCookies()
    await runner.stop()
  })

  afterEach(function () {
    takeScreenshotOnFailure(this.currentTest, webdriver)
  })

  it('delivers execution events for the new user after logging in without a page refresh', async () => {
    await getRootAndWait()
    await clickActionAndWaitForSuccess('Guest action')

    const loginLink = await webdriver.wait(until.elementLocated(By.css('a.login-link')), UI_WAIT_MS)
    await loginLink.click()

    const username = await webdriver.wait(until.elementLocated(By.id('username')), UI_WAIT_MS)
    await username.sendKeys('alice')
    await webdriver.findElement(By.id('password')).sendKeys('password')
    await webdriver.findElement(By.css('button.login-button')).click()

    await clickActionAndWaitForSuccess('Admin action')
  })
})
