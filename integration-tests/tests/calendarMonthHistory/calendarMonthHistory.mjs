import { By, until } from 'selenium-webdriver'
import { expect } from 'chai'
import {
  getRootAndWait,
  takeScreenshotOnFailure,
  waitForCurrentUrl,
} from '../../lib/elements.js'

async function getCalendarTitle () {
  const title = await webdriver.wait(until.elementLocated(By.css('.calendar-title')), 5000)
  return title.getText()
}

describe('logs calendar: month is kept in the URL', function () {
  before(async function () {
    await runner.start('calendarMonthHistory')
  })

  after(async () => {
    await runner.stop()
  })

  afterEach(function () {
    takeScreenshotOnFailure(this.currentTest, webdriver)
  })

  it('returns to the browsed month when going back', async () => {
    await getRootAndWait()
    await webdriver.get(new URL('logs/calendar', runner.baseUrl()).toString())

    const currentMonthTitle = await getCalendarTitle()

    const previousButton = await webdriver.findElement(By.xpath("//div[contains(@class, 'calendar-nav-buttons')]//button[normalize-space()='‹']"))
    await previousButton.click()
    await waitForCurrentUrl(url => url.includes('month='))

    const previousMonthTitle = await getCalendarTitle()
    expect(previousMonthTitle).to.not.equal(currentMonthTitle)

    const monthUrl = await webdriver.getCurrentUrl()

    const backToListLink = await webdriver.findElement(By.css('a.button.inline-icon[href$="/logs"]'))
    await backToListLink.click()
    await waitForCurrentUrl(url => !url.includes('calendar'))

    await webdriver.navigate().back()
    await waitForCurrentUrl(url => url === monthUrl)

    await webdriver.wait(async () => (await getCalendarTitle()) === previousMonthTitle, 5000)
    expect(await getCalendarTitle()).to.equal(previousMonthTitle)
  })
})
