import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'
import { By, until, Condition } from 'selenium-webdriver'
//import * as waitOn from 'wait-on'
import {
  getRootAndWait,
  getActionButtons,
  openSidebar,
  getNavigationLinks,
  takeScreenshotOnFailure,
} from '../lib/elements.js'

describe('config: dashboards with basic fieldsets', function () {
  before(async function () {
    await runner.start('dashboardsWithBasicFieldsets')
  })

  after(async () => {
    await runner.stop()
  })

  afterEach(function () {
    takeScreenshotOnFailure(this.currentTest, webdriver);
  });

  it('Dashboards with basic fieldsets', async function () {
    await getRootAndWait()

    const title = await webdriver.getTitle()
    expect(title).to.be.equal("OliveTin » Test")

    const navigationLinks = await getNavigationLinks()
    expect(navigationLinks.length).to.be.equal(2, 'Expected the nav to only have 2 links')

    const firstLink = await navigationLinks[0]

    expect(await firstLink.getAttribute('id')).to.be.equal('showActions', 'Expected the first link to be the actions link')
    expect(await firstLink.isDisplayed()).to.be.false
    
    const secondLink = await navigationLinks[1]
    expect(await secondLink.getAttribute('href')).to.be.equal('http://localhost:1337/Test', 'Expected the second link to be the test dashboard with basic fieldsets link')

    const actionButtons = await getActionButtons('Test')
    expect(actionButtons).to.have.length(5, 'Expected 5 action buttons')

    const fieldsets = await webdriver.findElements(By.css('section[title="Test"] fieldset'))
    expect(fieldsets).to.have.length(3, 'Expected 3 fieldsets in the Test dashboard')

  })
})
