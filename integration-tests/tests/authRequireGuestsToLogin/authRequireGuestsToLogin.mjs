import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'
import { By } from 'selenium-webdriver'
import {
  takeScreenshotOnFailure,
} from '../../lib/elements.js'

describe('config: authRequireGuestsToLogin', function () {
  this.timeout(30000)

  before(async function () {
    await runner.start('authRequireGuestsToLogin')
  })

  after(async () => {
    await runner.stop()
  })

  afterEach(function () {
    takeScreenshotOnFailure(this.currentTest, webdriver);
  });

  it('Guest is redirected to login', async function () {
    // Don't use getRootAndWait here because we want to test the redirect, and getRootAndWait waits for the dashboard to load

    await webdriver.get(runner.baseUrl())

    await webdriver.wait(async () => {
      const loginRequired = await webdriver.executeScript(
        'return !!(window.initResponse && window.initResponse.loginRequired)'
      )
      const url = await webdriver.getCurrentUrl()
      return loginRequired && url.includes('/login')
    }, 15000, 'Guest should be redirected to login after Init')

    // Verify login UI elements are present
    const loginElements = await webdriver.findElements(By.css('form.local-login-form, .login-oauth2, .login-disabled'))
    expect(loginElements.length).to.be.greaterThan(0)

    console.log('✓ Login page loaded correctly')
  })

  it('Guest does not receive navigation or dashboards before login', async function () {
    await webdriver.get(runner.baseUrl())

    await webdriver.wait(async () => {
      return await webdriver.executeScript(
        'return !!(window.initResponse && window.initResponse.loginRequired)'
      )
    }, 15000, 'Init should report loginRequired')

    const initState = await webdriver.executeScript(`
      const init = window.initResponse || {}
      return {
        showNavigation: !!init.showNavigation,
        rootDashboards: init.rootDashboards || [],
        rootDashboardEntries: init.rootDashboardEntries || []
      }
    `)

    expect(initState.showNavigation).to.equal(false)
    expect(initState.rootDashboards).to.deep.equal([])
    expect(initState.rootDashboardEntries).to.deep.equal([])

    const sidebars = await webdriver.findElements(By.css('#mainnav, #picocrank-sidebar'))
    expect(sidebars.length).to.equal(0)

    console.log('✓ Navigation and dashboards hidden while login is required')
  })
})
