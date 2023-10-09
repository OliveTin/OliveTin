import {expect} from 'chai';
import {By} from 'selenium-webdriver';

describe('config: general', function () {
  before(async function () {
    await runner.start('general')
  });

  after(async () => {
    await runner.stop()
  });

  it('Page title', async function () {
    await webdriver.get('http://localhost:1337')

    let title = await webdriver.getTitle();
    expect(title).to.be.equal("OliveTin")
  })

  it('Footer contains promo', async function () {
    let ftr = await webdriver.findElement(By.tagName('footer')).getText()

    expect(ftr).to.contain("Documentation")
  })

  it('Default buttons are rendered', async function() {
    await webdriver.get('http://localhost:1337')
    await webdriver.manage().setTimeouts({ implicit: 2000 });
    let buttons = await webdriver.findElement(By.id('root-group')).findElements(By.tagName('button'))

    expect(buttons).to.have.length(6);
  })
})
