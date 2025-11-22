import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'
import { By, until, Condition } from 'selenium-webdriver'
//import * as waitOn from 'wait-on'
import {
  getRootAndWait,
  openSidebar,
  getNavigationLinks,
  takeScreenshotOnFailure,
} from '../../lib/elements.js'

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

    await openSidebar()

    const title = await webdriver.getTitle()
    expect(title).to.be.equal("Actions - OliveTin")

    const navigationLinks = await getNavigationLinks()
    expect(navigationLinks).to.not.be.empty
    expect(navigationLinks.length).to.be.equal(4, 'Expected the nav to only have 4 links')
  })
})
