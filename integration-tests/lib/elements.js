import { By } from 'selenium-webdriver'
import fs from 'fs'
import { expect } from 'chai'
import { Condition } from 'selenium-webdriver'

export const DEFAULT_UI_WAIT_MS = 3000
const FAST_POLL_MS = 50

// Keep Selenium helpers in lockstep with the frontend DOM id helpers.
export {
  ARGUMENT_FIELD_ID_PREFIX,
  argumentFieldChoicesId,
  argumentFieldId,
  argumentFieldValidationElementId,
  argumentFieldValueId
} from '../../frontend/resources/vue/utils/argumentFieldIds.js'

const executionDialogStatusBy = By.css('.execution-dialog-status')
const sidebarIds = ['mainnav', 'picocrank-sidebar']

const sidebarNavLinksSelector = [
  sidebarCss('menu.navigation-links > li:not(.nav-section)'),
  sidebarCss('menu.nav-section-links > li')
].join(', ')

let loadedPageGeneration = null

function sidebarCss (suffix) {
  return sidebarIds.map((id) => `#${id} ${suffix}`).join(', ')
}

function isSamePagePath (currentUrl, targetUrl) {
  try {
    return new URL(currentUrl).pathname === new URL(targetUrl).pathname
  } catch {
    return false
  }
}

async function waitUntil (description, fn, timeoutMs = DEFAULT_UI_WAIT_MS) {
  await webdriver.wait(
    new Condition(description, fn),
    timeoutMs,
    undefined,
    FAST_POLL_MS
  )
}

async function getBodyAttribute (name) {
  return webdriver.executeScript(
    (attributeName) => document.body.getAttribute(attributeName),
    name
  )
}

async function isSidebarVisibleInBrowser () {
  return webdriver.executeScript((ids) => {
    for (const id of ids) {
      const sidebar = document.getElementById(id)
      if (!sidebar) {
        continue
      }

      const classes = sidebar.className
      if (classes.includes('shown') || classes.includes('stuck')) {
        return true
      }
    }

    return false
  }, sidebarIds)
}

async function countNavigationLinksInBrowser () {
  return webdriver.executeScript(
    (selector) => document.querySelectorAll(selector).length,
    sidebarNavLinksSelector
  )
}

export async function getActionButtons () {
  // Currently, only the active dashboard's contents are rendered,
  // so we don't need to scope the selector by dashboard title.
  return await webdriver.findElements(By.css('.action-button button'))
}

