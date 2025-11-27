import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'
import { By, until } from 'selenium-webdriver'
import {
  getRootAndWait,
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

    await webdriver.wait(until.urlContains('/login'), 10000)
    
    // Verify login UI elements are present
    const loginElements = await webdriver.findElements(By.css('form.local-login-form, .login-oauth2, .login-disabled'))
    expect(loginElements.length).to.be.greaterThan(0)

    console.log('✓ Login page loaded correctly')

  })
})

