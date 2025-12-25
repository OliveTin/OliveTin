import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'
import { By, Condition } from 'selenium-webdriver'
import {
  getRootAndWait,
  getActionButton,
  takeScreenshotOnFailure,
  getTerminalBuffer,
} from '../../lib/elements.js'

async function openCheckboxArgumentForm() {
  await getRootAndWait()
  const btn = await getActionButton(webdriver, 'Test checkbox argument')
  await btn.click()

  await webdriver.wait(
    new Condition('wait for argument form page', async () => {
      const url = await webdriver.getCurrentUrl()
      return url.includes('/actionBinding/') && url.includes('/argumentForm')
    }),
    5000
  )
}

async function getCheckboxInput() {
  return await webdriver.findElement(By.id('confirm'))
}

async function submitCheckboxForm() {
  const submitButton = await webdriver.findElement(By.css('button[name="start"]'))
  await submitButton.click()
}

async function waitForLogsPage() {
  await webdriver.wait(
    new Condition('wait for logs page', async () => {
      const url = await webdriver.getCurrentUrl()
      return url.includes('/logs/') && !url.endsWith('/logs')
    }),
    5000
  )
}

async function waitForExecutionComplete() {
  await webdriver.wait(
    new Condition('wait for execution status', async () => {
      const statusElements = await webdriver.findElements(By.id('execution-dialog-status'))
      return statusElements.length > 0
    }),
    5000
  )

  await webdriver.wait(
    new Condition('wait for execution to finish', async () => {
      try {
        const statusElement = await webdriver.findElement(By.id('execution-dialog-status'))
        const statusText = await statusElement.getText()
        return !statusText.includes('Executing')
      } catch (e) {
        return false
      }
    }),
    5000
  )

  // Small delay to allow terminal to write output
  await webdriver.sleep(500)
}

async function waitForTerminalOutput(expectedValue) {
  await webdriver.wait(
    new Condition(`wait for checkbox value ${expectedValue} in output`, async () => {
      try {
        const terminalReady = await webdriver.executeScript(`
          return !!(window.terminal && window.terminal.getBufferAsString);
        `)
        if (!terminalReady) {
          return false
        }
        
        const output = await getTerminalBuffer()
        if (!output) {
          return false
        }
        
        return output.trim().includes(`Checkbox value: ${expectedValue}`)
      } catch (e) {
        return false
      }
    }),
    5000
  )
}

describe('config: checkbox', function () {
  before(async function () {
    await runner.start('checkbox')
  })

  after(async () => {
    await runner.stop()
  })

  afterEach(function () {
    takeScreenshotOnFailure(this.currentTest, webdriver)
  })

  it('Checkbox argument is rendered as a checkbox input', async function () {
    await openCheckboxArgumentForm()

    const checkboxInput = await getCheckboxInput()

    expect(await checkboxInput.getTagName()).to.equal('input')
    expect(await checkboxInput.getAttribute('type')).to.equal('checkbox')

    const label = await webdriver.findElement(By.css('label[for="confirm"]'))
    expect(await label.getText()).to.contain('Confirm option')
  })

  it('Checkbox argument submits 0 by default when unchecked', async function () {
    this.timeout(15000)
    await openCheckboxArgumentForm()

    const checkboxInput = await getCheckboxInput()
    expect(await checkboxInput.isSelected()).to.be.false

    await submitCheckboxForm()
    await waitForLogsPage()
    await waitForExecutionComplete()
    await waitForTerminalOutput('0')
  })

  it('Checkbox argument can be toggled and submitted', async function () {
    this.timeout(15000)
    await openCheckboxArgumentForm()

    const checkboxInput = await getCheckboxInput()
    await checkboxInput.click()
    await webdriver.sleep(100)

    expect(await checkboxInput.isSelected()).to.be.true

    await submitCheckboxForm()
    await waitForLogsPage()
    await waitForExecutionComplete()
    await waitForTerminalOutput('1')
  })
})


