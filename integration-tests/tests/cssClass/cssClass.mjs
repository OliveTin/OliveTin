import { describe, it, before, after, afterEach } from 'mocha'
import { expect } from 'chai'
import { By } from 'selenium-webdriver'
import {
  getRootAndWait,
  takeScreenshotOnFailure,
} from '../../lib/elements.js'

describe('config: cssClass', function () {
  before(async function () {
    await runner.start('cssClass')
  })

  after(async () => {
    await runner.stop()
  })

  afterEach(function () {
    takeScreenshotOnFailure(this.currentTest, webdriver)
  })

  it('cssClass is applied to action button (link component)', async function () {
    await getRootAndWait()

    const buttonWithClass = await webdriver.findElements(By.css('.action-button button.test-custom-class'))
    expect(buttonWithClass).to.have.length.at.least(1, 'Action button should have cssClass test-custom-class on the button')

    const classAttr = await buttonWithClass[0].getAttribute('class')
    expect(classAttr).to.include('test-custom-class')
  })

  it('custom theme applies background color to action button via cssClass', async function () {
    await getRootAndWait()

    const buttonWithClass = await webdriver.findElements(By.css('.action-button button.test-custom-class'))
    expect(buttonWithClass).to.have.length.at.least(1, 'Action button with test-custom-class should exist')

    const bgColor = await buttonWithClass[0].getCssValue('background-color')
    expect(bgColor, 'Theme theme.css should set .action-button button.test-custom-class background to rgb(32, 64, 128)')
      .to.match(/rgba?\(\s*32\s*,\s*64\s*,\s*128\s*(,\s*1)?\s*\)/)
  })

  it('cssClass override: style rule targeting custom class wins over component styles', async function () {
    await getRootAndWait()

    const buttonWithClass = await webdriver.findElements(By.css('.action-button button.test-custom-class'))
    expect(buttonWithClass).to.have.length.at.least(1)

    const beforePx = await buttonWithClass[0].getCssValue('border-top-width')
    await webdriver.executeScript(`
      var style = document.getElementById('cssclass-test-override-style');
      if (!style) {
        style = document.createElement('style');
        style.id = 'cssclass-test-override-style';
        style.textContent = '.test-custom-class { border-top-width: 31px !important; }';
        document.head.appendChild(style);
      } else {
        style.textContent = '.test-custom-class { border-top-width: 31px !important; }';
      }
    `)
    await webdriver.sleep(150)

    const afterPx = await buttonWithClass[0].getCssValue('border-top-width')
    const afterNum = parseInt(afterPx, 10)
    expect(afterNum).to.be.greaterThan(10, 'Override targeting cssClass should win over component 1px (before=' + beforePx + ' after=' + afterPx + ') (#804)')
  })

  it('cssClass is applied to display component', async function () {
    await getRootAndWait()

    const displayElements = await webdriver.findElements(By.css('.display.test-display-class'))
    expect(displayElements).to.have.length.at.least(1, 'Display component with cssClass test-display-class should be in DOM')

    const classAttr = await displayElements[0].getAttribute('class')
    expect(classAttr).to.include('test-display-class')
  })
})
