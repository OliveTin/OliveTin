import * as process from 'node:process'
import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'
import { By, Condition } from 'selenium-webdriver'
import {
  takeScreenshot,
  takeScreenshotOnFailure,
  findExecutionDialog,
  requireExecutionDialogStatus,
  getRootAndWait,
  getActionButton
} from '../../lib/elements.js'

describe('config: sleep', function () {
  before(async function () {
    await runner.start('sleep')
  })

  after(async () => {
    await runner.stop()
  })

  afterEach(function () {
    takeScreenshotOnFailure(this.currentTest, webdriver);
  });

  it('Sleep action kill', async function() {
    await getRootAndWait()

    const btnSleep = await getActionButton(webdriver, "Sleep")

    await btnSleep.click()

    await webdriver.sleep(1000)

    const dialog = await findExecutionDialog(webdriver)

    expect(await dialog.isDisplayed()).to.be.true

    await requireExecutionDialogStatus(webdriver, "Still running...")

    const killButton = await webdriver.findElement(By.id('execution-dialog-kill-action'))
    expect(killButton).to.not.be.undefined

    await killButton.click()

    await requireExecutionDialogStatus(webdriver, "Completed (Exit code: -1)")
  })
})
