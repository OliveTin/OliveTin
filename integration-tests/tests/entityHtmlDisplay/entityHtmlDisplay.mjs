import { describe, it, before, after, afterEach } from 'mocha'
import { expect } from 'chai'
import { By } from 'selenium-webdriver'
import {
  getRootAndWait,
  takeScreenshotOnFailure,
  waitForDashboardLoaded,
} from '../../lib/elements.js'

describe('config: entityHtmlDisplay', function () {
  before(async function () {
    await runner.start('entityHtmlDisplay')
  })

  after(async () => {
    await runner.stop()
  })

  afterEach(function () {
    takeScreenshotOnFailure(this.currentTest, webdriver)
  })

  it('renders entity HTML content inside display components (#804)', async function () {
    await getRootAndWait()
    await webdriver.get(runner.baseUrl() + 'dashboards/Html%20Display%20Dashboard')
    await waitForDashboardLoaded()

    const contentDiv = await webdriver.findElements(By.css('.display.test-html-display .content'))
    expect(contentDiv).to.have.length.at.least(1, 'Entity HTML should render inside the display component')

    const text = await contentDiv[0].getText()
    expect(text).to.equal('entity-html-test')
  })
})
