import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'
import { By, until } from 'selenium-webdriver'
import {
  getRootAndWait,
  takeScreenshotOnFailure,
} from '../../lib/elements.js'

describe('config: enabledExpression', function () {
  before(async function () {
    await runner.start('enabledExpression')
  })

  after(async () => {
    await runner.stop()
  })

  afterEach(function () {
    takeScreenshotOnFailure(this.currentTest, webdriver);
  });

  it('Action with enabledExpression for lights enable the correct action', async function() {
    await getRootAndWait()

    // Navigate to the Lights Dashboard
    await webdriver.get(runner.baseUrl() + '/dashboard/Lights%20Dashboard')

    // Wait for dashboard to load
    await webdriver.wait(until.elementLocated(By.css('.action-button')), 10000)

    // Find action buttons
    const actionButtons = await webdriver.findElements(By.css('.action-button button'))

    // Find "Turn On Light" button for "Living Room Light" (powered_on: false, so Turn On should be enabled)
    // Find "Turn Off Light" button for "Bedroom Light" (powered_on: true, so Turn Off should be enabled)
    let turnOnButton = null
    let turnOffButton = null

    for (const btn of actionButtons) {
      const title = await btn.getAttribute('title')
      if (title && title.includes('Turn On Light') && title.includes('Living Room')) {
        turnOnButton = btn
      }
      if (title && title.includes('Turn Off Light') && title.includes('Bedroom')) {
        turnOffButton = btn
      }
    }

    expect(turnOnButton).to.not.be.null
    expect(turnOffButton).to.not.be.null

    // Check that Turn On button is enabled (light is off)
    const turnOnDisabled = await turnOnButton.getAttribute('disabled')
    expect(turnOnDisabled).to.be.null

    // Check that Turn Off button is enabled (light is on)
    const turnOffDisabled = await turnOffButton.getAttribute('disabled')
    expect(turnOffDisabled).to.be.null
  })

  it('Action without enabledExpression is always enabled', async function() {
    await getRootAndWait()

    // Navigate to actions view
    await webdriver.get(runner.baseUrl())

    // Wait for action buttons
    await webdriver.wait(until.elementLocated(By.css('.action-button')), 10000)

    // Find "Always Enabled Action" button
    const actionButtons = await webdriver.findElements(By.css('.action-button button'))
    let alwaysEnabledButton = null

    for (const btn of actionButtons) {
      const title = await btn.getAttribute('title')
      if (title === 'Always Enabled Action') {
        alwaysEnabledButton = btn
        break
      }
    }

    expect(alwaysEnabledButton).to.not.be.null

    // Check that it's enabled
    const disabled = await alwaysEnabledButton.getAttribute('disabled')
    expect(disabled).to.be.null
  })
})
