import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'
import {
  getRootAndWait,
  takeScreenshotOnFailure,
} from '../../lib/elements.js'

describe('config: themeLoading', function () {
  before(async function () {
    await runner.start('themeLoading')
  })

  after(async () => {
    await runner.stop()
  })

  afterEach(function () {
    takeScreenshotOnFailure(this.currentTest, webdriver)
  })

  it('Available themes are discovered and returned in Init response', async function () {
    await getRootAndWait()

    // Wait for initResponse to be available
    await webdriver.wait(async () => {
      const hasInitResponse = await webdriver.executeScript(
        'return typeof window.initResponse !== "undefined" && window.initResponse !== null'
      )
      return hasInitResponse
    }, 5000, 'Init response should be available')

    // Get available themes from the Init response
    const availableThemes = await webdriver.executeScript(
      'return window.initResponse ? (window.initResponse.availableThemes || []) : []'
    )

    // Verify themes array exists and is an array
    expect(availableThemes).to.be.an('array')

    // Verify that themes with theme.css are discovered
    // theme-one and theme-two have theme.css, invalid-theme does not
    expect(availableThemes).to.include('theme-one')
    expect(availableThemes).to.include('theme-two')
    
    // Verify that themes without theme.css are not included
    expect(availableThemes).to.not.include('invalid-theme')

    // Verify themes are sorted alphabetically
    const sortedThemes = [...availableThemes].sort()
    expect(availableThemes).to.deep.equal(sortedThemes)
  })

  it('Available themes list is accessible via JavaScript', async function () {
    await getRootAndWait()

    // Wait for initResponse to be available
    await webdriver.wait(async () => {
      const hasInitResponse = await webdriver.executeScript(
        'return typeof window.initResponse !== "undefined" && window.initResponse !== null'
      )
      return hasInitResponse
    }, 5000, 'Init response should be available')

    // Verify availableThemes is accessible
    const availableThemes = await webdriver.executeScript(
      'return window.initResponse ? (window.initResponse.availableThemes || []) : []'
    )

    expect(availableThemes).to.be.an('array')
    expect(availableThemes.length).to.be.at.least(2)
  })
})

