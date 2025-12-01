import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'
import { By } from 'selenium-webdriver'
import {
  getRootAndWait,
  takeScreenshotOnFailure,
} from '../../lib/elements.js'

describe('config: pageTitle', function () {
  before(async function () {
    await runner.start('pageTitle')
  })

  after(async () => {
    await runner.stop()
  })

  afterEach(function () {
    takeScreenshotOnFailure(this.currentTest, webdriver);
  });

  it('Init API returns custom pageTitle from config', async function () {
    await getRootAndWait()

    // Check that the Init API response (available via window.initResponse) contains pageTitle
    // This is how the frontend accesses it, so it's the most reliable way to test
    const initResponse = await webdriver.executeScript('return window.initResponse')
    
    expect(initResponse).to.not.be.null
    expect(initResponse).to.have.own.property('pageTitle')
    expect(initResponse.pageTitle).to.equal('Custom Test Title')
  })

  it('Header displays custom pageTitle from init response', async function () {
    await getRootAndWait()

    // Check that the pageTitle from init response is used in the header
    // First verify the init response has the correct pageTitle
    const pageTitle = await webdriver.executeScript('return window.initResponse?.pageTitle')
    expect(pageTitle).to.equal('Custom Test Title')

    // The Header component from picocrank should render the title prop
    // Check for the title in the header element
    const header = await webdriver.findElement(By.tagName('header'))
    const headerText = await header.getText()
    
    // The header should contain the custom page title
    expect(headerText).to.include('Custom Test Title')
  })
})

