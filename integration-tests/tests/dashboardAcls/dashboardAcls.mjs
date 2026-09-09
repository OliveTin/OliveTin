import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'
import {
  getRootAndWait,
  openSidebar,
  getNavigationLinkTitles,
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

    const linkTexts = await getNavigationLinkTitles()
    expect(linkTexts).to.not.be.empty

    expect(linkTexts).to.include('Public tools')
    expect(linkTexts).to.not.include('Services')
  })
})
