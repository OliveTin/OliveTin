import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'
import {
  getRootAndWait,
  openSidebar,
  getNavigationLinks,
  takeScreenshotOnFailure,
} from '../../lib/elements.js'

describe('config: dashboardAcls', function () {
  before(async function () {
    await runner.start('dashboardAcls')
  })

  after(async () => {
    await runner.stop()
  })

  afterEach(function () {
    takeScreenshotOnFailure(this.currentTest, webdriver)
  })

  it('hides ACL-restricted dashboards from guests in the side menu', async function () {
    await getRootAndWait()
    await openSidebar()

    const navigationLinks = await getNavigationLinks()
    expect(navigationLinks).to.not.be.empty

    const linkTexts = []
    for (const link of navigationLinks) {
      linkTexts.push(await link.getText())
    }

    expect(linkTexts).to.include('Public tools')
    expect(linkTexts).to.not.include('Services')
  })
})
