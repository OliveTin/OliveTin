import { By } from 'selenium-webdriver'

export async function getActionButtons (webdriver) {
  return await webdriver.findElement(By.id('contentActions')).findElements(By.tagName('button'))
}
