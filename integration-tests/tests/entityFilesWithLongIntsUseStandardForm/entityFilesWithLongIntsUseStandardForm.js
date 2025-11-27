// Issue: https://github.com/OliveTin/OliveTin/issues/616
import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'
import { By, until, Condition } from 'selenium-webdriver'
import { 
  getRootAndWait, 
  getActionButtons,
  takeScreenshotOnFailure,
} from '../../lib/elements.js'

describe('config: entities', function () {
  before(async function () {
    await runner.start('entityFilesWithLongIntsUseStandardForm')
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

    // Test INT with 10 numbers
    const buttonInt10 = await buttons[2]   
    expect(await buttonInt10.getAttribute('title')).to.be.equal('Test me INT with 10 numbers')
    await buttonInt10.click()

    // Wait for navigation to execution view
    await webdriver.wait(new Condition('wait for execution view', async () => {
      const url = await webdriver.getCurrentUrl()
      return url.includes('/logs/') && !url.endsWith('/logs')
    }), 10000)

    // Wait for execution to complete - look for the execution status
    await webdriver.wait(new Condition('wait for execution status', async () => {
      const statusElement = await webdriver.findElements(By.id('execution-dialog-status'))
      return statusElement.length > 0
    }), 15000)

    // Check that the execution completed successfully by looking at the status
    const statusElement = await webdriver.findElement(By.id('execution-dialog-status'))
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
