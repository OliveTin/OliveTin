import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'

import { By } from 'selenium-webdriver'

describe('config: hiddenFooter', function () {
  before(async function () {
    await runner.start('hiddenFooter')
  })

  after(async () => {
    await runner.stop()
  })

  it('Check that footer is hidden', async () => {
    await webdriver.get(runner.baseUrl())

    const footer = await webdriver.findElement(By.tagName('footer'))

    expect(await footer.isDisplayed()).to.be.false
  })
})
