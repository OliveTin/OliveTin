import { describe, it, before, after } from 'mocha'
import { assert } from 'chai'
import { By } from 'selenium-webdriver'
import {
  getRootAndWait,
  getActionButtons,
  getNavigationLinks,
  openSidebar,
  closeSidebar,
  takeScreenshotOnFailure,
} from '../lib/elements.js'

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

    const actionsLink = navLinks[0];
    assert.isNotNull(actionsLink, 'Actions link should not be null')
    assert.equal(await actionsLink.getAttribute('title'), 'Actions', 'Actions link should have the title "Actions"')
    assert.isFalse(await actionsLink.isDisplayed(), 'Actions link should not be displayed when there are only dashboards')

    const firstDashboardLink = await webdriver.findElement(By.css('li[title="My Dashboard"]'), 'The first dashboard link should be present')
    assert.isNotNull(firstDashboardLink, 'First dashboard link should not be null')
    assert.isTrue(await firstDashboardLink.isDisplayed(), 'First dashboard link should be displayed')
    
    const actionButtons = await getActionButtons()

    assert.isArray(actionButtons, 'Action buttons should be an array')
    assert.lengthOf(actionButtons, 0, 'Action buttons should be empty when everything is added to the dashboard')

    const actionButtonsOnDashboard = await getActionButtons('MyDashboard')
    assert.isArray(actionButtonsOnDashboard, 'Action buttons on dashboard should be an array')
    assert.lengthOf(actionButtonsOnDashboard, 3, 'Action buttons on dashboard should have 3 buttons')
  })
})
