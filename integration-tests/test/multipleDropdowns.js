import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'
import { By, until, Condition } from 'selenium-webdriver'
import { 
  getRootAndWait, 
  getActionButtons,
  takeScreenshotOnFailure,
} from '../lib/elements.js'


describe('config: multipleDropdowns', function () {
  before(async function () {
    await runner.start('multipleDropdowns')
  })

  after(async () => {
    await runner.stop()
  })

  afterEach(function () {
    takeScreenshotOnFailure(this.currentTest, webdriver);
  });

  it('Multiple dropdowns are possible', async function() {
    await getRootAndWait()

    // Wait for action buttons to be rendered
    await webdriver.wait(new Condition('wait for action buttons', async () => {
      const btns = await webdriver.findElements(By.css('.action-button button'))
      return btns.length >= 2
    }), 10000)

    const buttons = await getActionButtons()

    let button = null
    for (const b of buttons) {
      const title = await b.getAttribute('title')

      if (title === 'Test multiple dropdowns') {
        button = b
      }
    }

    expect(buttons).to.have.length(2)
    expect(button).to.not.be.null

    await button.click()

    // Wait for navigation to argument form page
    await webdriver.wait(new Condition('wait for argument form page', async () => {
      const url = await webdriver.getCurrentUrl()
      return url.includes('/actionBinding/') && url.includes('/argumentForm')
    }), 8000)

    // Wait for form elements to be rendered
    await webdriver.wait(new Condition('wait for form elements', async () => {
      const selects = await webdriver.findElements(By.tagName('select'))
      return selects.length >= 2
    }), 5000)

    // Find the select elements after the wait condition
    const selects = await webdriver.findElements(By.tagName('select'))

    expect(selects).to.have.length(2)
    expect(await selects[0].findElements(By.tagName('option'))).to.have.length(2)
    expect(await selects[1].findElements(By.tagName('option'))).to.have.length(3)
  })
})
