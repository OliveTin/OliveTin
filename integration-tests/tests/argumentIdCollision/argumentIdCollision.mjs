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

async function waitForStartButtonEnabled () {
  await webdriver.wait(
    new Condition('wait for Start button to be enabled', async () => {
      const submitButton = await webdriver.findElement(By.css('button[name="start"]'))
      return await submitButton.isEnabled()
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

describe('config: argumentIdCollision', function () {
  this.timeout(10000)

  before(async function () {
    await runner.start('argumentIdCollision')
  })

  after(async () => {
    await runner.stop()
  })

  afterEach(function () {
    takeScreenshotOnFailure(this.currentTest, webdriver)
  })

  it('Argument named content validates and submits (#1071)', async function () {
    await openArgumentForm('Test content argument collision')

    const contentWrapper = await webdriver.findElement(By.id('content'))
    expect(await contentWrapper.getTagName()).to.equal('div')

    const textarea = await webdriver.findElement(By.id(argumentFieldId('content')))
    expect(await textarea.getTagName()).to.equal('textarea')
    expect(await textarea.getAttribute('id')).to.not.equal('content')

    const label = await webdriver.findElement(By.css(`label[for="${argumentFieldId('content')}"]`))
    expect(await label.getText()).to.contain('Cmd input')

    await textarea.sendKeys('hello from collision test')

    const submitButton = await webdriver.findElement(By.css('button[name="start"]'))
    await waitForStartButtonEnabled()
    await submitButton.click()
    await waitForLogsPage()
    await waitForExecutionComplete()
    await waitForTerminalOutput('Cmd input: hello from collision test')
  })

  it('Argument named layout validates and submits (#1071)', async function () {
    await openArgumentForm('Test layout argument collision')

    const layoutWrapper = await webdriver.findElement(By.id('layout'))
    expect(await layoutWrapper.getTagName()).to.equal('div')

    const input = await webdriver.findElement(By.id(argumentFieldId('layout')))
    expect(await input.getAttribute('type')).to.equal('text')
    expect(await input.getAttribute('id')).to.not.equal('layout')

    await input.sendKeys('testlayoutvalue')

    const submitButton = await webdriver.findElement(By.css('button[name="start"]'))
    await waitForStartButtonEnabled()
    await submitButton.click()
    await waitForLogsPage()
    await waitForExecutionComplete()
    await waitForTerminalOutput('Layout value: testlayoutvalue')
  })

  it('Namespaced argument ids do not match app-shell ids', async function () {
    await openArgumentForm('Test content argument collision')

    const namespacedIds = await webdriver.executeScript(`
      const requiredShellIds = ['content', 'layout'];
      const optionalShellIds = ['banner', 'app', 'mainnav', 'big-error'];
      const fieldId = arguments[0];
      const fieldElement = document.getElementById(fieldId);

      return {
        fieldTag: fieldElement?.tagName?.toLowerCase() ?? null,
        required: requiredShellIds.map((shellId) => {
          const shellElement = document.getElementById(shellId);
          return {
            shellId,
            shellTag: shellElement?.tagName?.toLowerCase() ?? null,
            sameElement: shellElement != null && shellElement === fieldElement
          };
        }),
        optional: optionalShellIds.map((shellId) => {
          const shellElement = document.getElementById(shellId);
          return {
            shellId,
            sameElement: shellElement != null && shellElement === fieldElement
          };
        })
      };
    `, argumentFieldId('content'))

    expect(namespacedIds.fieldTag, `argument field ${argumentFieldId('content')} should exist`).to.equal('textarea')

    for (const entry of namespacedIds.required) {
      expect(entry.shellTag, `app-shell #${entry.shellId} should exist`).to.not.equal(null)
      expect(entry.sameElement, `arg field must not resolve to #${entry.shellId}`).to.be.false
    }

    for (const entry of namespacedIds.optional) {
      expect(entry.sameElement, `arg field must not resolve to #${entry.shellId}`).to.be.false
    }
  })
})
