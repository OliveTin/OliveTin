import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'
import { By, Condition } from 'selenium-webdriver'
import {
  DEFAULT_UI_WAIT_MS,
  argumentFieldChoicesId,
  argumentFieldId,
  getRootAndWait,
  getActionButton,
  takeScreenshotOnFailure,
  waitForDashboardLoaded,
  waitForLogsPage,
  waitForArgumentFormPage,
  waitForArgumentFormReady,
  waitForExecutionComplete,
} from '../../lib/elements.js'

async function clickBackFromLogsPage() {
  const goBackButtons = await webdriver.findElements(By.css('button[title="Go back"]'))
  if (goBackButtons.length > 0) {
    await goBackButtons[0].click()
    return 'history'
  }

  const dashboardBackButtons = await webdriver.findElements(By.css('button[title^="Back to "]'))
  if (dashboardBackButtons.length > 0) {
    await dashboardBackButtons[0].click()
    return 'dashboard'
  }

  throw new Error('No back button found on execution logs page')
}

async function ensureOnDashboard() {
  let url = await webdriver.getCurrentUrl()

  if (url.includes('/logs/')) {
    const backType = await clickBackFromLogsPage()
    if (backType === 'history') {
      await webdriver.wait(
        new Condition('wait for argument form after logs back', async () => {
          const currentUrl = await webdriver.getCurrentUrl()
          return currentUrl.includes('/argumentForm')
        }),
        DEFAULT_UI_WAIT_MS
      )
      url = await webdriver.getCurrentUrl()
    } else {
      await waitForDashboardLoaded()
      url = await webdriver.getCurrentUrl()
    }
  }

  if (url.includes('/argumentForm')) {
    const cancelButton = await webdriver.findElement(By.css('button[name="cancel"]'))
    await cancelButton.click()
    await waitForDashboardLoaded()
  }

  const actionButtons = await webdriver.findElements(By.css('[title="Test suggestionsBrowserKey"]'))
  if (actionButtons.length === 1) {
    return
  }

  await getRootAndWait()
}

async function openArgumentForm() {
  await ensureOnDashboard()
  const btn = await getActionButton(webdriver, 'Test suggestionsBrowserKey')
  await btn.click()

  await waitForArgumentFormPage()
  await waitForArgumentFormReady()
}

async function getTestInput() {
  return await webdriver.findElement(By.id(argumentFieldId('testInput')))
}

async function getTestInput2() {
  return await webdriver.findElement(By.id(argumentFieldId('testInput2')))
}

async function getDatalistOptions(inputName = 'testInput') {
  return await webdriver.findElements(By.css(`datalist#${argumentFieldChoicesId(inputName)} option`))
}

