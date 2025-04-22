import { expect } from 'chai'
import { By, until } from 'selenium-webdriver'
import {
  getRootAndWait,
  waitForInitialMarshall,
  takeScreenshotOnFailure,
} from '../lib/elements.js'



describe('config: subpath', function () {
  before(async function () {
    await runner.start('subpath')
  })

  after(async () => {
    await runner.stop()
  })

  afterEach(function () {
    takeScreenshotOnFailure(this.currentTest, webdriver);
  });

  it('Page title', async function () {
    await webdriver.get(runner.baseUrl())

    const title = await webdriver.getTitle()
    expect(title).to.be.equal("OliveTin")
  })

  it('Start dir action (popup)', async function () {
    await getRootAndWait()

    const buttons = await webdriver.findElements(By.css('[title="dir-popup"]'))

    expect(buttons).to.have.length(1)

    const buttonCMD = buttons[0]

    expect(buttonCMD).to.not.be.null

    buttonCMD.click()

    const dialog = await webdriver.findElement(By.id('execution-results-popup'))
    expect(await dialog.isDisplayed()).to.be.true

    const title = await webdriver.findElement(By.id('execution-dialog-title'))
    expect(await webdriver.wait(until.elementTextIs(title, 'dir-popup'), 2000)).to.be.exist

    const dialogErr = await webdriver.findElement(By.id('big-error'))
    expect(dialogErr).to.not.be.null
    expect(await dialogErr.isDisplayed()).to.be.false
  })

  it('Load Logs Page', async function () {
    await webdriver.get(runner.baseUrl() + "/logs")
    await waitForInitialMarshall()

    const title = await webdriver.getTitle()
    expect(title).to.be.equal("OliveTin » Logs")
  })

  it('Load Diagnostics Page', async function () {
    await webdriver.get(runner.baseUrl() + "/diagnostics")
    await waitForInitialMarshall()

    const title = await webdriver.getTitle()
    expect(title).to.be.equal("OliveTin » Diagnostics")
  })
  it('Check WebSocket is connected', async function () {
    await getRootAndWait()
    const websocketStatus = await webdriver.findElement(By.id('serverConnectionWebSocket'))
    expect(await webdriver.wait(until.elementTextIs(websocketStatus, "WebSocket"), 2000)).to.be.exist
    const classAttribute = await websocketStatus.getAttribute("class");
    if (classAttribute && classAttribute.includes('error')) {
      throw new Error('Test failed: Element has the undesired class "error".');
    }
  })
  it('Check Rest is connected', async function () {
    await getRootAndWait()
    const restStatus = await webdriver.findElement(By.id('serverConnectionRest'))
    expect(await webdriver.wait(until.elementTextIs(restStatus, "REST"), 2000)).to.be.exist
    const classAttribute = await restStatus.getAttribute("class");
    if (classAttribute && classAttribute.includes('error')) {
      throw new Error('Test failed: Element has the undesired class "error".');
    }
  })
  it('Load MyServers/Hypervisors Page', async function () {
    await webdriver.get(runner.baseUrl() + "/MyServers/Hypervisors")
    await waitForInitialMarshall()
    const buttons = await webdriver.findElements(By.css('[title="Ping hypervisor1"]'))

    expect(buttons).to.have.length(1)

    const buttonCMD = buttons[0]

    expect(buttonCMD).to.not.be.null

    buttonCMD.click()
  })
  it('Load Logs Page for successful hypervisor ping', async function () {
    await webdriver.get(runner.baseUrl() + "/logs")
    await waitForInitialMarshall()

    const title = await webdriver.getTitle()
    expect(title).to.be.equal("OliveTin » Logs")
    const logRows = await webdriver.findElements(By.css('.log-row[title="Ping hypervisor1"]'))

    expect(logRows).to.have.length(1)

    const logRow = logRows[0]
    let actionStatus = await logRow.findElement(By.css('.action-status'));
    expect(await webdriver.wait(until.elementTextIs(actionStatus, "Completed"), 2000)).to.be.exist
  })
})
