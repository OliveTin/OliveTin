import { By } from 'selenium-webdriver'
import fs from 'fs'

export async function getActionButtons (webdriver) {
  return await webdriver.findElement(By.id('contentActions')).findElements(By.tagName('button'))
}

export function takeScreenshot (webdriver) {
  return webdriver.takeScreenshot().then((img) => {
    fs.writeFileSync('out.png', img, 'base64')
  })
}
