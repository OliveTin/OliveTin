import {expect} from 'chai';
import {By} from 'selenium-webdriver';

describe('config: hiddenNav', function () {
  before(async function () {
    runner.start('hiddenNav')
  });

  after(async () => {
    runner.stop()
  });

  it('nav is hidden', async () => {
    await webdriver.get('http://localhost:1337')

    let toggler = await webdriver.findElement(By.id('sidebar-toggle-wrapper'))

    console.log("DOM", toggler.dom)

    expect(await toggler.isDisplayed()).to.be.false
  })
})