async function submitForm() {
  const submitButton = await webdriver.findElement(By.css('button[name="start"]'))
  await submitButton.click()
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
    await getRootAndWait()
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

    const label1 = await webdriver.findElement(By.css(`label[for="${argumentFieldId('testInput')}"]`))
    expect(await label1.getText()).to.contain('Test Input')

    const input2 = await getTestInput2()
    expect(await input2.getTagName()).to.equal('input')
    expect(await input2.getAttribute('type')).to.equal('text')

    const label2 = await webdriver.findElement(By.css(`label[for="${argumentFieldId('testInput2')}"]`))
    expect(await label2.getText()).to.contain('Test Input 2')
  })

  it('Submitting form saves value to localStorage', async function () {
    await clearLocalStorage()
    await openArgumentForm()

    const input = await getTestInput()
    // Use default argument type "ascii" (alphanumeric only) so tests pass when
    // config does not set a looser type (e.g. CI merge base without type lines).
    const testValue = 'testvalue123'
    await input.clear()
    await input.sendKeys(testValue)

    await submitForm()
    await waitForLogsPage()
    await waitForExecutionComplete()

    const stored = await getLocalStorageItem('olivetin-suggestions-test-suggestions-key')
    expect(stored).to.not.be.null

    const suggestions = JSON.parse(stored)
    expect(suggestions).to.be.an('array')
    expect(suggestions).to.include(testValue)
  })

  it('Previously saved values appear in datalist', async function () {
    const testValue = 'savedsuggestion456'
    await webdriver.executeScript(`
      const key = 'olivetin-suggestions-test-suggestions-key';
      localStorage.setItem(key, JSON.stringify(['${testValue}']));
    `)

    await openArgumentForm()

    const datalist = await webdriver.findElement(By.id(argumentFieldChoicesId('testInput')))
    expect(datalist).to.not.be.null

    const options = await getDatalistOptions()
    expect(options.length).to.be.greaterThan(0)

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
    await clearLocalStorage()

    await openArgumentForm()
    const input1 = await getTestInput()
    await input1.clear()
    await input1.sendKeys('firstvalue')
    await submitForm()
    await waitForLogsPage()
    await waitForExecutionComplete()

    await openArgumentForm()
    const input2 = await getTestInput()
    await input2.clear()
    await input2.sendKeys('secondvalue')
    await submitForm()
    await waitForLogsPage()
    await waitForExecutionComplete()

    const stored = await getLocalStorageItem('olivetin-suggestions-test-suggestions-key')
    expect(stored).to.not.be.null

    const suggestions = JSON.parse(stored)
    expect(suggestions).to.be.an('array')
    expect(suggestions).to.include('firstvalue')
    expect(suggestions).to.include('secondvalue')
    expect(suggestions[0]).to.equal('secondvalue')
  })

  it('Empty values are not saved to localStorage', async function () {
    await clearLocalStorage()
    await openArgumentForm()

    const input = await getTestInput()
    await input.clear()

    await submitForm()
    await waitForLogsPage()
    await waitForExecutionComplete()

    const stored = await getLocalStorageItem('olivetin-suggestions-test-suggestions-key')
    if (stored !== null) {
      const suggestions = JSON.parse(stored)
      expect(suggestions).to.be.an('array')
      expect(suggestions).to.have.length(0)
    }
  })

  it('Suggestions are shared across inputs with the same suggestionsBrowserKey', async function () {
    this.timeout(12000)

    await clearLocalStorage()

    await openArgumentForm()
    const input1 = await getTestInput()
    await input1.clear()
    await input1.sendKeys('sharedfrominput1')
    await submitForm()
    await waitForLogsPage()
    await waitForExecutionComplete()

    await openArgumentForm()

    const datalist1 = await webdriver.findElement(By.id(argumentFieldChoicesId('testInput')))
    expect(datalist1).to.not.be.null
    const options1 = await getDatalistOptions('testInput')
    let foundInInput1 = false
    for (const option of options1) {
      const value = await option.getAttribute('value')
      if (value === 'sharedfrominput1') {
        foundInInput1 = true
        break
      }
    }
    expect(foundInInput1).to.be.true

    const datalist2 = await webdriver.findElement(By.id(argumentFieldChoicesId('testInput2')))
    expect(datalist2).to.not.be.null
    const options2 = await getDatalistOptions('testInput2')
    let foundInInput2 = false
    for (const option of options2) {
      const value = await option.getAttribute('value')
      if (value === 'sharedfrominput1') {
        foundInInput2 = true
        break
      }
    }
    expect(foundInInput2).to.be.true

    const input2 = await getTestInput2()
    await input2.clear()
    await input2.sendKeys('sharedfrominput2')
    await submitForm()
    await waitForLogsPage()
    await waitForExecutionComplete()

    await openArgumentForm()

    const options1After = await getDatalistOptions('testInput')
    let foundValue1 = false
    let foundValue2 = false
    for (const option of options1After) {
      const value = await option.getAttribute('value')
      if (value === 'sharedfrominput1') {
        foundValue1 = true
      }
      if (value === 'sharedfrominput2') {
        foundValue2 = true
      }
    }
    expect(foundValue1).to.be.true
    expect(foundValue2).to.be.true

    const options2After = await getDatalistOptions('testInput2')
    foundValue1 = false
    foundValue2 = false
    for (const option of options2After) {
      const value = await option.getAttribute('value')
      if (value === 'sharedfrominput1') {
        foundValue1 = true
      }
      if (value === 'sharedfrominput2') {
        foundValue2 = true
      }
    }
    expect(foundValue1).to.be.true
    expect(foundValue2).to.be.true
  })
})
