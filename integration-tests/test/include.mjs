import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'
import { By, until } from 'selenium-webdriver'
import {
  getRootAndWait,
  getActionButtons,
  takeScreenshotOnFailure,
} from '../lib/elements.js'

describe('config: include', function () {
  this.timeout(30000)

  before(async function () {
    await runner.start('include')
  })

  after(async () => {
    await runner.stop()
  })

  afterEach(function () {
    takeScreenshotOnFailure(this.currentTest, webdriver);
  });

  it('Should load actions from base config and included files', async function () {
    await getRootAndWait()

    // Wait for the page to be ready
    await webdriver.wait(until.elementLocated(By.css('.action-button')), 10000)

    const buttons = await getActionButtons()
    
    // We should have:
    // 1. Base Action from config.yaml
    // 2. First Included Action from 00-first.yml
    // 3. Second Included Action from 01-second.yml
    expect(buttons.length).to.be.at.least(3, 'Should have at least 3 actions from base + includes')

    // Verify all actions are present
    const buttonTexts = await Promise.all(buttons.map(btn => btn.getText()))
    
    console.log('Found actions:', buttonTexts)
    
    // Text includes newline, so check with includes
    const allText = buttonTexts.join(' ')
    expect(allText).to.include('Base Action')
    expect(allText).to.include('First Included Action')
    expect(allText).to.include('Second Included Action')

    console.log('✓ Include directive loaded actions from all files')
  })
})

