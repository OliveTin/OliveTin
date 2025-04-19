import {
  getRootAndWait,
  takeScreenshotOnFailure,
} from '../lib/elements.js'

import { By } from 'selenium-webdriver'
import { expect } from 'chai'

describe('config: policy-all-false', function () {
  before(async function () {
    await runner.start('policy-all-false')
  });

  after(async () => {
    await runner.stop()
  });

  afterEach(function () {
    takeScreenshotOnFailure(this.currentTest, webdriver);
  });


  it('navbar should not contain default policy links', async function () {
    await getRootAndWait()

    const logListLink = await webdriver.findElements(By.css('[href="/logs"]'))
    expect(logListLink).to.be.empty

    const diagnosticsLink = await webdriver.findElements(By.css('[href="/diagnostics"]'))
    expect(diagnosticsLink).to.be.empty
  })
})
