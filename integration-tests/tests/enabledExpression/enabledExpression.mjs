import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'
import { By, until, Condition } from 'selenium-webdriver'
import {
  getRootAndWait,
  takeScreenshotOnFailure,
} from '../../lib/elements.js'

describe('config: enabledExpression', function () {
  this.timeout(30000) // Increase timeout for async operations

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
    // Use the path with space encoded as %20 - Vue Router should decode it
    await webdriver.get(runner.baseUrl() + '/dashboards/LightDashboard')

    // Wait for the URL to change and the route to be processed
    await webdriver.wait(new Condition('wait for URL to contain dashboards', async function() {
      const url = await webdriver.getCurrentUrl()
      return url.includes('/dashboards/')
    }), 5000)

    // Wait for dashboard to load by checking the loaded-dashboard attribute
    // The attribute should be set to the decoded title "LightDashboard"
    await webdriver.wait(new Condition('wait for loaded-dashboard', async function() {
      const body = await webdriver.findElement(By.tagName('body'))
      const attr = await body.getAttribute('loaded-dashboard')
      if (attr) {
        console.log('Current loaded-dashboard attribute:', attr)
      }
      // Accept either decoded or encoded version (component should decode, but handle both)
      return attr === 'LightDashboard'
    }), 10000)

    // Verify we got the correct dashboard (prefer decoded, but accept encoded)
    const body = await webdriver.findElement(By.tagName('body'))
    const attr = await body.getAttribute('loaded-dashboard')
    if (attr !== 'LightDashboard') {
      const currentUrl = await webdriver.getCurrentUrl()
      throw new Error(`Dashboard not loaded correctly. Expected "LightDashboard", got "${attr}". Current URL: ${currentUrl}`)
    }

    // Wait for dashboard content to appear - check for dashboard rows first
    await webdriver.wait(until.elementsLocated(By.css('.dashboard-row')), 5000)

    // Debug: Check what's on the page
    const dashboardRows = await webdriver.findElements(By.css('.dashboard-row'))
    console.log(`Found ${dashboardRows.length} dashboard rows`)
    
    for (let i = 0; i < dashboardRows.length; i++) {
      const row = dashboardRows[i]
      const h2Elements = await row.findElements(By.css('h2'))
      if (h2Elements.length > 0) {
        const h2Text = await h2Elements[0].getText()
        console.log(`Row ${i} h2: "${h2Text}"`)
      }
      const fieldsets = await row.findElements(By.css('fieldset'))
      console.log(`Row ${i} has ${fieldsets.length} fieldsets`)
      if (fieldsets.length > 0) {
        const buttons = await fieldsets[0].findElements(By.css('.action-button button'))
        console.log(`Row ${i} fieldset has ${buttons.length} buttons`)
      }
    }

    // Find buttons by looking within entity fieldsets
    // Both rows have h2 title "Light Controls", so we identify them by which buttons are enabled
    // Living Room Light (powered_on: false) - Turn On should be enabled, Turn Off disabled
    // Bedroom Light (powered_on: true) - Turn Off should be enabled, Turn On disabled
    let turnOnButton = null
    let turnOffButton = null
    
    for (const row of dashboardRows) {
      // Get the fieldset in this row
      const fieldsets = await row.findElements(By.css('fieldset'))
      if (fieldsets.length === 0) continue
      
      const buttons = await fieldsets[0].findElements(By.css('.action-button button'))
      
      // Check each button to identify which entity this row represents
      for (const btn of buttons) {
        const title = await btn.getAttribute('title')
        const disabled = await btn.getAttribute('disabled')
        const isEnabled = disabled === null
        
        if (title === 'Turn On Light' && isEnabled) {
          // This is the Living Room Light row (Turn On is enabled because powered_on: false)
          turnOnButton = btn
        }
        
        if (title === 'Turn Off Light' && isEnabled) {
          // This is the Bedroom Light row (Turn Off is enabled because powered_on: true)
          turnOffButton = btn
        }
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
