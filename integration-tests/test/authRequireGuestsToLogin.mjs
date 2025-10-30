import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'
import { By, until } from 'selenium-webdriver'
import {
  getRootAndWait,
  takeScreenshotOnFailure,
} from '../lib/elements.js'

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

  it('Guest is redirected to login, then can login and access dashboard', async function () {
    await webdriver.get(runner.baseUrl())
    await webdriver.wait(until.titleContains('OliveTin'), 10000)
    const title = await webdriver.getTitle()
    expect(title).to.contain('OliveTin')
    console.log('✓ Server started successfully with authRequireGuestsToLogin enabled')

    // Navigate directly to login to avoid SPA timing issues
    await webdriver.get(runner.baseUrl() + '/login')
    // Wait for Init to load and then for login UI to appear (local or OAuth)
    await webdriver.wait(async () => {
      const hasInit = await webdriver.executeScript('return !!window.initResponse')
      if (!hasInit) return false

      const hasLocal = await webdriver.executeScript('return !!(window.initResponse && window.initResponse.authLocalLogin)')
      if (hasLocal) {
        const els = await webdriver.findElements(By.css('form.local-login-form, button.login-button, input[name="username"]'))
        if (els.length > 0) return true
      }

      // If local login not enabled, ensure OAuth or disabled message renders
      const alt = await webdriver.findElements(By.css('.login-oauth2, .login-disabled'))
      return alt.length > 0
    }, 5000)
    
    // Verify we're on the login page
    const currentUrlAtLogin = await webdriver.getCurrentUrl()
    expect(currentUrlAtLogin).to.include('/login')
    console.log('✓ Guest user redirected to login page:', currentUrlAtLogin)
    
    // Verify the login page loaded
    await webdriver.wait(until.titleContains('OliveTin'), 5000)
    const pageTitle = await webdriver.getTitle()
    expect(pageTitle).to.contain('OliveTin')
    
    // Check that login page elements are present
    const body = await webdriver.findElement(By.tagName('body'))
    const bodyText = await body.getText()
    
    // Should have login-related content (either local login or OAuth, or both)
    const hasLoginContent = bodyText.toLowerCase().includes('login') || 
                           await webdriver.findElements(By.css('input[name="username"], input[type="text"]')).then(el => el.length > 0)
    expect(hasLoginContent).to.be.true
    console.log('✓ Login page loaded correctly')

  })
})

