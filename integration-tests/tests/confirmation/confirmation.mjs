import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'
import { By, Condition } from 'selenium-webdriver'
import {
  DEFAULT_UI_WAIT_MS,
  argumentFieldId,
  getRootAndWait,
  getActionButton,
  takeScreenshotOnFailure,
  waitForArgumentFormPage,
  waitForArgumentFormReady,
  waitForLogsPage,
  waitForExecutionComplete,
  getTerminalBuffer,
} from '../../lib/elements.js'

async function openArgumentForm (actionTitle) {
  await getRootAndWait()
  const btn = await getActionButton(webdriver, actionTitle)
  await btn.click()
  await waitForArgumentFormPage()
  await waitForArgumentFormReady()
}

async function getStartButton () {
  return await webdriver.findElement(By.css('button[name="start"]'))
}

async function waitForStartButtonEnabled () {
  await webdriver.wait(
    new Condition('wait for Start button to be enabled', async () => {
      const submitButton = await getStartButton()
      return await submitButton.isEnabled()
    }),
    DEFAULT_UI_WAIT_MS
  )
}

async function waitForStartButtonDisabled () {
  await webdriver.wait(
    new Condition('wait for Start button to be disabled', async () => {
      const submitButton = await getStartButton()
      return !(await submitButton.isEnabled())
    }),
    DEFAULT_UI_WAIT_MS
  )
}

async function waitForTerminalOutput (expectedSubstring) {
  await webdriver.wait(
    new Condition(`wait for terminal output containing ${expectedSubstring}`, async () => {
      try {
        const terminalReady = await webdriver.executeScript(`
          return !!(window.terminal && window.terminal.getBufferAsString);
        `)
        if (!terminalReady) {
          return false
        }

        const output = await getTerminalBuffer()
        return output && output.includes(expectedSubstring)
      } catch (e) {
        return false
      }
    }),
    DEFAULT_UI_WAIT_MS
  )
}

async function confirmAndSubmit (checkbox) {
  await checkbox.click()
  await waitForStartButtonEnabled()
  const submitButton = await getStartButton()
  await submitButton.click()
  await waitForLogsPage()
  await waitForExecutionComplete()
}

describe('config: confirmation', function () {
  this.timeout(10000)

  before(async function () {
    await runner.start('confirmation')
  })

  after(async () => {
    await runner.stop()
  })

  afterEach(function () {
    takeScreenshotOnFailure(this.currentTest, webdriver)
  })

  it('Unnamed confirmation renders a checkbox and disables Start until ticked', async function () {
    await openArgumentForm('Test unnamed confirmation argument')

    const checkbox = await webdriver.findElement(By.css('#argument-popup input[type="checkbox"]'))
    expect(await checkbox.getTagName()).to.equal('input')
    expect(await checkbox.getAttribute('type')).to.equal('checkbox')
    expect(await checkbox.isSelected()).to.be.false

    const label = await webdriver.findElement(By.css(`label[for="${argumentFieldId('')}"]`))
    expect(await label.getText()).to.contain('Are you sure?!')

    const submitButton = await getStartButton()
    expect(await submitButton.isEnabled()).to.be.false
  })

  it('Unnamed confirmation runs the action without substituting a value', async function () {
    await openArgumentForm('Test unnamed confirmation argument')

    const checkbox = await webdriver.findElement(By.css('#argument-popup input[type="checkbox"]'))
    await confirmAndSubmit(checkbox)
    await waitForTerminalOutput('Confirmed action ran')
  })

  it('Named confirmation disables Start until ticked and submits 1 when checked', async function () {
    await openArgumentForm('Test named confirmation argument')

    const checkbox = await webdriver.findElement(By.id(argumentFieldId('agree')))
    expect(await checkbox.isSelected()).to.be.false

    const label = await webdriver.findElement(By.css(`label[for="${argumentFieldId('agree')}"]`))
    expect(await label.getText()).to.contain('I understand the consequences')

    const submitButton = await getStartButton()
    expect(await submitButton.isEnabled()).to.be.false

    await confirmAndSubmit(checkbox)
    await waitForTerminalOutput('Confirmation value: 1')
  })

  it('Named confirmation works with shell actions', async function () {
    await openArgumentForm('Test shell with named confirmation')

    const checkbox = await webdriver.findElement(By.id(argumentFieldId('agree')))
    const submitButton = await getStartButton()
    expect(await submitButton.isEnabled()).to.be.false

    await confirmAndSubmit(checkbox)
    await waitForTerminalOutput('Shell confirmation value: 1')
  })

  it('Confirmation keeps Start disabled while unchecked (unlike checkbox arguments)', async function () {
    await openArgumentForm('Test named confirmation argument')

    const checkbox = await webdriver.findElement(By.id(argumentFieldId('agree')))
    expect(await checkbox.isSelected()).to.be.false
    await waitForStartButtonDisabled()

    await checkbox.click()
    await webdriver.sleep(100)
    expect(await checkbox.isSelected()).to.be.true
    await waitForStartButtonEnabled()

    await checkbox.click()
    await webdriver.sleep(100)
    expect(await checkbox.isSelected()).to.be.false
    await waitForStartButtonDisabled()
  })
})
