import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'
import { By, Condition } from 'selenium-webdriver'
import {
  DEFAULT_UI_WAIT_MS,
  getRootAndWait,
  getActionButton,
  takeScreenshotOnFailure,
  waitForArgumentFormPage,
  waitForArgumentFormReady,
  waitForDashboardLoaded,
} from '../../lib/elements.js'

async function waitForStartButtonEnabled () {
  await webdriver.wait(
    new Condition('wait for Start button to be enabled', async () => {
      const submitButton = await webdriver.findElement(By.css('button[name="start"]'))
      return await submitButton.isEnabled()
    }),
    DEFAULT_UI_WAIT_MS
  )
}

async function waitForActionSuccessFlash (actionTitle) {
  await webdriver.wait(
    new Condition(`wait for ${actionTitle} success flash`, async () => {
      try {
        const button = await getActionButton(webdriver, actionTitle)
        const classAttr = await button.getAttribute('class')
        return classAttr && classAttr.includes('action-success')
      } catch (e) {
        return false
      }
    }),
    DEFAULT_UI_WAIT_MS
  )
}

describe('config: argumentActionFlash', function () {
  this.timeout(15000)

  before(async function () {
    await runner.start('argumentActionFlash')
  })

  after(async () => {
    await runner.stop()
  })

  afterEach(function () {
    takeScreenshotOnFailure(this.currentTest, webdriver)
  })

  it('Action with arguments flashes success after returning to dashboard (#920)', async function () {
    await getRootAndWait()

    const argButton = await getActionButton(webdriver, 'Hello world')
    await argButton.click()

    await waitForArgumentFormPage()
    await waitForArgumentFormReady()
    await waitForStartButtonEnabled()

    const submitButton = await webdriver.findElement(By.css('button[name="start"]'))
    await submitButton.click()

    await webdriver.wait(
      new Condition('wait to leave argument form', async () => {
        const url = await webdriver.getCurrentUrl()
        return !url.includes('/argumentForm')
      }),
      DEFAULT_UI_WAIT_MS
    )
    await waitForDashboardLoaded()

    await waitForActionSuccessFlash('Hello world')

    const flashedButton = await getActionButton(webdriver, 'Hello world')
    const classAttr = await flashedButton.getAttribute('class')
    expect(classAttr).to.include('action-success')
  })

  it('Action without arguments still flashes success on the dashboard', async function () {
    await getRootAndWait()

    const simpleButton = await getActionButton(webdriver, 'Simple action')
    await simpleButton.click()

    await waitForActionSuccessFlash('Simple action')

    const flashedButton = await getActionButton(webdriver, 'Simple action')
    const classAttr = await flashedButton.getAttribute('class')
    expect(classAttr).to.include('action-success')
  })
})
