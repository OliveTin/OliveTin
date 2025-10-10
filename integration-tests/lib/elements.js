import { By } from 'selenium-webdriver'
import fs from 'fs'
import { expect } from 'chai'
import { Condition } from 'selenium-webdriver'

export async function getActionButtons (dashboardTitle = null) {
  // New Vue UI renders action buttons using ActionButton.vue structure
  // Each button lives under a container with class .action-button
  if (dashboardTitle == null) {
    return await webdriver.findElements(By.css('.action-button button'))
  } else {
    return await webdriver.findElements(By.css('section[title="' + dashboardTitle + '"] .action-button button'))
  }
}

export async function getExecutionDialogOutput() {
    await webdriver.wait(new Condition('Dialog with long int is visible', async () => { 
      const dialog = await webdriver.findElement({ id: 'execution-results-popup' })
      return await dialog.isDisplayed()
    }));
    
    const ret = await webdriver.executeScript('return window.logEntries.get(window.executionDialog.executionTrackingId).output')

    return ret
}

export async function closeExecutionDialog() {
    const btnClose = await webdriver.findElements(By.css('[title="Close"]'))
    await btnClose[0].click()
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

  title = title.replaceAll('config: ', '')
	title = title.replaceAll(/[\(\)\|\*\<\>\:]/g, "_")
	title = title + '.failed-test'

    fs.writeFileSync('screenshots/' + title + '.png', img, 'base64')
  })
}

export async function getRootAndWait() {
  await webdriver.get(runner.baseUrl())
  await webdriver.wait(new Condition('wait for loaded-dashboard', async function() {
    const body = await webdriver.findElement(By.tagName('body'))
    const attr = await body.getAttribute('loaded-dashboard')

    console.log('loaded-dashboard: ', attr)

    if (attr) {
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
  const navigationLinks = await webdriver.findElements(By.css('.navigation-links li'))

  return navigationLinks
}

export async function requireExecutionDialogStatus (webdriver, expected) {
  await webdriver.wait(new Condition('wait for action to be running', async function () {
    const dialogStatus = await webdriver.findElement(By.id('execution-dialog-status'))
    const actual = await dialogStatus.getText()

    if (actual === expected) {
      return true
    } else {
      console.log('Waiting for domStatus text to be: ', expected, ', it is currently: ', actual)
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
