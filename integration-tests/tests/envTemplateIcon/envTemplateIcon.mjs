import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'
import { By, Condition } from 'selenium-webdriver'
import * as process from 'node:process'
import {
  getRootAndWait,
  takeScreenshotOnFailure,
} from '../../lib/elements.js'

describe('config: envTemplateIcon', function () {
  before(async function () {
    // Set the environment variable before starting the runner
    process.env.ADGUARD_ICON = 'test.png'

    await runner.start('envTemplateIcon')
  })

  after(async () => {
    await runner.stop()

    // Clean up the environment variable
    delete process.env.ADGUARD_ICON
  })

  afterEach(function () {
    takeScreenshotOnFailure(this.currentTest, webdriver);
  });

  it('Action icon is set from .Env template variable', async function () {
    await getRootAndWait()

    // Get the dashboard data which contains the actions
    // executeScript automatically waits for Promises returned from the script
    const dashboardResponse = await webdriver.executeScript(`
      return window.client.getDashboard({ title: 'Actions' })
    `)

    expect(dashboardResponse).to.not.be.null
    expect(dashboardResponse).to.have.own.property('dashboard')
    expect(dashboardResponse.dashboard).to.have.own.property('contents')
    expect(dashboardResponse.dashboard.contents).to.be.an('array')
    expect(dashboardResponse.dashboard.contents.length).to.be.greaterThan(0)

    // Actions are nested in dashboard contents - find the action with the expected title
    // The structure is: dashboard.contents[] -> component.contents[] -> action
    let testAction = null
    for (const component of dashboardResponse.dashboard.contents) {
      if (component.contents && Array.isArray(component.contents)) {
        for (const subcomponent of component.contents) {
          if (subcomponent.title === 'Test Action with Env Icon') {
            testAction = subcomponent
            break
          }
        }
      }
      // Also check if the component itself is the action
      if (component.title === 'Test Action with Env Icon') {
        testAction = component
        break
      }
      if (testAction) break
    }

    expect(testAction).to.not.be.null
    expect(testAction).to.have.own.property('icon')
    expect(testAction.icon).to.equal('test.png')
  })

  it('Action button displays icon from .Env template variable', async function () {
    await getRootAndWait()

    // Wait for the action button to be rendered
    await webdriver.wait(new Condition('wait for action button', async () => {
      const btns = await webdriver.findElements(By.css('[title="Test Action with Env Icon"]'))
      return btns.length === 1
    }), 10000)

    // Verify the button exists
    const buttons = await webdriver.findElements(By.css('[title="Test Action with Env Icon"]'))
    expect(buttons).to.have.length(1)

    // The icon should be rendered in the button - we can check via the GetDashboard API response
    // which is the most reliable way to verify the template parsing worked
    // executeScript automatically waits for Promises returned from the script
    const dashboardResponse = await webdriver.executeScript(`
      return window.client.getDashboard({ title: 'Actions' })
    `)

    expect(dashboardResponse).to.not.be.null
    expect(dashboardResponse).to.have.own.property('dashboard')

    // Find the action with the expected title
    // The structure is: dashboard.contents[] -> component.contents[] -> action
    let testAction = null
    for (const component of dashboardResponse.dashboard.contents) {
      if (component.contents && Array.isArray(component.contents)) {
        for (const subcomponent of component.contents) {
          if (subcomponent.title === 'Test Action with Env Icon') {
            testAction = subcomponent
            break
          }
        }
      }
      // Also check if the component itself is the action
      if (component.title === 'Test Action with Env Icon') {
        testAction = component
        break
      }
      if (testAction) break
    }

    expect(testAction).to.not.be.null
    expect(testAction).to.have.own.property('icon')
    expect(testAction.icon).to.equal('test.png')
  })
})
