import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'
import { By, until, Condition } from 'selenium-webdriver'
//import * as waitOn from 'wait-on'
import { 
  getRootAndWait, 
  getActionButtons,
  takeScreenshotOnFailure,
} from '../lib/elements.js'

describe('config: general', function () {
  before(async function () {
    await runner.start('general')
  })

  after(async () => {
    await runner.stop()
  })

  afterEach(function () {
    takeScreenshotOnFailure(this.currentTest, webdriver);
  });

  it('Page title', async function () {
    await webdriver.get(runner.baseUrl())

    const title = await webdriver.getTitle()
    expect(title).to.be.equal("OliveTin")
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
    await getRootAndWait()

    const buttons = await getActionButtons(webdriver)

    expect(buttons).to.have.length(8)
  })

  it('Start dir action (popup)', async function () {
    await getRootAndWait()

    const buttons = await webdriver.findElements(By.css('[title="dir-popup"]'))

    expect(buttons).to.have.length(1)

    const buttonCMD = buttons[0]

    expect(buttonCMD).to.not.be.null

    buttonCMD.click()

    const dialog = await webdriver.findElement(By.id('execution-results-popup'))
    expect(await dialog.isDisplayed()).to.be.true

    const title = await webdriver.findElement(By.id('execution-dialog-title'))
    expect(await webdriver.wait(until.elementTextIs(title, 'dir-popup'), 2000))

    const dialogErr = await webdriver.findElement(By.id('big-error'))
    expect(dialogErr).to.not.be.null
    expect(await dialogErr.isDisplayed()).to.be.false
  })

  it('Start cd action (passive)', async function () {
    await getRootAndWait()

    const buttons = await webdriver.findElements(By.css('[title="cd-passive"]'))

    expect(buttons).to.have.length(1)

    const buttonCMD = buttons[0]

    expect(buttonCMD).to.not.be.null

    buttonCMD.click()

    const dialog = await webdriver.findElement(By.id('execution-results-popup'))
    expect(await dialog.isDisplayed()).to.be.false

    const title = await webdriver.findElement(By.id('execution-dialog-title'))
    expect(await title.getAttribute('innerText')).to.be.equal('?')

    const dialogErr = await webdriver.findElement(By.id('big-error'))
    expect(dialogErr).to.not.be.null
    expect(await dialogErr.isDisplayed()).to.be.false
  })

})
