import { describe, it, before, after } from 'mocha'
import { expect, assert } from 'chai'
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
    expect(title).to.be.equal("Test - OliveTin")

    await openSidebar()

    const navigationLinks = await getNavigationLinks()
    assert.equal(navigationLinks.length, 5, 'Expected the nav to only have 5 links') // test dashboard + logs + diagnostics + entities + separator

    const firstLink = await navigationLinks[0]

    expect(await firstLink.getAttribute('title')).to.be.equal('Test', 'Expected the first link to be the actions link')

    const actionButtons = await getActionButtons()
    expect(actionButtons).to.have.length(5, 'Expected 5 action buttons')

    // Check that we have the expected number of fieldsets
    const allFieldsets = await webdriver.findElements(By.css('fieldset'))
    expect(allFieldsets).to.have.length(5, 'Expected 5 fieldsets total')
    
    // Check that we have fieldsets with the expected titles
    const fieldsetTitles = []
    for (let i = 0; i < allFieldsets.length; i++) {
      const legend = await allFieldsets[i].findElements(By.css('legend'))
      if (legend.length > 0) {
        const title = await legend[0].getText()
        fieldsetTitles.push(title)
      }
    }
    
    // We should have fieldsets for: Fieldset 1, Fieldset 2, and Actions fieldsets
    expect(fieldsetTitles).to.include('Fieldset 1')
    expect(fieldsetTitles).to.include('Fieldset 2')

  })
})
