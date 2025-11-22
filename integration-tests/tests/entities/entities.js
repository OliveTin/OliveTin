import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'
import { By, until } from 'selenium-webdriver'
import { 
  getRootAndWait, 
  takeScreenshot,
  takeScreenshotOnFailure,
} from '../../lib/elements.js'

describe('config: entities', function () {
  before(async function () {
    await runner.start('entities')
  })

  after(async () => {
    await runner.stop()
  })

  afterEach(function () {
    takeScreenshotOnFailure(this.currentTest, webdriver);
  });

  it('Entity buttons are rendered', async function() {
    await getRootAndWait()

    // The old test was looking for #root-group, but that doesn't exist in the new Vue UI
    // Instead, we should look for action buttons directly
    const actionButtons = await webdriver.findElements(By.css('.action-button button'))
    expect(actionButtons).to.not.be.null
    expect(actionButtons).to.have.length(3)

    expect(await actionButtons[0].getAttribute('title')).to.be.equal('Ping server1')
    expect(await actionButtons[1].getAttribute('title')).to.be.equal('Ping server2')
    expect(await actionButtons[2].getAttribute('title')).to.be.equal('Ping server3')

    // Check that there's no error dialog visible
    const dialogErr = await webdriver.findElements(By.id('big-error'))
    if (dialogErr.length > 0) {
      expect(await dialogErr[0].isDisplayed()).to.be.false
    }
  })
})
