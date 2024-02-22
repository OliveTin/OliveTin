import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'
import { By } from 'selenium-webdriver'
//import * as waitOn from 'wait-on'

describe('config: general', function () {
  before(async function () {
    await runner.start('general')
  })

  after(async () => {
    await runner.stop()
  })

  it('Page title', async function () {
    await webdriver.get(runner.baseUrl())

    const title = await webdriver.getTitle()
    expect(title).to.be.equal("OliveTin")
  })

  it('Footer contains promo', async function () {
    const ftr = await webdriver.findElement(By.tagName('footer')).getText()

    expect(ftr).to.contain('Documentation')
  })

  it('Default buttons are rendered', async function() {
    await webdriver.get(runner.baseUrl())

    // await webdriver.manage().setTimeouts({ implicit: 2000 })

    const buttons = await webdriver.findElement(By.id('root-group')).findElements(By.tagName('button'))

    expect(buttons).to.have.length(6)
  })
})
