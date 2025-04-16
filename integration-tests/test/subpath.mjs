import { expect } from 'chai'
import { By } from 'selenium-webdriver'
import {
  getRootAndWait,
  getActionButtons,
  takeScreenshotOnFailure,
} from '../lib/elements.js'



describe('config: subpath', function () {
  before(async function () {
    await runner.start('subpath')
  })

  after(async () => {
    await runner.stop()
  })

  afterEach(function () {
    takeScreenshotOnFailure(this.currentTest, webdriver);
  });

  it('Page title', async function () {
    await webdriver.get(runner.baseUrl()+"/subpath/")

    const title = await webdriver.getTitle()
    expect(title).to.be.equal("OliveTin")
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
})
