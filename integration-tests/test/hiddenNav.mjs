import { expect } from 'chai'
import { By } from 'selenium-webdriver'
import { 
  getRootAndWait, 
  getActionButtons,
  takeScreenshotOnFailure,
} from '../lib/elements.js'


describe('config: hiddenNav', function () {
  before(async function () {
    await runner.start('hiddenNav')
  })

  after(async () => {
    await runner.stop()
  })

  afterEach(function () {
    takeScreenshotOnFailure(this.currentTest, webdriver);
  });

  it('nav is hidden', async () => {
    await webdriver.get(runner.baseUrl())

    const toggler = await webdriver.findElement(By.tagName('header'))

    expect(await toggler.isDisplayed()).to.be.false
  })
})
