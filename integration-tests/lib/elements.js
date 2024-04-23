import { By } from 'selenium-webdriver'
import fs from 'fs'
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
