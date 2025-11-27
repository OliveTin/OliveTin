import { describe, it, before, after } from 'mocha'
import { assert, expect } from 'chai'
import { By } from 'selenium-webdriver'
import {
  getRootAndWait,
  getActionButtons,
  getNavigationLinks,
  openSidebar,
  closeSidebar,
  takeScreenshotOnFailure,
} from '../../lib/elements.js'

describe('config: onlyDashboards', function () {
  before(async function () {
    await runner.start('onlyDashboards')
  })

  after(async () => {
    await runner.stop()
  })

  afterEach(function () {
    takeScreenshotOnFailure(this.currentTest, webdriver);
  });

  it('When there are only dashboards, actions are hidden', async function () {
    await getRootAndWait()

    await openSidebar()

    const navLinks = await getNavigationLinks()
    expect(navLinks).to.not.be.empty

    for (const link of navLinks) {
      console.log(await link.getAttribute('title'))
    }

    const firstLink = await navLinks[0];
    assert.isNotNull(firstLink, 'Actions link should not be null')

    assert.equal(await firstLink.getAttribute('title'), 'My Dashboard', 'First link should have the title "My Dashboard"')

    const firstDashboardLink = await webdriver.findElement(By.css('li[title="My Dashboard"]'), 'The first dashboard link should be present')
    assert.isNotNull(firstDashboardLink, 'First dashboard link should not be null')
    assert.isTrue(await firstDashboardLink.isDisplayed(), 'First dashboard link should be displayed')
    
    const actionButtonsOnDashboard = await getActionButtons()
    assert.isArray(actionButtonsOnDashboard, 'Action buttons on dashboard should be an array')
    assert.lengthOf(actionButtonsOnDashboard, 3, 'Action buttons on dashboard should have 3 buttons')
  })
})
