import { By } from 'selenium-webdriver'
import fs from 'fs'
import { expect } from 'chai'
import { Condition } from 'selenium-webdriver'

export async function getActionButtons (webdriver) {
  return await webdriver.findElement(By.id('contentActions')).findElements(By.tagName('button'))
}

export function takeScreenshot (webdriver) {
  return webdriver.takeScreenshot().then((img) => {
    fs.writeFileSync('out.png', img, 'base64')
  })
}

export async function getRootAndWait() {
  await webdriver.get(runner.baseUrl())
  await webdriver.wait(new Condition('wait for initial-marshal-complete', async function() {
    const body = await webdriver.findElement(By.tagName('body'))
    const attr = await body.getAttribute('initial-marshal-complete')

    if (attr == 'true') {
      return true
    } else {
      return false
    }
  }))
}

export async function requireExecutionDialogStatus (webdriver, expected) {
  // It seems that webdriver will not give us text if domStatus is hidden (which it will be until complete)
  await webdriver.executeScript('window.executionDialog.domExecutionDetails.hidden = false')

  await webdriver.wait(new Condition('wait for action to be running', async function () {
    const actual = await webdriver.executeScript('return window.executionDialog.domStatus.getText()')

    if (actual === expected) {
      return true
    } else {
      console.log('Waiting for domStatus text to be: ', expected, ', it is currently: ', actual)
      console.log(await webdriver.executeScript('return window.executionDialog.res'))
      return false
    }
  }))
}

export async function findExecutionDialog (webdriver) {
  return webdriver.findElement(By.id('execution-results-popup'))
}

export async function getActionButton (webdriver, title) {
  const buttons = await webdriver.findElements(By.css('[title="' + title + '"]'))

  expect(buttons).to.have.length(1)

  return buttons[0]
}
