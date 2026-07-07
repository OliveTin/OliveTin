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

async function openChecklistArgumentForm(actionTitle = 'Test checklist argument') {
  await getRootAndWait()
  const btn = await getActionButton(webdriver, actionTitle)
  await btn.click()

  await waitForArgumentFormPage()
  await waitForArgumentFormReady()
}

async function submitChecklistForm() {
  const submitButton = await webdriver.findElement(By.css('button[name="start"]'))
  await submitButton.click()
}

async function pollTerminal(matcher, timeoutMs = DEFAULT_UI_WAIT_MS) {
  await webdriver.wait(
    new Condition('wait for terminal output', async () => {
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

        return matcher(output.trim())
      } catch (e) {
        return false
      }
    }),
    timeoutMs
  )
}

async function waitForTerminalOutput(expectedValue, label = 'Selected segments') {
  await pollTerminal(
    (output) => output.includes(`${label}: ${expectedValue}`),
    DEFAULT_UI_WAIT_MS
  )
}

async function waitForTerminalOutputPattern(pattern) {
  await pollTerminal(
    (output) => pattern.test(output),
    DEFAULT_UI_WAIT_MS
  )
}

async function waitForChecklistValue(expectedValue) {
  await webdriver.wait(
    new Condition('wait for checklist hidden value', async () => {
      const valueInput = await webdriver.findElement(By.css('.choice-checklist > input'))
      return (await valueInput.getAttribute('value')) === expectedValue
    }),
    DEFAULT_UI_WAIT_MS
  )
}

async function getCheckboxByValueIndex(index) {
  const checkboxes = await webdriver.findElements(
    By.css('.choice-checklist-item input[type="checkbox"]')
  )
  return checkboxes[index]
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
    await waitForChecklistValue('')

    const valueInput = await webdriver.findElement(By.css('.choice-checklist > input'))
    expect(await valueInput.getAttribute('value')).to.equal('')

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
    await waitForTerminalOutput('["kitchen","bedroom","hallway"]')
  })

  it('Checklist toggles individual choices before submit', async function () {
    await openChecklistArgumentForm()

    const hallway = await getCheckboxByValueIndex(2)
    await hallway.click()

    await submitChecklistForm()
    await waitForLogsPage()
    await waitForExecutionComplete()
    await waitForTerminalOutput('["kitchen","bedroom","hallway"]')
  })

  it('Checklist entity argument renders choices from entities', async function () {
    await openChecklistArgumentForm('Test checklist entity argument')

    const checkboxes = await webdriver.findElements(
      By.css('.choice-checklist-item input[type="checkbox"]')
    )
    expect(checkboxes).to.have.length(2)

    const labels = await webdriver.findElements(By.css('.choice-checklist-item span'))
    expect(await labels[0].getText()).to.equal('attic')
    expect(await labels[1].getText()).to.equal('basement')

    await checkboxes[0].click()

    await submitChecklistForm()
    await waitForLogsPage()
    await waitForExecutionComplete()
    await waitForTerminalOutput('["attic"]', 'Selected rooms')
  })
})
