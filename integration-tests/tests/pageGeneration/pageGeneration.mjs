import { describe, it, after, afterEach } from 'mocha'
import { expect } from 'chai'
import {
  getRootAndWait,
  takeScreenshotOnFailure,
} from '../../lib/elements.js'

describe('runner: pageGeneration', function () {
  after(async () => {
    await runner.stop()
  })

  afterEach(function () {
    takeScreenshotOnFailure(this.currentTest, webdriver)
  })

  it('reloads the dashboard after sequential config changes', async function () {
    await runner.start('pageGenerationFirst')
    await getRootAndWait()

    let loadedDashboard = await webdriver.executeScript(
      'return document.body.getAttribute("loaded-dashboard")'
    )
    expect(loadedDashboard).to.equal('Page Generation First')

    await runner.start('pageGenerationSecond')
    await getRootAndWait()

    loadedDashboard = await webdriver.executeScript(
      'return document.body.getAttribute("loaded-dashboard")'
    )
    expect(loadedDashboard).to.equal('Page Generation Second')
  })
})
