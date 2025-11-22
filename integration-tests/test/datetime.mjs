import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'
import { By, Condition } from 'selenium-webdriver'
import {
  getRootAndWait,
  getActionButton,
  takeScreenshotOnFailure,
} from '../lib/elements.js'

describe('config: datetime', function () {
  before(async function () {
    await runner.start('datetime')
  })

  after(async () => {
    await runner.stop()
  })

  afterEach(function () {
    takeScreenshotOnFailure(this.currentTest, webdriver)
  })

  it('Datetime argument uses datetime-local input type', async function () {
    await getRootAndWait()

    const btn = await getActionButton(webdriver, 'Test datetime argument')

    await btn.click()

    // Wait for navigation to argument form page
    await webdriver.wait(
      new Condition('wait for argument form page', async () => {
        const url = await webdriver.getCurrentUrl()
        return url.includes('/actionBinding/') && url.includes('/argumentForm')
      }),
      8000
    )

    // Find the datetime input field
    const datetimeInput = await webdriver.findElement(By.id('datetime'))

    // Verify it's a datetime-local input type
    const inputType = await datetimeInput.getAttribute('type')
    expect(inputType).to.equal('datetime-local', 'Input type should be datetime-local')

    // Verify it has the step attribute set to 1 (for seconds precision)
    const step = await datetimeInput.getAttribute('step')
    expect(step).to.equal('1', 'Step attribute should be 1')

    // Verify the label is present
    const label = await webdriver.findElement(By.css('label[for="datetime"]'))
    expect(await label.getText()).to.contain('Select a date and time')
  })

  it('Datetime argument can be filled and submitted', async function () {
    await getRootAndWait()

    const btn = await getActionButton(webdriver, 'Test datetime argument')

    await btn.click()

    // Wait for navigation to argument form page
    await webdriver.wait(
      new Condition('wait for argument form page', async () => {
        const url = await webdriver.getCurrentUrl()
        return url.includes('/actionBinding/') && url.includes('/argumentForm')
      }),
      8000
    )

    // Find the datetime input field
    const datetimeInput = await webdriver.findElement(By.id('datetime'))

    // Set a datetime value (format: YYYY-MM-DDTHH:mm)
    // datetime-local returns values without seconds, backend will add :00
    const testDateTime = '2023-12-25T15:30'
    
    // Use JavaScript to set the value directly (more reliable for datetime-local inputs)
    await webdriver.executeScript(
      'arguments[0].value = arguments[1]',
      datetimeInput,
      testDateTime
    )

    // Trigger input event to ensure Vue reactivity
    await webdriver.executeScript(
      'arguments[0].dispatchEvent(new Event("input", { bubbles: true }))',
      datetimeInput
    )

    // Small wait for Vue to process the change
    await webdriver.sleep(100)

    // Verify the value was set
    const value = await datetimeInput.getAttribute('value')
    expect(value).to.equal(testDateTime)

    // Find and click the submit button
    const submitButton = await webdriver.findElement(
      By.css('button[name="start"]')
    )
    await submitButton.click()

    // Wait for navigation to logs page
    await webdriver.wait(
      new Condition('wait for logs page', async () => {
        const url = await webdriver.getCurrentUrl()
        return url.includes('/logs/')
      }),
      8000
    )

    // Verify we're on the logs page (action was executed)
    const url = await webdriver.getCurrentUrl()
    expect(url).to.include('/logs/')
  })
})

