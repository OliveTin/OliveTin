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

  it('zzz', async function () {
    console.log('zzz')
  })

  it('aaa', async function () {
    console.log('aaa')
  })

  it('Page title', async function () {
    /*
    console.log("Page title started")
    await webdriver.get(runner.baseUrl())

    const title = await webdriver.getTitle()
    expect(title).to.be.equal("OliveTin")
    */
  })

  it('Page title2', async function () {
    /*
    await webdriver.get(runner.baseUrl())

    const title = await webdriver.getTitle()
    expect(title).to.be.equal("OliveTin")
    */
  })


  it('Footer contains promo', async function () {
    const ftr = await webdriver.findElement(By.tagName('footer')).getText()

    expect(ftr).to.contain('Documentation')
  })

  it('Default buttons are rendered', async function() {
    await webdriver.get(runner.baseUrl())

    const buttons = await webdriver.findElement(By.id('root-group')).findElements(By.tagName('button'))

    expect(buttons).to.have.length(8)
  })

  it('Start date action (popup)', async function() {
    await webdriver.get(runner.baseUrl())

    const buttons = await webdriver.findElements(By.css('[title="date-popup"]'))

    expect(buttons).to.have.length(1)

    const buttonDate = buttons[0]

    expect(buttonDate).to.not.be.null

    buttonDate.click()

    const dialog = await webdriver.findElement(By.id('execution-results-popup'))
    expect(await dialog.isDisplayed()).to.be.true

    const title = await webdriver.findElement(By.id('execution-dialog-title'))
    expect(await title.getAttribute('innerText')).to.be.equal('date-popup')

    const dialogErr = await webdriver.findElement(By.id('big-error'))
    expect(dialogErr).to.not.be.null
    expect(await dialogErr.isDisplayed()).to.be.false
  })

  it('Start date action (passive)', async function() {
    await webdriver.get(runner.baseUrl())

    const buttons = await webdriver.findElements(By.css('[title="date-passive"]'))

    expect(buttons).to.have.length(1)

    const buttonDate = buttons[0]

    expect(buttonDate).to.not.be.null

    buttonDate.click()

    const dialog = await webdriver.findElement(By.id('execution-results-popup'))
    expect(await dialog.isDisplayed()).to.be.false

    const title = await webdriver.findElement(By.id('execution-dialog-title'))
    expect(await title.getAttribute('innerText')).to.be.equal('?')

    const dialogErr = await webdriver.findElement(By.id('big-error'))
    expect(dialogErr).to.not.be.null
    expect(await dialogErr.isDisplayed()).to.be.false
  })

})
