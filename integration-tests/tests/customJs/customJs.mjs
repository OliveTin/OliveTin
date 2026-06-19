import { describe, it, before, after, afterEach } from 'mocha'
import { expect } from 'chai'
import { By } from 'selenium-webdriver'
import {
  getRootAndWait,
  takeScreenshotOnFailure,
} from '../../lib/elements.js'

describe('config: customJs', function () {
  before(async function () {
    await runner.start('customJs')
  })

  after(async () => {
    await runner.stop()
  })

  afterEach(function () {
    takeScreenshotOnFailure(this.currentTest, webdriver)
  })

  it('loads custom.js when enableCustomJs is true (#803)', async function () {
    await getRootAndWait()

    await webdriver.wait(async () => {
      return await webdriver.executeScript('return window.olivetinCustomJsLoaded === true')
    }, 5000, 'custom.js should set window.olivetinCustomJsLoaded when enableCustomJs is enabled')

    const loaded = await webdriver.executeScript('return window.olivetinCustomJsLoaded === true')
    expect(loaded).to.equal(true)

    const scripts = await webdriver.findElements(By.css('#olivetin-custom-js'))
    expect(scripts).to.have.length(1, 'custom.js script tag should be injected into the page')
  })
})
