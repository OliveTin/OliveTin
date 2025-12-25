import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'
import { By } from 'selenium-webdriver'
import {
  getRootAndWait,
  takeScreenshotOnFailure,
} from '../../lib/elements.js'

describe('config: customTheme', function () {
  before(async function () {
    await runner.start('customTheme')
  })

  after(async () => {
    await runner.stop()
  })

  afterEach(function () {
    takeScreenshotOnFailure(this.currentTest, webdriver)
  })

  it('Custom theme CSS is loaded and accessible', async function () {
    await getRootAndWait()

    // Fetch the theme CSS directly
    const themeCssUrl = runner.baseUrl() + 'theme.css'
    await webdriver.get(themeCssUrl)

    // Wait for the page to load
    await webdriver.sleep(500)

    // Get the page source (should be CSS content)
    const pageSource = await webdriver.getPageSource()

    // Verify the theme CSS contains our custom styles
    expect(pageSource).to.include('background-color: #ff6b6b')
    expect(pageSource).to.include('border: 5px solid #4ecdc4')
    expect(pageSource).to.include('Custom theme for integration testing')
  })

  it('Custom theme styles are applied to the page', async function () {
    await getRootAndWait()

    // Wait a bit for CSS to be fully loaded
    await webdriver.sleep(500)

    // Get computed background color of body
    const backgroundColor = await webdriver.executeScript(
      'return window.getComputedStyle(document.body).backgroundColor'
    )

    // The background color should be rgb(255, 107, 107) which is #ff6b6b
    // Different browsers might return different formats, so we check for the RGB values
    expect(backgroundColor).to.include('255, 107, 107')

    // Get computed border color of content element
    const borderColor = await webdriver.executeScript(
      'const el = document.getElementById("content"); return el ? window.getComputedStyle(el).borderColor : null'
    )

    // The border color should be rgb(78, 205, 196) which is #4ecdc4
    // borderColor might return "rgb(78, 205, 196)" or similar format
    expect(borderColor).to.include('78, 205, 196')
  })

  it('Theme CSS link is present in the HTML', async function () {
    await getRootAndWait()

    // Find the theme.css link in the head
    const themeLink = await webdriver.findElement(By.css('link[href="/theme.css"]'))

    expect(themeLink).to.not.be.null

    // Verify it's a stylesheet
    const rel = await themeLink.getAttribute('rel')
    expect(rel).to.equal('stylesheet')

    const type = await themeLink.getAttribute('type')
    expect(type).to.equal('text/css')
  })
})

