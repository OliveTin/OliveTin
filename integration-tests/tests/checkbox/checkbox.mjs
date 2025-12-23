import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'
import { By, Condition } from 'selenium-webdriver'
import {
  getRootAndWait,
  getActionButton,
  takeScreenshotOnFailure,
} from '../../lib/elements.js'

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
    await getRootAndWait()

    const btn = await getActionButton(webdriver, 'Test checkbox argument')

    await btn.click()

    // Wait for navigation to argument form page
    await webdriver.wait(
      new Condition('wait for argument form page', async () => {
        const url = await webdriver.getCurrentUrl()
        return url.includes('/actionBinding/') && url.includes('/argumentForm')
      }),
      8000
    )

    // Find the checkbox input field
    const checkboxInput = await webdriver.findElement(By.id('confirm'))

    // Verify it's an input of type checkbox
    const tagName = await checkboxInput.getTagName()
    expect(tagName).to.equal('input')

    const inputType = await checkboxInput.getAttribute('type')
    expect(inputType).to.equal('checkbox')

    // Verify the label is present
    const label = await webdriver.findElement(By.css('label[for="confirm"]'))
    expect(await label.getText()).to.contain('Confirm option')
  })

  it('Checkbox argument can be toggled and submitted', async function () {
    await getRootAndWait()

    const btn = await getActionButton(webdriver, 'Test checkbox argument')

    await btn.click()

    // Wait for navigation to argument form page
    await webdriver.wait(
      new Condition('wait for argument form page', async () => {
        const url = await webdriver.getCurrentUrl()
        return url.includes('/actionBinding/') && url.includes('/argumentForm')
      }),
      8000
    )

    const checkboxInput = await webdriver.findElement(By.id('confirm'))

    // Toggle the checkbox
    await checkboxInput.click()

    // Small wait for Vue to process the change
    await webdriver.sleep(100)

    // Verify the checkbox is checked
    const isChecked = await checkboxInput.isSelected()
    expect(isChecked).to.be.true

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


