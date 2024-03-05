import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'
import { By, until } from 'selenium-webdriver'
import { getActionButtons } from '../lib/elements.js'

describe('config: multipleDropdowns', function () {
  before(async function () {
    await runner.start('multipleDropdowns')
  })

  after(async () => {
    await runner.stop()
  })

  it('Multiple dropdowns are possible', async function() {
    await webdriver.get(runner.baseUrl())
    await webdriver.manage().setTimeouts({ implicit: 2000 })

    const buttons = await getActionButtons(webdriver)

    let button = null
    for (const b of buttons) {
      const title = await b.getAttribute('title')

      console.log('title: ' + title)
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
