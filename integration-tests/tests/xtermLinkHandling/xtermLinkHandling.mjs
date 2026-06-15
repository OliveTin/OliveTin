import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'
import { By, Condition } from 'selenium-webdriver'
import {
  DEFAULT_UI_WAIT_MS,
  getRootAndWait,
  takeScreenshotOnFailure,
  getTerminalBuffer,
  waitForLogsPage,
  waitForExecutionComplete,
} from '../../lib/elements.js'

describe('config: xtermLinkHandling', function () {
  before(async function () {
    await runner.start('xtermLinkHandling')
  })

  after(async () => {
    await runner.stop()
  })

  afterEach(function () {
    takeScreenshotOnFailure(this.currentTest, webdriver)
  })

  it('xterm output shows URL and link handling is configured', async function () {
    await getRootAndWait()

    await webdriver.wait(new Condition('wait for Echo URL button', async () => {
      const btns = await webdriver.findElements(By.css('[title="Echo URL"]'))
      return btns.length === 1
    }), DEFAULT_UI_WAIT_MS)

    const echoUrlButton = await webdriver.findElement(By.css('[title="Echo URL"]'))
    await echoUrlButton.click()

    await waitForLogsPage()
    await waitForExecutionComplete()

    const bufferText = await getTerminalBuffer()
    expect(bufferText).to.not.be.null
    expect(bufferText).to.include('https://example.com')

    const linkHandlerSet = await webdriver.executeScript(`
      try {
        return !!(window.terminal && window.terminal.linkHandlerConfigured === true)
      } catch (e) {
        return false
      }
    `)
    expect(linkHandlerSet).to.equal(true)
  })
})
