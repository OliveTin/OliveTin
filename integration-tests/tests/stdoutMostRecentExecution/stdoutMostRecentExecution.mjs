import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'
import { By, Condition } from 'selenium-webdriver'
import {
  getRootAndWait,
  getActionButtons,
  takeScreenshotOnFailure,
} from '../../lib/elements.js'

describe('config: stdout-most-recent-execution', function () {
  before(async function () {
    await runner.start('stdoutMostRecentExecution')
  })

  after(async () => {
    await runner.stop()
  })

  afterEach(function () {
    takeScreenshotOnFailure(this.currentTest, webdriver)
  })

  it('stdout-most-recent-execution component is rendered', async function () {
    await getRootAndWait()

    const title = await webdriver.getTitle()
    expect(title).to.be.equal('Test Dashboard - OliveTin')

    // Wait for the mre-output element to appear
    await webdriver.wait(
      new Condition('wait for mre-output element', async () => {
        const elements = await webdriver.findElements(By.css('.mre-output'))
        return elements.length > 0
      }),
      10000
    )

    const mreElements = await webdriver.findElements(By.css('.mre-output'))
    expect(mreElements).to.have.length(1, 'Expected one stdout-most-recent-execution component')
  })

  it('stdout-most-recent-execution displays initial state', async function () {
    await getRootAndWait()

    await webdriver.wait(
      new Condition('wait for mre-output element', async () => {
        const elements = await webdriver.findElements(By.css('.mre-output'))
        return elements.length > 0
      }),
      10000
    )

    const mreElement = await webdriver.findElement(By.css('.mre-output'))
    const text = await mreElement.getText()

    // Should show either "Waiting...", "No execution found", or actual output
    expect(text).to.be.a('string')
    expect(text.length).to.be.greaterThan(0)
  })

  it('stdout-most-recent-execution updates after action execution', async function () {
    this.timeout(45000)
    await getRootAndWait()

    // Wait for the mre-output element
    await webdriver.wait(
      new Condition('wait for mre-output element', async () => {
        const elements = await webdriver.findElements(By.css('.mre-output'))
        return elements.length > 0
      }),
      10000
    )

    const mreElement = await webdriver.findElement(By.css('.mre-output'))
    const initialText = await mreElement.getText()

    // Find the "Check status" action button (button text is the action title, not ID)
    await webdriver.wait(
      new Condition('wait for Check status button', async () => {
        const buttons = await webdriver.findElements(By.css('.action-button button'))
        for (const btn of buttons) {
          const text = await btn.getText()
          if (text.includes('Check status')) {
            return true
          }
        }
        return false
      }),
      10000
    )

    const buttons = await webdriver.findElements(By.css('.action-button button'))
    let statusButton = null
    for (const btn of buttons) {
      const text = await btn.getText()
      if (text.includes('Check status')) {
        statusButton = btn
        break
      }
    }
    expect(statusButton).to.not.be.null

    // Click the button to execute the action
    await statusButton.click()

    // Wait a moment for the action to start
    await webdriver.sleep(2000)

    // Wait for the output to update (the component listens to EventExecutionFinished events)
    // We'll wait for the output to change from the initial state
    await webdriver.wait(
      new Condition('wait for output to update after execution', async () => {
        try {
          const mreElement = await webdriver.findElement(By.css('.mre-output'))
          const newText = await mreElement.getText()
          // Output should change from initial state and contain actual output
          // (not "Waiting...", "No execution found", or the same as initialText)
          const hasChanged = newText !== initialText
          const hasValidOutput = newText && 
                                 !newText.includes('Waiting...') && 
                                 !newText.includes('No execution found') && 
                                 !newText.includes('Error:') &&
                                 newText.trim().length > 0
          return hasChanged && hasValidOutput
        } catch (e) {
          return false
        }
      }),
      20000
    )

    const updatedMreElement = await webdriver.findElement(By.css('.mre-output'))
    const updatedText = await updatedMreElement.getText()

    // The date command should produce output, so verify it's not empty and not an error state
    expect(updatedText).to.not.include('Waiting...')
    expect(updatedText).to.not.include('No execution found')
    expect(updatedText.trim().length).to.be.greaterThan(0)
  })
})
