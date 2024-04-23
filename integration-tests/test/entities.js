import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'
import { By, until } from 'selenium-webdriver'
import { getRootAndWait, takeScreenshot } from '../lib/elements.js'

describe('config: entities', function () {
  before(async function () {
    await runner.start('entities')
  })

  after(async () => {
    await runner.stop()
  })

  it('Entity buttons are rendered', async function() {
    await getRootAndWait()

    const buttons = await webdriver.findElement(By.id('root-group')).findElements(By.tagName('button'))
    expect(buttons).to.not.be.null
    expect(buttons).to.have.length(3)

    expect(await buttons[0].getAttribute('title')).to.be.equal('Ping server1')
    expect(await buttons[1].getAttribute('title')).to.be.equal('Ping server2')
    expect(await buttons[2].getAttribute('title')).to.be.equal('Ping server3')

    const dialogErr = await webdriver.findElement(By.id('big-error'))
    expect(dialogErr).to.not.be.null
    expect(await dialogErr.isDisplayed()).to.be.false
  })
})
