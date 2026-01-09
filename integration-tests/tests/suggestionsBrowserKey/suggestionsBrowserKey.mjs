import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'
import { By, Condition } from 'selenium-webdriver'
import {
  getRootAndWait,
  getActionButton,
  takeScreenshotOnFailure,
  getTerminalBuffer,
} from '../../lib/elements.js'

async function openArgumentForm() {
  await getRootAndWait()
  const btn = await getActionButton(webdriver, 'Test suggestionsBrowserKey')
  await btn.click()

  await webdriver.wait(
    new Condition('wait for argument form page', async () => {
      const url = await webdriver.getCurrentUrl()
      return url.includes('/actionBinding/') && url.includes('/argumentForm')
    }),
    5000
  )
}

async function getTestInput() {
  return await webdriver.findElement(By.id('testInput'))
}

async function getTestInput2() {
  return await webdriver.findElement(By.id('testInput2'))
}

async function getDatalistOptions(inputName = 'testInput') {
  return await webdriver.findElements(By.css(`datalist#${inputName}-choices option`))
}

async function submitForm() {
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

  await webdriver.sleep(500)
}

async function getLocalStorageItem(key) {
  return await webdriver.executeScript(`return localStorage.getItem('${key}')`)
}

async function clearLocalStorage() {
  await webdriver.executeScript('return localStorage.clear()')
}

