import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'
import { By } from 'selenium-webdriver'
//import * as waitOn from 'wait-on'
import {
  getRootAndWait,
  getActionButtons,
  takeScreenshotOnFailure,
  openSidebar,
  findSidebarNavHref,
  waitForCurrentUrl,
  waitForSelectorCount,
} from '../../lib/elements.js'

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
    await getRootAndWait()

    const title = await webdriver.getTitle()
    expect(title).to.be.equal("Actions - OliveTin")
  })

  it('navbar contains default policy links', async function () {
    await getRootAndWait()
    await openSidebar()


    const logsLink = await findSidebarNavHref('/logs')
    const diagnosticsLink = await findSidebarNavHref('/diagnostics')

    expect(logsLink).to.not.be.empty
    expect(diagnosticsLink).to.not.be.empty
  })

  it('Footer contains promo', async function () {
    const ftr = await webdriver.findElement(By.tagName('footer')).getText()

    expect(ftr).to.contain('Documentation')
  })

  it('Default buttons are rendered', async function() {
    await getRootAndWait()

    await waitForSelectorCount('[title="dir-popup"], [title="cd-passive"], .action-button button', 1, 10000)

    const buttons = await getActionButtons()
    expect(buttons.length).to.be.greaterThanOrEqual(4)
  })

  it('Start dir action (popup)', async function () {
    await getRootAndWait()

    await waitForSelectorCount('[title="dir-popup"]', 1, 10000)

    const buttons = await webdriver.findElements(By.css('[title="dir-popup"]'))

    expect(buttons).to.have.length(1)

    const buttonCMD = buttons[0]

    expect(buttonCMD).to.not.be.null

    buttonCMD.click()

    // New UI navigates to /logs/<id> instead of showing old dialog
    await waitForCurrentUrl((url) => url.includes('/logs/'), 8000)
  })

  it('Start cd action (passive)', async function () {
    await getRootAndWait()

    await waitForSelectorCount('[title="cd-passive"]', 1, 10000)

    const buttons = await webdriver.findElements(By.css('[title="cd-passive"]'))

    expect(buttons).to.have.length(1)

    const buttonCMD = buttons[0]

    expect(buttonCMD).to.not.be.null

    buttonCMD.click()

    // Should not navigate to logs for passive action; brief pause allows spurious navigation to show up
    await webdriver.sleep(500)
    const url = await webdriver.getCurrentUrl()
    expect(url.includes('/logs/')).to.be.false
  })

})
