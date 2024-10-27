import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'
import { By, until } from 'selenium-webdriver'
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

    const buttons = await getActionButtons(webdriver)

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

    const dialog = await webdriver.findElement(By.id('argument-popup'))

    await webdriver.wait(until.elementIsVisible(dialog), 3500)

    const selects = await dialog.findElements(By.tagName('select'))

    expect(selects).to.have.length(2)
    expect(await selects[0].findElements(By.tagName('option'))).to.have.length(2)
    expect(await selects[1].findElements(By.tagName('option'))).to.have.length(3)
  })
})
