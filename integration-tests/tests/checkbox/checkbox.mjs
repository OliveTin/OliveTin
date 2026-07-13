import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'
import { By, Condition } from 'selenium-webdriver'
import {
  DEFAULT_UI_WAIT_MS,
  getRootAndWait,
  getActionButton,
  takeScreenshotOnFailure,
  getTerminalBuffer,
  waitForArgumentFormPage,
  waitForLogsPage,
  waitForExecutionComplete,
  argumentFieldId,
} from '../../lib/elements.js'

async function openCheckboxArgumentForm() {
  await getRootAndWait()
  const btn = await getActionButton(webdriver, 'Test checkbox argument')
  await btn.click()

  await waitForArgumentFormPage()
}

async function getCheckboxInput() {
  return await webdriver.findElement(By.id(argumentFieldId('confirm')))
}

async function submitCheckboxForm() {
  const submitButton = await webdriver.findElement(By.css('button[name="start"]'))
  await submitButton.click()
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
    DEFAULT_UI_WAIT_MS
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

    const label = await webdriver.findElement(By.css(`label[for="${argumentFieldId('confirm')}"]`))
    expect(await label.getText()).to.contain('Confirm option')
  })

  it('Checkbox argument submits 0 by default when unchecked', async function () {
    await openCheckboxArgumentForm()

    const checkboxInput = await getCheckboxInput()
    expect(await checkboxInput.isSelected()).to.be.false

    await submitCheckboxForm()
    await waitForLogsPage()
    await waitForExecutionComplete()
    await waitForTerminalOutput('0')
  })

  it('Checkbox argument can be toggled and submitted', async function () {
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
