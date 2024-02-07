import { expect } from 'chai'
import { By, until } from 'selenium-webdriver'

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

    const button = await webdriver.findElement(By.id('actionButton-bdc45101bbd12c1397557790d9f3e059')).findElement(By.tagName('button'))

    expect(button).to.not.be.undefined

    await button.click()

    const dialog = await webdriver.findElement(By.id('argument-popup'))

    await webdriver.wait(until.elementIsVisible(dialog), 2000)

    const selects = await dialog.findElements(By.tagName('select'))

    expect(selects).to.have.length(2)
    expect(await selects[0].findElements(By.tagName('option'))).to.have.length(2)
    expect(await selects[1].findElements(By.tagName('option'))).to.have.length(3)
  })
})
