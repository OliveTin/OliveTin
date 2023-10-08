import {By, Builder, Browser} from 'selenium-webdriver';
import getRunner from './runner.mjs';

export async function mochaGlobalSetup() {
  global.webdriver = await new Builder().forBrowser('chrome').build();
  global.runner = getRunner()
}

export async function mochaGlobalTeardown() {
  await global.webdriver.quit();
}
