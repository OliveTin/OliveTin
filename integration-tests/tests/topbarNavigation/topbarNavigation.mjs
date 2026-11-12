import { By, until } from 'selenium-webdriver'
import { expect } from 'chai'
import {
  getRootAndWait,
  takeScreenshotOnFailure,
} from '../../lib/elements.js'

describe('config: sectionNavigationStyle topbar', function () {
  let originalWindowRect

  before(async function () {
    originalWindowRect = await webdriver.manage().window().getRect()
    await runner.start('topbarNavigation')
  })

  after(async () => {
    await webdriver.manage().window().setRect(originalWindowRect)
    await runner.stop()
  })

  afterEach(function () {
    takeScreenshotOnFailure(this.currentTest, webdriver)
  })

  it('shows dashboard links in the top bar', async () => {
    await webdriver.manage().window().setRect({ width: 1600, height: 900 })
    await getRootAndWait()

    const betaLink = await webdriver.wait(
      until.elementLocated(By.xpath("//nav[contains(@class, 'topbar')]//a[contains(., 'Beta')]")),
      5000
    )
    await betaLink.click()

    await webdriver.wait(until.urlContains('Beta'), 5000)
    expect(await webdriver.getCurrentUrl()).to.contain('Beta')
  })
})
