import {expect} from 'chai';

import {By} from 'selenium-webdriver';

describe('config: hiddenFooter', function () {
  before(async function () {
    runner.start('hiddenFooter')
  });

  after(async () => {
    runner.stop()
  });

  it('Footer is hidden', async () => {
    await webdriver.get('http://localhost:1337')

    let footer = await webdriver.findElement(By.tagName('footer'))

    expect(await footer.isDisplayed()).to.be.false
  })
})