export async function getExecutionDialogOutput() {
    await waitUntil('Dialog with long int is visible', async () => {
      const dialog = await webdriver.findElement({ id: 'execution-results-popup' })
      return await dialog.isDisplayed()
    })

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

export async function waitForDashboardLoaded(timeoutMs = DEFAULT_UI_WAIT_MS, expectedTitle = null) {
  await waitUntil('wait for loaded-dashboard', async function () {
    const attr = await getBodyAttribute('loaded-dashboard')

    if (attr == null || attr === '') {
      return false
    }

    if (expectedTitle != null) {
      return attr === expectedTitle
    }

    return true
  }, timeoutMs)
}

export async function waitForLogsPage(timeoutMs = DEFAULT_UI_WAIT_MS) {
  await waitUntil('wait for logs page', async () => {
    const url = await webdriver.getCurrentUrl()
    return url.includes('/logs/') && !url.endsWith('/logs')
  }, timeoutMs)
}

export async function waitForArgumentFormPage(timeoutMs = DEFAULT_UI_WAIT_MS) {
  await waitUntil('wait for argument form page', async () => {
    const url = await webdriver.getCurrentUrl()
    return url.includes('/actionBinding/') && url.includes('/argumentForm')
  }, timeoutMs)
}

export async function waitForArgumentFormReady(timeoutMs = DEFAULT_UI_WAIT_MS) {
  await waitUntil('wait for argument form ready', async () => {
    const attr = await getBodyAttribute('loaded-argument-form')
    return attr != null && attr !== ''
  }, timeoutMs)
}

export async function waitForExecutionComplete(timeoutMs = DEFAULT_UI_WAIT_MS) {
  await waitUntil('wait for execution to finish', async () => {
    const statusElements = await webdriver.findElements(executionDialogStatusBy)
    if (statusElements.length === 0) {
      return false
    }

    try {
      const statusText = await statusElements[0].getText()
      return !statusText.includes('Still running') && !statusText.includes('Queued')
    } catch {
      return false
    }
  }, timeoutMs)
}

export async function getRootAndWait() {
  const targetUrl = runner.baseUrl()
  const sameConfig = loadedPageGeneration === runner.pageGeneration

  if (sameConfig && isSamePagePath(await webdriver.getCurrentUrl(), targetUrl)) {
    const attr = await getBodyAttribute('loaded-dashboard')
    if (attr != null && attr !== '') {
      return
    }
  }

  await webdriver.get(targetUrl)
  await waitForDashboardLoaded()
  loadedPageGeneration = runner.pageGeneration
}

async function isSidebarVisible () {
  return isSidebarVisibleInBrowser()
}

export async function closeSidebar() {
  if (await isSidebarVisible()) {
    await webdriver.findElement(By.id('sidebar-toggler-button')).click()
  }

  await waitUntil('wait for sidebar to close', async () => {
    return !(await isSidebarVisible())
  })
}

export async function openSidebar() {
  if (await isSidebarVisible()) {
    return
  }

  await webdriver.findElement(By.id('sidebar-toggler-button')).click()

  await waitUntil('wait for sidebar to open', async () => {
    return await isSidebarVisible()
  })
}

export async function getNavigationLinks() {
  return await webdriver.findElements(By.css(sidebarNavLinksSelector))
}

export async function getNavigationLinkTitles () {
  return webdriver.executeScript((selector) => {
    return [...document.querySelectorAll(selector)].map((linkElement) => {
      const title = linkElement.getAttribute('title')
      if (title) {
        return title
      }

      const anchor = linkElement.querySelector('a[href]')
      return anchor ? anchor.textContent.trim() : ''
    })
  }, sidebarNavLinksSelector)
}

export async function waitForNavigationLinks (minimumCount = 1, timeoutMs = DEFAULT_UI_WAIT_MS) {
  await waitUntil(`wait for at least ${minimumCount} navigation links`, async () => {
    const count = await countNavigationLinksInBrowser()
    return count >= minimumCount
  }, timeoutMs)
}

export async function getNavigationLinkTitle (linkElement) {
  const title = await linkElement.getAttribute('title')
  if (title) {
    return title
  }

  const anchor = await linkElement.findElement(By.css('a[href]'))
  return await anchor.getText()
}

export async function findSidebarNavHref (href) {
  return await webdriver.findElements(By.css(sidebarCss(`a[href="${href}"]`)))
}

export async function requireExecutionDialogStatus (webdriver, expected) {
  await waitUntil('wait for action to be running', async function () {
    const dialogStatus = await webdriver.findElement(executionDialogStatusBy)
    const actual = await dialogStatus.getText()
    return actual === expected
  })
}

export async function findExecutionDialog (webdriver) {
  return webdriver.findElement(By.id('execution-results-popup'))
}

export async function getActionButton (webdriver, title) {
  const buttons = await webdriver.findElements(By.css('[title="' + title + '"]'))

  expect(buttons).to.have.length(1)

  return buttons[0]
}

export async function getTerminalBuffer() {
  try {
    const output = await webdriver.executeScript(`
      if (window.terminal && window.terminal.getBufferAsString) {
        return window.terminal.getBufferAsString();
      }
      return null;
    `)
    return output
  } catch (e) {
    console.log('[getTerminalBuffer] Error:', e.message)
    return null
  }
}

export async function waitForCurrentUrl (predicate, timeoutMs = DEFAULT_UI_WAIT_MS) {
  await waitUntil('wait for url', async () => {
    const url = await webdriver.getCurrentUrl()
    return predicate(url)
  }, timeoutMs)
}

export async function waitForSelectorCount (selector, minimumCount = 1, timeoutMs = DEFAULT_UI_WAIT_MS) {
  await waitUntil(`wait for at least ${minimumCount} ${selector}`, async () => {
    const count = await webdriver.executeScript(
      (cssSelector) => document.querySelectorAll(cssSelector).length,
      selector
    )
    return count >= minimumCount
  }, timeoutMs)
}
