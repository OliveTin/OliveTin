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

async function getJustificationInput () {
  return await webdriver.findElement(By.id('justification'))
}

async function getJustificationValue () {
  const input = await getJustificationInput()
  return await input.getAttribute('value')
}

async function waitForJustificationValue (expected) {
  await webdriver.wait(
    new Condition(`wait for justification value "${expected}"`, async () => {
      const value = await getJustificationValue()
      return value === expected
    }),
    DEFAULT_UI_WAIT_MS
  )
}

async function fillJustification (text) {
  const input = await getJustificationInput()
  await input.clear()
  await input.sendKeys(text)
  await webdriver.sleep(100)
}

async function fillArgumentField (argumentName, value) {
  const input = await webdriver.findElement(By.id(argumentFieldId(argumentName)))
  await webdriver.executeScript(
    `const input = arguments[0];
     input.value = arguments[1];
     input.dispatchEvent(new Event('input', { bubbles: true }));
     input.dispatchEvent(new Event('change', { bubbles: true }));`,
    input,
    value
  )
  await webdriver.sleep(100)
}

async function submitForm () {
  const submitButton = await getStartButton()
  await submitButton.click()
}

async function assertStillOnArgumentForm () {
  const url = await webdriver.getCurrentUrl()
  expect(url).to.include('/argumentForm')
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
      } catch {
        return false
      }
    }),
    DEFAULT_UI_WAIT_MS
  )
}

async function waitForExecutionJustification (expectedText) {
  await webdriver.wait(
    new Condition(`wait for execution justification "${expectedText}"`, async () => {
      const elements = await webdriver.findElements(By.css('.log-justification'))
      if (elements.length === 0) {
        return false
      }

      const text = await elements[0].getText()
      return text.includes(expectedText)
    }),
    DEFAULT_UI_WAIT_MS
  )
}

async function submitWithJustification (text) {
  await fillJustification(text)
  await submitForm()
  await waitForLogsPage()
  await waitForExecutionComplete()
}

describe('config: justification', function () {
  this.timeout(10000)

  before(async function () {
    await runner.start('justification')
  })

  after(async () => {
    await runner.stop()
  })

  afterEach(function () {
    takeScreenshotOnFailure(this.currentTest, webdriver)
  })

  it('Actions without justification skip the argument form', async function () {
    await getRootAndWait()
    const btn = await getActionButton(webdriver, 'Test action without justification')
    await btn.click()

    await webdriver.wait(
      new Condition('navigates away from argument form', async () => {
        const url = await webdriver.getCurrentUrl()
        return url.includes('/logs/') && !url.includes('/argumentForm')
      }),
      DEFAULT_UI_WAIT_MS
    )

    await waitForExecutionComplete()
    await waitForTerminalOutput('No justification needed')
  })

  it('Manual justification opens a required empty field with no arguments', async function () {
    await openArgumentForm('Test manual justification')

    const justificationInput = await getJustificationInput()
    expect(await justificationInput.getTagName()).to.equal('input')
    expect(await justificationInput.getAttribute('type')).to.equal('text')
    expect(await justificationInput.getAttribute('required')).to.equal('true')
    expect((await getJustificationValue()).trim()).to.equal('')

    const label = await webdriver.findElement(By.css('label[for="justification"]'))
    expect(await label.getText()).to.contain('Justification')

    const argumentFields = await webdriver.findElements(By.css('[id^="arg-field-"]'))
    expect(argumentFields).to.have.length(0)
  })

  it('Manual justification blocks submit until filled', async function () {
    await openArgumentForm('Test manual justification')

    await submitForm()
    await assertStillOnArgumentForm()

    const justificationInput = await getJustificationInput()
    const validationMessage = await webdriver.executeScript(
      'return arguments[0].validationMessage',
      justificationInput
    )
    expect(validationMessage).to.not.equal('')
  })

  it('Manual justification is stored and shown on the execution page', async function () {
    await openArgumentForm('Test manual justification')

    await submitWithJustification('Approved maintenance window')
    await waitForTerminalOutput('Manual justification action ran')
    await waitForExecutionJustification('Approved maintenance window')
  })

  it('Templated justification prefills from argument defaults', async function () {
    await openArgumentForm('Test templated justification')

    const targetInput = await webdriver.findElement(By.id(argumentFieldId('target')))
    expect(await targetInput.getAttribute('value')).to.equal('dbserver')
    await waitForJustificationValue('dbserver')

    await submitForm()
    await waitForLogsPage()
    await waitForExecutionComplete()
    await waitForTerminalOutput('Target: dbserver')
    await waitForExecutionJustification('dbserver')
  })

  it('Templated justification updates when argument values change', async function () {
    await openArgumentForm('Test templated justification')

    await waitForJustificationValue('dbserver')

    await fillArgumentField('target', 'appserver')
    await waitForJustificationValue('appserver')
  })

  it('Manual edits to templated justification are preserved when arguments change', async function () {
    await openArgumentForm('Test templated justification')

    await fillArgumentField('target', 'hosta')
    await waitForJustificationValue('hosta')

    await fillJustification('Manual override reason')
    expect(await getJustificationValue()).to.equal('Manual override reason')

    await fillArgumentField('target', 'hostb')
    await webdriver.sleep(200)

    expect(await getJustificationValue()).to.equal('Manual override reason')
  })

  it('Justification is required alongside regular arguments', async function () {
    await openArgumentForm('Test justification with arguments')

    const targetInput = await webdriver.findElement(By.id(argumentFieldId('target')))
    expect(await targetInput.getAttribute('value')).to.equal('testhost')

    await submitForm()
    await assertStillOnArgumentForm()

    await submitWithJustification('Deploy to testhost')
    await waitForTerminalOutput('Target: testhost')
    await waitForExecutionJustification('Deploy to testhost')
  })
})
