import { By } from 'selenium-webdriver'
import fs from 'fs'
import { expect } from 'chai'
import { Condition } from 'selenium-webdriver'

export async function getActionButtons (dashboardTitle = null) {
  if (dashboardTitle == null) { 
    return await webdriver.findElement(By.id('contentActions')).findElements(By.tagName('button'))
  } else {
    return await webdriver.findElements(By.css('section[title="' + dashboardTitle + '"] button'))
  }
}

export function takeScreenshotOnFailure (test, webdriver) {
    if (test.state === 'failed') {
      const title = test.fullTitle();

      console.log(`Test failed, taking screenshot: ${title}`);
      takeScreenshot(webdriver, title);
    }
}

export function takeScreenshot (webdriver, title) {
  return webdriver.takeScreenshot().then((img) => {
    fs.mkdirSync('screenshots', { recursive: true });

	title = title.replaceAll(/[\(\)\|\*\<\>\:]/g, "_")
	title = 'failed-test.' + title

    fs.writeFileSync('screenshots/' + title + '.png', img, 'base64')
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

export async function closeSidebar() {
  await webdriver.findElement(By.id('sidebar-toggler-button')).click()

  const sidebar = await webdriver.findElement(By.id('mainnav'))

  const neededLeft = '-250px' // Assuming sidebar is closed at this position

  let lastLeft = ''

  await webdriver.wait(new Condition('wait for sidebar to close', async function() {
    const left = await sidebar.getCssValue('left')

    if (left !== lastLeft) {
      lastLeft = left
      console.log('Sidebar left changed to: ', left)
      return false
    } else {
      console.log('Sidebar closed, left is: *' + left, left === neededLeft ? ' (as expected)' : '')
      return left === neededLeft
    }
  }), 10000); // Wait up to 10 seconds for the sidebar to close
}

export async function openSidebar() {
  await webdriver.findElement(By.id('sidebar-toggler-button')).click()

  const sidebar = await webdriver.findElement(By.id('mainnav'))

  let lastLeft = 0

  await webdriver.wait(new Condition('wait for sidebar to open', async function() {
    const left = await sidebar.getCssValue('left')

    if (left !== lastLeft) {
      lastLeft = left
      console.log('Sidebar left changed to: ', left)
      return false
    } else {
      console.log('Sidebar opened, left is: ', left)
      return true
    }
  }));
}

export async function getNavigationLinks() {
  const navigationLinks = await webdriver.findElements(By.css('#navigation-links a'))

  return navigationLinks
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
