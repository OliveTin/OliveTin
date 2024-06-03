import * as process from 'node:process'
import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'
import { By, Condition } from 'selenium-webdriver'
import {
  takeScreenshot,
  findExecutionDialog,
  requireExecutionDialogStatus,
  getRootAndWait,
  getActionButton
} from '../lib/elements.js'

describe('config: sleep', function () {
  before(async function () {
    await runner.start('sleep')
  })

  after(async () => {
    await runner.stop()
  })

  it('Sleep action kill', async function() {
    await getRootAndWait()

    const btnSleep = await getActionButton(webdriver, "Sleep")

    const dialog = await findExecutionDialog(webdriver)

    expect(await dialog.isDisplayed()).to.be.false

    await btnSleep.click()

    expect(await dialog.isDisplayed()).to.be.true

    await requireExecutionDialogStatus(webdriver, "unknown")

    const killButton = await webdriver.findElement(By.id('execution-dialog-kill-action'))
    expect(killButton).to.not.be.undefined

    await killButton.click()

    // FIXME hack
    if (process.env.CI === 'true') {
      await requireExecutionDialogStatus(webdriver, "Non-Zero Exite")
    }
  })
})
