import { describe, it, before, after } from 'mocha'
import { expect, assert } from 'chai'
import { By } from 'selenium-webdriver'
import {
  getRootAndWait,
  getActionButtons,
  getNavigationLinks,
  openSidebar,
  takeScreenshotOnFailure,
} from '../../lib/elements.js'

describe('config: multi-dashboard-includes', function () {
  this.timeout(30000)

  before(async function () {
    await runner.start('multi-dashboard-includes')
  })

  after(async () => {
    await runner.stop()
  })

  afterEach(function () {
    takeScreenshotOnFailure(this.currentTest, webdriver);
  });

  async function clickNavigationLinkByTitle (title) {
    await openSidebar()

    const navigationLinks = await getNavigationLinks()
    assert.isAbove(navigationLinks.length, 0, 'Expected at least one navigation link')

    const matching = []
    for (const li of navigationLinks) {
      const liTitle = await li.getAttribute('title')
      if (liTitle === title) {
        matching.push(li)
      }
    }

    assert.strictEqual(matching.length, 1, `Expected exactly one navigation link with title "${title}"`)

    await matching[0].click()
  }

  async function getActionTitlesOnDashboard (dashboardTitle = null) {
    const buttons = await getActionButtons(dashboardTitle)
    const titles = []

    for (const button of buttons) {
      titles.push(await button.getAttribute('title'))
    }

    return titles
  }

  it('Should expose both dashboards from included files in navigation', async function () {
    await getRootAndWait()

    await openSidebar()
    const navigationLinks = await getNavigationLinks()
    assert.isAbove(navigationLinks.length, 0, 'Expected navigation to have at least one link')

    const titles = []
    for (const li of navigationLinks) {
      titles.push(await li.getAttribute('title'))
    }

    expect(titles).to.include('First Dashboard')
    expect(titles).to.include('Second Dashboard')
  })

  it('First Dashboard shows First and Second inline actions from include', async function () {
    await getRootAndWait()

    // Navigate to "First Dashboard"
    await clickNavigationLinkByTitle('First Dashboard')

    // Buttons on this dashboard only
    const titles = await getActionTitlesOnDashboard('First Dashboard')

    expect(titles).to.include('First Action')
    expect(titles).to.include('Second Action')

    // Ensure actions from the second dashboard are not rendered here
    expect(titles).to.not.include('Third Action')
    expect(titles).to.not.include('Forth Action')
  })

  it('Second Dashboard shows Third and Forth inline actions from include', async function () {
    await getRootAndWait()

    // Navigate to "Second Dashboard"
    await clickNavigationLinkByTitle('Second Dashboard')

    const titles = await getActionTitlesOnDashboard('Second Dashboard')

    expect(titles).to.include('Third Action')
    expect(titles).to.include('Forth Action')

    // Ensure actions from the first dashboard are not rendered here
    expect(titles).to.not.include('First Action')
    expect(titles).to.not.include('Second Action')
  })

})


