import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'
import { By, until, Condition, Key } from 'selenium-webdriver'
import {
  getRootAndWait,
  getActionButtons,
  takeScreenshotOnFailure,
  waitForArgumentFormPage,
  waitForArgumentFormReady,
} from '../../lib/elements.js'


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

    await waitForArgumentFormPage(8000)
    await waitForArgumentFormReady(10000)

    await webdriver.wait(new Condition('wait for choice comboboxes', async () => {
      const boxes = await webdriver.findElements(By.css('main .choice-combobox'))
      return boxes.length >= 2
    }), 10000)

    const comboboxes = await webdriver.findElements(By.css('main .choice-combobox'))

    expect(comboboxes).to.have.length(2)

    const firstInput = await comboboxes[0].findElement(By.css('.choice-combobox-input'))
    await firstInput.click()
    await webdriver.wait(new Condition('wait for first combobox list', async () => {
      const lists = await comboboxes[0].findElements(By.css('.choice-combobox-list li'))
      return lists.length === 2
    }), 2000)

    await firstInput.sendKeys(Key.TAB)

    await webdriver.wait(new Condition('wait for second combobox list', async () => {
      const lists = await comboboxes[1].findElements(By.css('.choice-combobox-list li'))
      return lists.length === 3
    }), 2000)
  })
})
