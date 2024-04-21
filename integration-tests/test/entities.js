import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'
import { By } from 'selenium-webdriver'
//import * as waitOn from 'wait-on'

describe('config: entities', function () {
  before(async function () {
    await runner.start('entities')
  })

  after(async () => {
    await runner.stop()
  })

  it('Entity buttons are rendered', async function() {
    await webdriver.get(runner.baseUrl())

    //await webdriver.manage().setTimeouts({ implicit: 2000 })

    const buttons = await webdriver.findElement(By.id('root-group')).findElements(By.tagName('button'))

    expect(buttons).to.have.length(3)
    expect(await buttons[0].getAttribute('title')).to.be.equal('Ping server1')
    expect(await buttons[1].getAttribute('title')).to.be.equal('Ping server2')
    expect(await buttons[2].getAttribute('title')).to.be.equal('Ping server3')
  })
})
