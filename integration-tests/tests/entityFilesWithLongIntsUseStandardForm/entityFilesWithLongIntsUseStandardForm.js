// Issue: https://github.com/OliveTin/OliveTin/issues/616
import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'
import { By } from 'selenium-webdriver'
import {
  getRootAndWait,
  getActionButtons,
  takeScreenshotOnFailure,
  waitForLogsPage,
  waitForExecutionComplete,
} from '../../lib/elements.js'

describe('config: entityFilesWithLongIntsUseStandardForm', function () {
  before(async function () {
    await runner.start('entityFilesWithLongIntsUseStandardForm')
    await getRootAndWait()
  })

  after(async () => {
    await runner.stop()
  })

  afterEach(function () {
    takeScreenshotOnFailure(this.currentTest, webdriver);
  });

  it('Entity buttons are rendered', async function() {
    await getRootAndWait()

    const buttons = await getActionButtons()

    expect(buttons).to.not.be.null
    expect(buttons).to.have.length(5)

    // Entity buttons are in numeric key order (0,1,2,3,4); first row is "INT with 10 numbers"
    const buttonInt10 = await buttons[0]
    expect(await buttonInt10.getAttribute('title')).to.be.equal('Test me INT with 10 numbers')
    await buttonInt10.click()

    await waitForLogsPage()
    await waitForExecutionComplete()

    const statusElement = await webdriver.findElement(By.css('.execution-dialog-status'))
    const statusText = await statusElement.getText()

    // The status should indicate success (not "Executing..." or "Failed")
    expect(statusText).to.not.include('Executing')
    expect(statusText).to.not.include('Failed')

    // Verify that we're on the execution page by checking the URL
    const currentUrl = await webdriver.getCurrentUrl()
    expect(currentUrl).to.include('/logs/')
    expect(currentUrl).to.not.equal(runner.baseUrl() + '/logs')

  });
});
