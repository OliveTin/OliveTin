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
  waitForArgumentFormReady,
  waitForLogsPage,
  waitForExecutionComplete,
} from '../../lib/elements.js'

async function openChecklistArgumentForm() {
  await getRootAndWait()
  const btn = await getActionButton(webdriver, 'Test checklist argument')
  await btn.click()

  await waitForArgumentFormPage()
  await waitForArgumentFormReady()
}

async function submitChecklistForm() {
  const submitButton = await webdriver.findElement(By.css('button[name="start"]'))
  await submitButton.click()
}

async function waitForTerminalOutput(expectedValue) {
  await webdriver.wait(
    new Condition(`wait for checklist value ${expectedValue} in output`, async () => {
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

        return output.trim().includes(`Selected segments: ${expectedValue}`)
      } catch (e) {
        return false
      }
    }),
    DEFAULT_UI_WAIT_MS
  )
}

async function waitForTerminalOutputPattern(pattern) {
  await webdriver.wait(
    new Condition(`wait for terminal output matching ${pattern}`, async () => {
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

        return pattern.test(output.trim())
      } catch (e) {
        return false
      }
    }),
    10000
  )
}

async function getCheckboxByValueIndex(index) {
  return await webdriver.findElement(By.id(`segments-${index}`))
}

describe('config: checklist', function () {
  this.timeout(10000)

  before(async function () {
    await runner.start('checklist')
  })

  after(async () => {
    await runner.stop()
  })

  afterEach(function () {
    takeScreenshotOnFailure(this.currentTest, webdriver)
  })

  it('Checklist argument renders multiple checkbox inputs', async function () {
    await openChecklistArgumentForm()

    const kitchen = await getCheckboxByValueIndex(0)
    const bedroom = await getCheckboxByValueIndex(1)
    const hallway = await getCheckboxByValueIndex(2)

    expect(await kitchen.getAttribute('type')).to.equal('checkbox')
    expect(await bedroom.getAttribute('type')).to.equal('checkbox')
    expect(await hallway.getAttribute('type')).to.equal('checkbox')
    expect(await kitchen.isSelected()).to.be.true
    expect(await bedroom.isSelected()).to.be.true
    expect(await hallway.isSelected()).to.be.false
  })

  it('Checklist select none submits an empty value', async function () {
    await openChecklistArgumentForm()

    const selectNone = await webdriver.findElement(By.xpath("//button[normalize-space()='Select none']"))
    await selectNone.click()
    await webdriver.sleep(300)

    const hidden = await webdriver.findElement(By.id('segments-value'))
    expect(await hidden.getAttribute('value')).to.equal('')

    await submitChecklistForm()
    await waitForLogsPage()
    await waitForExecutionComplete()
    await waitForTerminalOutputPattern(/Selected segments:\s*(\r?\n|$)/)
  })

  it('Checklist select all submits every choice value', async function () {
    await openChecklistArgumentForm()

    const selectNone = await webdriver.findElement(By.xpath("//button[normalize-space()='Select none']"))
    await selectNone.click()

    const selectAll = await webdriver.findElement(By.xpath("//button[normalize-space()='Select all']"))
    await selectAll.click()

    await submitChecklistForm()
    await waitForLogsPage()
    await waitForExecutionComplete()
    await waitForTerminalOutput('kitchen,bedroom,hallway')
  })

  it('Checklist toggles individual choices before submit', async function () {
    await openChecklistArgumentForm()

    const hallway = await getCheckboxByValueIndex(2)
    await hallway.click()

    await submitChecklistForm()
    await waitForLogsPage()
    await waitForExecutionComplete()
    await waitForTerminalOutput('kitchen,bedroom,hallway')
  })
})