describe('config: suggestionsBrowserKey', function () {
  before(async function () {
    await runner.start('suggestionsBrowserKey')
  })

  after(async () => {
    await runner.stop()
  })

  afterEach(function () {
    takeScreenshotOnFailure(this.currentTest, webdriver)
  })

  it('Input fields with suggestionsBrowserKey are rendered', async function () {
    await openArgumentForm()

    const input1 = await getTestInput()
    expect(await input1.getTagName()).to.equal('input')
    expect(await input1.getAttribute('type')).to.equal('text')

    const label1 = await webdriver.findElement(By.css('label[for="testInput"]'))
    expect(await label1.getText()).to.contain('Test Input')

    const input2 = await getTestInput2()
    expect(await input2.getTagName()).to.equal('input')
    expect(await input2.getAttribute('type')).to.equal('text')

    const label2 = await webdriver.findElement(By.css('label[for="testInput2"]'))
    expect(await label2.getText()).to.contain('Test Input 2')
  })

  it('Submitting form saves value to localStorage', async function () {
    this.timeout(15000)
    
    // Clear localStorage first
    await clearLocalStorage()
    
    await openArgumentForm()

    const input = await getTestInput()
    const testValue = 'test-value-123'
    await input.clear()
    await input.sendKeys(testValue)

    await submitForm()
    await waitForLogsPage()
    await waitForExecutionComplete()

    // Verify value was saved to localStorage
    const stored = await getLocalStorageItem('olivetin-suggestions-test-suggestions-key')
    expect(stored).to.not.be.null
    
    const suggestions = JSON.parse(stored)
    expect(suggestions).to.be.an('array')
    expect(suggestions).to.include(testValue)
  })

  it('Previously saved values appear in datalist', async function () {
    this.timeout(15000)
    
    // First, save a value to localStorage
    const testValue = 'saved-suggestion-456'
    await webdriver.executeScript(`
      const key = 'olivetin-suggestions-test-suggestions-key';
      localStorage.setItem(key, JSON.stringify(['${testValue}']));
    `)

    // Open the form
    await openArgumentForm()

    // Check that datalist exists and contains the saved value
    const datalist = await webdriver.findElement(By.id('testInput-choices'))
    expect(datalist).to.not.be.null

    const options = await getDatalistOptions()
    expect(options.length).to.be.greaterThan(0)

    // Check if the saved value appears in the datalist
    let foundValue = false
    for (const option of options) {
      const value = await option.getAttribute('value')
      if (value === testValue) {
        foundValue = true
        break
      }
    }
    expect(foundValue).to.be.true
  })

  it('Multiple submissions accumulate suggestions', async function () {
    this.timeout(20000)
    
    // Clear localStorage first
    await clearLocalStorage()

    // Submit first value
    await openArgumentForm()
    const input1 = await getTestInput()
    await input1.clear()
    await input1.sendKeys('first-value')
    await submitForm()
    await waitForLogsPage()
    await waitForExecutionComplete()

    // Submit second value
    await openArgumentForm()
    const input2 = await getTestInput()
    await input2.clear()
    await input2.sendKeys('second-value')
    await submitForm()
    await waitForLogsPage()
    await waitForExecutionComplete()

    // Verify both values are in localStorage
    const stored = await getLocalStorageItem('olivetin-suggestions-test-suggestions-key')
    expect(stored).to.not.be.null
    
    const suggestions = JSON.parse(stored)
    expect(suggestions).to.be.an('array')
    expect(suggestions).to.include('first-value')
    expect(suggestions).to.include('second-value')
    expect(suggestions[0]).to.equal('second-value') // Most recent should be first
  })

  it('Empty values are not saved to localStorage', async function () {
    this.timeout(15000)
    
    // Clear localStorage first
    await clearLocalStorage()

    await openArgumentForm()

    const input = await getTestInput()
    // Leave input empty (or clear it if it has a default)
    await input.clear()

    await submitForm()
    await waitForLogsPage()
    await waitForExecutionComplete()

    // Verify empty value was not saved - localStorage should be null or not contain the key
    const stored = await getLocalStorageItem('olivetin-suggestions-test-suggestions-key')
    // Should be null since empty values are not saved
    expect(stored).to.be.null
  })

  it('Suggestions are shared across inputs with the same suggestionsBrowserKey', async function () {
    this.timeout(20000)
    
    // Clear localStorage first
    await clearLocalStorage()

    // Submit a value using the first input
    await openArgumentForm()
    const input1 = await getTestInput()
    await input1.clear()
    await input1.sendKeys('shared-value-from-input1')
    await submitForm()
    await waitForLogsPage()
    await waitForExecutionComplete()

    // Open the form again and verify the value appears in both datalists
    await openArgumentForm()
    
    // Check first input's datalist
    const datalist1 = await webdriver.findElement(By.id('testInput-choices'))
    expect(datalist1).to.not.be.null
    const options1 = await getDatalistOptions('testInput')
    let foundInInput1 = false
    for (const option of options1) {
      const value = await option.getAttribute('value')
      if (value === 'shared-value-from-input1') {
        foundInInput1 = true
        break
      }
    }
    expect(foundInInput1).to.be.true

    // Check second input's datalist
    const datalist2 = await webdriver.findElement(By.id('testInput2-choices'))
    expect(datalist2).to.not.be.null
    const options2 = await getDatalistOptions('testInput2')
    let foundInInput2 = false
    for (const option of options2) {
      const value = await option.getAttribute('value')
      if (value === 'shared-value-from-input1') {
        foundInInput2 = true
        break
      }
    }
    expect(foundInInput2).to.be.true

    // Now submit a value using the second input
    const input2 = await getTestInput2()
    await input2.clear()
    await input2.sendKeys('shared-value-from-input2')
    await submitForm()
    await waitForLogsPage()
    await waitForExecutionComplete()

    // Verify both values appear in both datalists
    await openArgumentForm()
    
    // Check that both values are in the first input's datalist
    const options1After = await getDatalistOptions('testInput')
    let foundValue1 = false
    let foundValue2 = false
    for (const option of options1After) {
      const value = await option.getAttribute('value')
      if (value === 'shared-value-from-input1') {
        foundValue1 = true
      }
      if (value === 'shared-value-from-input2') {
        foundValue2 = true
      }
    }
    expect(foundValue1).to.be.true
    expect(foundValue2).to.be.true

    // Check that both values are in the second input's datalist
    const options2After = await getDatalistOptions('testInput2')
    foundValue1 = false
    foundValue2 = false
    for (const option of options2After) {
      const value = await option.getAttribute('value')
      if (value === 'shared-value-from-input1') {
        foundValue1 = true
      }
      if (value === 'shared-value-from-input2') {
        foundValue2 = true
      }
    }
    expect(foundValue1).to.be.true
    expect(foundValue2).to.be.true
  })
})
