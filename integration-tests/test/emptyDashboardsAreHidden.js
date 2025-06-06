import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'
import { By, until, Condition } from 'selenium-webdriver'
//import * as waitOn from 'wait-on'
import {
  getRootAndWait,
  takeScreenshotOnFailure,
} from '../lib/elements.js'

describe('config: empty dashboards are hidden', function () {
  before(async function () {
    await runner.start('emptyDashboardsAreHidden')
  })

  after(async () => {
    await runner.stop()
  })

  afterEach(function () {
    takeScreenshotOnFailure(this.currentTest, webdriver);
  });

  it('Test hidden dashboard', async function () {
    await getRootAndWait()

    const title = await webdriver.getTitle()
    expect(title).to.be.equal("OliveTin")

    await webdriver.findElement(By.id('sidebar-toggler-button')).click()

    const navigationLinks = await webdriver.findElements(By.css('#navigation-links a'))

    console.log('navigationLinks', navigationLinks)

    expect(navigationLinks).to.not.be.empty
    expect(navigationLinks.length).to.be.equal(1, 'Expected the nav to only have 1 link')

    const firstLinkId = await navigationLinks[0].getAttribute('id')

    expect(firstLinkId).to.be.equal('showActions', 'Expected the first link to be the actions link')
  })
})
